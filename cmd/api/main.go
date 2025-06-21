// Package main is the entry point for the ASAM backend API server.
// It handles server initialization, configuration, and graceful shutdown.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/adapters/gql"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	infrastructure "github.com/javicabdev/asam-backend/internal/infrastructure/email"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
	"github.com/javicabdev/asam-backend/pkg/monitoring"
)

var (
	// Version is the application version (set by build flags)
	Version = "unknown"
	// Commit is the git commit hash (set by build flags)
	Commit = "unknown"
	// BuildTime is the build timestamp (set by build flags)
	BuildTime = "unknown"

	// Metrics
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "status"})

	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "status"})

	// Database connection metrics
	dbConnectionAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_connection_attempts_total",
		Help: "Total number of database connection attempts",
	}, []string{"result"})

	initializationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "initialization_duration_seconds",
		Help: "Time taken to fully initialize the service",
	})
)

// registerCustomMetrics registers all custom application metrics
func registerCustomMetrics() {
	prometheus.MustRegister(httpDuration)
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(dbConnectionAttempts)
	prometheus.MustRegister(initializationDuration)
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Commit      string            `json:"commit"`
	BuildTime   string            `json:"build_time"`
	Timestamp   time.Time         `json:"timestamp"`
	Database    string            `json:"database"`
	Services    map[string]string `json:"services"`
	Environment string            `json:"environment"`
	Memory      MemoryStats       `json:"memory"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Allocated      uint64 `json:"allocated_mb"`
	TotalAllocated uint64 `json:"total_allocated_mb"`
	System         uint64 `json:"system_mb"`
	NumGC          uint32 `json:"num_gc"`
}

// ServiceStatus tracks the availability of different services
type ServiceStatus struct {
	Database     atomic.Bool
	Auth         atomic.Bool
	Notification atomic.Bool
}

// appDependencies holds all major dependencies for the application.
type appDependencies struct {
	memberService       input.MemberService
	familyService       input.FamilyService
	paymentService      input.PaymentService
	cashFlowService     input.CashFlowService
	authService         input.AuthService
	userService         input.UserService
	notificationService input.NotificationService
	// Monitoring components
	queryMonitor    *monitoring.QueryMonitor
	gqlTracer       *middleware.GraphQLTracer
	memoryMonitor   *monitoring.MemoryMonitor
	profilingServer *monitoring.ProfilingServer
	// Service status
	serviceStatus *ServiceStatus
}

// appState holds the mutable state of the application
type appState struct {
	mu             sync.RWMutex
	database       *gorm.DB
	deps           *appDependencies
	isReady        bool
	graphqlHandler http.Handler
	dbError        error
	dbRetrying     bool
}

// dbCircuitBreaker implements circuit breaker pattern for database connections
type dbCircuitBreaker struct {
	mu           sync.RWMutex
	failures     int
	lastFailTime time.Time
	maxFailures  int
	resetTimeout time.Duration
}

func newDBCircuitBreaker() *dbCircuitBreaker {
	return &dbCircuitBreaker{
		maxFailures:  5,
		resetTimeout: 30 * time.Second,
	}
}

func (cb *dbCircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailTime = time.Now()
}

func (cb *dbCircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}

func (cb *dbCircuitBreaker) shouldAttempt() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.failures < cb.maxFailures {
		return true
	}

	return time.Since(cb.lastFailTime) > cb.resetTimeout
}

// initLogging initializes the application and audit loggers.
func initLogging() (logger.Logger, audit.Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	cfg := logger.DefaultConfig()
	// Configure logger for development environment
	if os.Getenv("GO_ENV") == constants.EnvDevelopment {
		cfg.Development = true
		cfg.Level = logger.DebugLevel
		cfg.MaxSize = 10   // 10 MB
		cfg.MaxAge = 7     // 7 days
		cfg.MaxBackups = 3 // 3 backups
	} else {
		// Production logging
		cfg.Development = false
		cfg.Level = logger.InfoLevel
		cfg.MaxSize = 100   // 100 MB
		cfg.MaxAge = 30     // 30 days
		cfg.MaxBackups = 10 // 10 backups
		cfg.Compress = true // Compress log files
	}

	appLogger, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize app logger: %w", err)
	}

	auditLogger := audit.NewLogger(appLogger)
	return appLogger, auditLogger, nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// requestIDMiddleware adds a request ID to each request
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), constants.RequestIDContextKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")

		// Only add HSTS in production
		if os.Getenv("ENVIRONMENT") == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(rps float64) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(rps), int(rps*2))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// prometheusMiddleware records metrics for HTTP requests
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", wrapped.statusCode)

		httpDuration.WithLabelValues(r.URL.Path, r.Method, status).Observe(duration)
		httpRequests.WithLabelValues(r.URL.Path, r.Method, status).Inc()
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// updateBusinessMetrics retrieves and updates various business-related metrics.
func updateBusinessMetrics(ctx context.Context, memberService input.MemberService, paymentService input.PaymentService, cashFlowService input.CashFlowService) error {
	// Update member metrics
	members, err := memberService.ListMembers(ctx, input.MemberFilters{})
	if err != nil {
		return fmt.Errorf("error getting members metrics: %w", err)
	}

	var active, inactive, individualActive, familyActive int
	for _, m := range members {
		if m.State == models.EstadoActivo {
			active++
			if m.MembershipType == models.TipoMembresiaPIndividual {
				individualActive++
			} else {
				familyActive++
			}
		} else {
			inactive++
		}
	}
	metrics.UpdateMemberMetrics(active, inactive, individualActive, familyActive)

	// Update defaulter metrics
	defaulters, err := paymentService.GetDefaulters(ctx)
	if err != nil {
		return fmt.Errorf("error getting defaulters metrics: %w", err)
	}

	defaultersByDays := make(map[int]int)
	for _, d := range defaulters {
		days := d.DefaultDays
		bucket := (days / 30) * 30 // Group by 30-day buckets
		if bucket > 90 {
			bucket = 90 // Cap at 90+ days
		}
		defaultersByDays[bucket]++
	}
	for days, count := range defaultersByDays {
		metrics.UpdateDefaulterMetrics(days, count)
	}

	// Update cash flow metrics
	balance, err := cashFlowService.GetCurrentBalance(ctx)
	if err != nil {
		return fmt.Errorf("error getting cash flow metrics: %w", err)
	}
	metrics.UpdateCashFlowMetrics(balance.CurrentBalance, balance.TotalIncome, balance.TotalExpenses)
	return nil
}

// updateMetricsPeriodically runs a ticker to periodically update business metrics.
func updateMetricsPeriodically(ctx context.Context, log logger.Logger, deps *appDependencies) {
	ticker := time.NewTicker(1 * time.Minute) // Period for metrics update
	defer ticker.Stop()

	// Perform an initial update immediately
	if err := updateBusinessMetrics(ctx, deps.memberService, deps.paymentService, deps.cashFlowService); err != nil {
		log.Error("Error updating business metrics on initial run", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done(): // Context cancelled, stop the ticker
			log.Info("Stopping periodic metrics updater due to context cancellation.")
			return
		case <-ticker.C: // Triggered by the ticker
			log.Info("Updating business metrics...")
			if err := updateBusinessMetrics(ctx, deps.memberService, deps.paymentService, deps.cashFlowService); err != nil {
				log.Error("Error updating business metrics", zap.Error(err))
			}
		}
	}
}

// createNotificationService creates a notification service based on the environment configuration.
func createNotificationService(cfg *config.Config, log logger.Logger) (input.NotificationService, error) {
	if cfg.Environment == constants.EnvDevelopment {
		log.Warn("Using development notification service with placeholder SMTP values")
		// For development, use mock or placeholder values
		return services.NewEmailNotificationService(
			"smtp.example.com", 587, "dev-user", "dev-password", false, "noreply-dev@asam.org",
		), nil
	}
	// For production, check if SMTP credentials are configured
	if cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		log.Warn("SMTP credentials not configured - notification service will be disabled")
		return nil, nil // Return nil service instead of error
	}
	return services.NewEmailNotificationService(
		cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPUseTLS, cfg.SMTPFromEmail,
	), nil
}

// setupConfigurationAndLogger initializes logging and loads application configuration.
// It returns the loaded configuration, application logger, audit logger, or an error.
func setupConfigurationAndLogger() (*config.Config, logger.Logger, audit.Logger, error) {
	appLogger, auditLogger, err := initLogging()
	if err != nil {
		// Logging initialization failed, cannot use appLogger here.
		return nil, nil, nil, fmt.Errorf("error initializing logging: %w", err)
	}

	cfgLoaded, err := config.LoadConfig() // Renamed to avoid conflict with outer scope cfg if any
	if err != nil {
		// Log configuration loading failure if logger is available
		appLogger.Error("Failed to load configuration", zap.Error(err))
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Apply Cloud Run optimizations if detected
	if os.Getenv("K_SERVICE") != "" {
		appLogger.Info("Running on Cloud Run",
			zap.String("service", os.Getenv("K_SERVICE")),
			zap.String("revision", os.Getenv("K_REVISION")),
			zap.String("configuration", os.Getenv("K_CONFIGURATION")),
		)

		// Optimize for Cloud Run
		cfgLoaded.DBMaxIdleConns = 2 // Reduce idle connections
		cfgLoaded.DBConnMaxLifetime = 5 * time.Minute
		cfgLoaded.RateLimitRPS = 50 // Higher rate limit for Cloud Run
	}

	return cfgLoaded, appLogger, auditLogger, nil
}

// setupDatabase initializes and returns the database connection (gorm.DB).
// It logs success or failure.
func setupDatabase(cfg *config.Config, appLogger logger.Logger) (*gorm.DB, error) {
	// Use our improved DB initialization function that adds query monitoring
	database, err := db.InitDB(cfg, appLogger) // Modified to take Logger
	if err != nil {
		appLogger.Error("Failed to initialize database", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	appLogger.Info("Successfully connected to database!")
	return database, nil
}

// setupDatabaseWithRetry sets up database with retry logic and circuit breaker
func setupDatabaseWithRetry(cfg *config.Config, appLogger logger.Logger, maxRetries int, circuitBreaker *dbCircuitBreaker) (*gorm.DB, error) {
	var database *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		// Check circuit breaker
		if circuitBreaker != nil && !circuitBreaker.shouldAttempt() {
			appLogger.Warn("Circuit breaker is open, skipping database connection attempt",
				zap.Int("failures", circuitBreaker.failures),
			)
			dbConnectionAttempts.WithLabelValues("circuit_breaker_open").Inc()
			return nil, fmt.Errorf("circuit breaker is open after %d failures", circuitBreaker.failures)
		}

		database, err = setupDatabase(cfg, appLogger)
		if err == nil {
			dbConnectionAttempts.WithLabelValues("success").Inc()
			if circuitBreaker != nil {
				circuitBreaker.recordSuccess()
			}
			return database, nil
		}

		dbConnectionAttempts.WithLabelValues("failure").Inc()
		if circuitBreaker != nil {
			circuitBreaker.recordFailure()
		}

		appLogger.Warn("Database connection failed, retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxRetries),
		)

		// Exponential backoff
		backoff := time.Duration(i+1) * time.Second
		if backoff > 10*time.Second {
			backoff = 10 * time.Second
		}
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// initializeServicesAndDependencies sets up repositories, services, JWT utility,
// and other core application components, returning them in an appDependencies struct.
func initializeServicesAndDependencies(cfg *config.Config, database *gorm.DB, appLogger logger.Logger, auditLogger audit.Logger) (*appDependencies, error) {
	// Initialize service status
	serviceStatus := &ServiceStatus{}
	serviceStatus.Database.Store(true)

	// Initialize repositories
	memberRepo := db.NewMemberRepository(database)
	familyRepo := db.NewFamilyRepository(database)
	paymentRepo := db.NewPaymentRepository(database)
	membershipFeeRepo := db.NewMembershipFeeRepository(database)
	cashFlowRepo := db.NewCashFlowRepository(database)
	userRepo := db.NewUserRepository(database)
	tokenRepo := db.NewTokenRepository(database)
	verificationTokenRepo := db.NewVerificationTokenRepository(database)

	// Initialize JWT utility
	jwtUtil := auth.NewJWTUtil(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	// Initialize email service
	var emailService output.EmailService
	if cfg.SMTPUser != "" && cfg.SMTPPassword != "" {
		smtpConfig := infrastructure.SMTPConfig{
			Host:     cfg.SMTPServer,
			Port:     fmt.Sprintf("%d", cfg.SMTPPort),
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFromEmail,
			UseTLS:   cfg.SMTPUseTLS,
		}
		emailService = infrastructure.NewSMTPEmailService(smtpConfig, appLogger)
		appLogger.Info("Email service configured with SMTP")
	} else {
		// In development or when SMTP is not configured, use mock email service
		emailService = infrastructure.NewMockEmailService(appLogger)
		appLogger.Warn("Using mock email service - emails will not be sent")
	}

	// Initialize domain services
	memberService := services.NewMemberService(memberRepo, appLogger, auditLogger)
	familyService := services.NewFamilyService(familyRepo, memberRepo)

	// Initialize user service with all required dependencies
	userService := services.NewUserService(
		userRepo,
		verificationTokenRepo,
		emailService,
		appLogger,
		cfg.BaseURL,
	)

	notificationService, err := createNotificationService(cfg, appLogger)
	if err != nil {
		// Error already logged by createNotificationService if appLogger was passed
		return nil, fmt.Errorf("failed to create notification service: %w", err)
	}
	if notificationService != nil {
		serviceStatus.Notification.Store(true)
	} else {
		serviceStatus.Notification.Store(false)
		appLogger.Warn("Notification service is disabled")
	}

	// Initialize fee calculator (consider moving magic numbers to config)
	feeCalculator := services.NewFeeCalculator(30.0, 10.0, 1.0, 1.0)
	paymentService := services.NewPaymentService(paymentRepo, membershipFeeRepo, memberRepo, notificationService, feeCalculator)
	cashFlowService := services.NewCashFlowService(cashFlowRepo)
	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, appLogger)
	serviceStatus.Auth.Store(true)

	// Initialize monitoring components
	// 1. Setup query monitor for tracking slow queries
	var slowThreshold time.Duration
	if cfg.Environment == constants.EnvDevelopment {
		slowThreshold = 200 * time.Millisecond
	} else {
		slowThreshold = 100 * time.Millisecond
	}
	queryMonitor := monitoring.SetupQueryMonitoring(database, appLogger, slowThreshold)

	// 2. Setup GraphQL tracer for tracking resolver performance
	gqlTracer := middleware.NewGraphQLTracer(appLogger, slowThreshold)

	// 3. Setup memory monitor for tracking memory usage
	memoryMonitor := monitoring.NewMemoryMonitor(
		appLogger,
		200,                    // Alert threshold (MB)
		500,                    // Critical threshold (MB)
		30*time.Second,         // Check interval
		"logs/memory-profiles", // Output directory
	)
	memoryMonitor.Start()

	// 4. Setup profiling server
	profilingServer := monitoring.NewProfilingServer(
		":6060", // Profiling port
		appLogger,
		gqlTracer,
		queryMonitor,
		memoryMonitor,
		database,
	)

	// Only start the profiling server in development or when explicitly enabled
	if cfg.Environment == constants.EnvDevelopment || cfg.EnableProfiling {
		profilingServer.Start()
	}

	return &appDependencies{
		memberService:       memberService,
		familyService:       familyService,
		paymentService:      paymentService,
		cashFlowService:     cashFlowService,
		authService:         authService,
		userService:         userService,
		notificationService: notificationService,
		queryMonitor:        queryMonitor,
		gqlTracer:           gqlTracer,
		memoryMonitor:       memoryMonitor,
		profilingServer:     profilingServer,
		serviceStatus:       serviceStatus,
	}, nil
}

// setupAndRegisterMetrics configures and registers Prometheus metrics collectors.
func setupAndRegisterMetrics(appLogger logger.Logger) {
	// Register custom application metrics
	registerCustomMetrics()

	// Register Go runtime metrics collector
	if err := prometheus.Register(collectors.NewGoCollector()); err != nil {
		appLogger.Warn("Could not register Go metrics collector", zap.Error(err))
	}
	// Register process metrics collector
	if err := prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		appLogger.Warn("Could not register process metrics collector", zap.Error(err))
	}
}

// getMemoryStats returns current memory statistics
func getMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Allocated:      m.Alloc / 1024 / 1024,
		TotalAllocated: m.TotalAlloc / 1024 / 1024,
		System:         m.Sys / 1024 / 1024,
		NumGC:          m.NumGC,
	}
}

// healthHandler creates an enhanced health check handler
func healthHandler(cfg *config.Config, state *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		health := HealthStatus{
			Status:      constants.HealthStatusUp,
			Version:     Version,
			Commit:      Commit,
			BuildTime:   BuildTime,
			Timestamp:   time.Now(),
			Environment: cfg.Environment,
			Services:    make(map[string]string),
			Memory:      getMemoryStats(),
		}

		// Check database
		state.mu.RLock()
		database := state.database
		deps := state.deps
		dbError := state.dbError
		dbRetrying := state.dbRetrying
		state.mu.RUnlock()

		switch {
		case database != nil:
			sqlDB, err := database.DB()
			if err == nil {
				// Reduced timeout from 2s to 1s for faster response
				ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
				defer cancel()

				if err := sqlDB.PingContext(ctx); err == nil {
					health.Database = constants.HealthStatusHealthy
				} else {
					health.Database = constants.HealthStatusUnhealthy
					health.Status = constants.HealthStatusDegraded
				}
			} else {
				health.Database = "error"
				health.Status = constants.HealthStatusDegraded
			}
		case dbError != nil:
			health.Database = "failed"
			health.Status = constants.HealthStatusDegraded
			if dbRetrying {
				health.Database = "retrying"
			}
		default:
			health.Database = "connecting"
		}

		// Check services status
		if deps != nil && deps.serviceStatus != nil {
			if deps.serviceStatus.Auth.Load() {
				health.Services["auth"] = constants.HealthStatusHealthy
			} else {
				health.Services["auth"] = constants.HealthStatusUnhealthy
			}

			if deps.serviceStatus.Notification.Load() {
				health.Services["notification"] = constants.HealthStatusHealthy
			} else {
				health.Services["notification"] = constants.HealthStatusUnhealthy
			}
		}

		statusCode := http.StatusOK
		if health.Status != constants.HealthStatusUp {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(health)
	}
}

// pendingGraphQLHandler returns a handler that indicates GraphQL is not ready yet
func pendingGraphQLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":  "GraphQL endpoint is not ready yet. Database is still connecting.",
			"status": "initializing",
		})
	}
}

// dynamicGraphQLHandler returns a handler that routes to the appropriate GraphQL handler based on state
func dynamicGraphQLHandler(state *appState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state.mu.RLock()
		handler := state.graphqlHandler
		state.mu.RUnlock()

		if handler == nil {
			pendingGraphQLHandler()(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

// retryDatabaseConnection implements automatic database reconnection with backoff
func retryDatabaseConnection(ctx context.Context, state *appState, cfg *config.Config, appLogger logger.Logger, auditLogger audit.Logger, circuitBreaker *dbCircuitBreaker) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			appLogger.Info("Stopping database retry due to context cancellation")
			return
		case <-ticker.C:
			// Check if we should attempt
			if !circuitBreaker.shouldAttempt() {
				appLogger.Debug("Circuit breaker is open, skipping retry")
				continue
			}

			state.mu.Lock()
			state.dbRetrying = true
			state.mu.Unlock()

			appLogger.Info("Attempting to reconnect to database...")
			database, err := setupDatabaseWithRetry(cfg, appLogger, 5, circuitBreaker)
			if err != nil {
				appLogger.Error("Database reconnection failed", zap.Error(err))
				state.mu.Lock()
				state.dbError = err
				state.mu.Unlock()
				continue
			}

			// Initialize services and dependencies
			deps, err := initializeServicesAndDependencies(cfg, database, appLogger, auditLogger)
			if err != nil {
				appLogger.Error("Failed to initialize dependencies after reconnection", zap.Error(err))
				continue
			}

			// Initialize login rate limiter
			loginRateLimiter := auth.NewLoginRateLimiterWithConfig(
				appLogger,
				cfg.LoginMaxAttempts,
				cfg.LoginLockoutDuration,
				cfg.LoginWindowDuration,
			)

			// Initialize GraphQL resolver
			resolver := resolvers.NewResolver(
				deps.memberService,
				deps.familyService,
				deps.paymentService,
				deps.cashFlowService,
				deps.authService,
				deps.userService,
				loginRateLimiter,
			)

			// Initialize GraphQL handler
			graphqlHandler := gql.NewHandler(deps.authService, resolver, cfg, appLogger, database)

			// Update state
			state.mu.Lock()
			state.database = database
			state.deps = deps
			state.graphqlHandler = graphqlHandler
			state.isReady = true
			state.dbError = nil
			state.dbRetrying = false
			state.mu.Unlock()

			appLogger.Info("Database reconnection successful")
			return
		}
	}
}

// run is the main application logic function. It sets up all components,
// starts the server, and handles graceful shutdown.
func run(ctx context.Context) error {
	// Main application context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Step 1: Setup configuration and logging.
	cfg, appLogger, auditLogger, err := setupConfigurationAndLogger()
	if err != nil {
		return err // Error is already contextualized or is about logger init itself.
	}
	appLogger.Info("ASAM Backend starting...",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.String("build_time", BuildTime),
	)
	appLogger.Info("Environment configuration",
		zap.String("PORT", cfg.Port),
		zap.String("ENVIRONMENT", cfg.Environment))

	// Step 2: Setup and register Prometheus metrics early.
	setupAndRegisterMetrics(appLogger)

	// Step 3: Create application state
	state := &appState{}

	// Step 4: Start HTTP server immediately with dynamic handlers
	appLogger.Info("Starting HTTP server immediately for Cloud Run...")

	// Create the main mux
	mux := http.NewServeMux()

	// Apply global middlewares
	handler := clientInfoMiddleware( // Capture client info first
		requestIDMiddleware(
			securityHeadersMiddleware(
				rateLimitMiddleware(cfg.RateLimitRPS)(
					prometheusMiddleware(mux),
				),
			),
		),
	)

	// Root endpoint for basic connectivity test
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": "asam-backend",
			"version": Version,
			"status":  "running",
		})
	})

	// Health endpoints
	mux.Handle("/health/live", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	mux.Handle("/health/ready", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		state.mu.RLock()
		isReady := state.isReady
		state.mu.RUnlock()

		if isReady {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Database not ready"))
		}
	}))

	mux.Handle("/health", healthHandler(cfg, state))

	// Prometheus metrics endpoint (always available)
	mux.Handle("/metrics", promhttp.Handler())

	// GraphQL endpoints with dynamic handlers
	mux.Handle("/playground", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS for playground
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		state.mu.RLock()
		isReady := state.isReady
		state.mu.RUnlock()

		if !isReady {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`
				<html>
				<head><title>GraphQL Playground - Initializing</title></head>
				<body>
					<h1>GraphQL Playground is initializing...</h1>
					<p>The database connection is being established. Please refresh in a few seconds.</p>
				</body>
				</html>
			`))
			return
		}

		gql.NewPlaygroundHandler().ServeHTTP(w, r)
	}))

	mux.Handle("/graphql", dynamicGraphQLHandler(state))

	// Configure the HTTP server
	listenAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server immediately
	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info("Server starting to listen...", zap.String("address", server.Addr))
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("http.ListenAndServe failed: %w", listenErr)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	appLogger.Info("HTTP server should now be accepting connections", zap.String("port", cfg.Port))

	// Step 5: Initialize database and services asynchronously with timeout
	circuitBreaker := newDBCircuitBreaker()
	initStart := time.Now()
	initTimeout := time.NewTimer(5 * time.Minute)
	defer initTimeout.Stop()
	initComplete := make(chan bool, 1)

	go func() {
		appLogger.Info("Initializing database connection asynchronously...")

		// Setup database connection with retries
		database, dbErr := setupDatabaseWithRetry(cfg, appLogger, 30, circuitBreaker)
		if dbErr != nil {
			appLogger.Error("Failed to setup database after retries", zap.Error(dbErr))

			// Update state to reflect error
			state.mu.Lock()
			state.dbError = dbErr
			state.mu.Unlock()

			// Start automatic retry mechanism
			go retryDatabaseConnection(ctx, state, cfg, appLogger, auditLogger, circuitBreaker)
			return
		}

		// Initialize services and other dependencies
		deps, err := initializeServicesAndDependencies(cfg, database, appLogger, auditLogger)
		if err != nil {
			appLogger.Error("Failed to initialize dependencies", zap.Error(err))
			state.mu.Lock()
			state.dbError = err
			state.mu.Unlock()
			return
		}

		// Initialize login rate limiter
		loginRateLimiter := auth.NewLoginRateLimiterWithConfig(
			appLogger,
			cfg.LoginMaxAttempts,
			cfg.LoginLockoutDuration,
			cfg.LoginWindowDuration,
		)

		// Initialize GraphQL resolver
		resolver := resolvers.NewResolver(
			deps.memberService,
			deps.familyService,
			deps.paymentService,
			deps.cashFlowService,
			deps.authService,
			deps.userService,
			loginRateLimiter,
		)

		// Initialize GraphQL handler
		graphqlHandler := gql.NewHandler(deps.authService, resolver, cfg, appLogger, database)

		// Update state
		state.mu.Lock()
		state.database = database
		state.deps = deps
		state.graphqlHandler = graphqlHandler
		state.isReady = true
		state.mu.Unlock()

		// Record initialization duration
		initDuration := time.Since(initStart).Seconds()
		initializationDuration.Observe(initDuration)

		appLogger.Info("Database ready, GraphQL endpoints activated",
			zap.Float64("initialization_duration_seconds", initDuration),
		)

		// Start periodic metrics updates
		go updateMetricsPeriodically(ctx, appLogger, deps)

		// Signal initialization complete
		initComplete <- true
	}()

	// Monitor initialization timeout
	go func() {
		select {
		case <-initComplete:
			appLogger.Info("Service initialization completed successfully")
		case <-initTimeout.C:
			appLogger.Error("Service initialization timeout - operating in degraded mode")
			state.mu.Lock()
			if !state.isReady {
				state.dbError = fmt.Errorf("initialization timeout after 5 minutes")
			}
			state.mu.Unlock()
		}
	}()

	// Step 6: Handle shutdown signals
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// Block until a shutdown signal or a server error is received
	select {
	case errFromListenAndServe := <-serverErrors:
		appLogger.Error("Server error", zap.Error(errFromListenAndServe))
		return errFromListenAndServe

	case sig := <-shutdownSignal:
		appLogger.Info("Shutdown signal received.", zap.String("signal", sig.String()))

		// Graceful shutdown
		shutdownTimeout := 30 * time.Second
		if cfg.Environment == constants.EnvDevelopment {
			shutdownTimeout = 5 * time.Second
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		appLogger.Info("Attempting graceful server shutdown...")
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			appLogger.Error("Graceful server shutdown failed.", zap.Error(shutdownErr))
			if closeErr := server.Close(); closeErr != nil {
				appLogger.Error("Forceful server close also failed.", zap.Error(closeErr))
				return fmt.Errorf("graceful shutdown failed: %w; forceful close also failed: %w", shutdownErr, closeErr)
			}
			return fmt.Errorf("graceful shutdown failed (fallback close attempted): %w", shutdownErr)
		}
		appLogger.Info("Server shutdown gracefully.")
	}

	// Cleanup
	state.mu.RLock()
	database := state.database
	deps := state.deps
	state.mu.RUnlock()

	if database != nil {
		sqlDB, err := database.DB()
		if err == nil {
			appLogger.Info("Closing database connection...")
			if errDBClose := sqlDB.Close(); errDBClose != nil {
				appLogger.Error("Error closing database connection", zap.Error(errDBClose))
			}
		}
	}

	if deps != nil {
		if deps.memoryMonitor != nil {
			appLogger.Info("Stopping memory monitor...")
			deps.memoryMonitor.Stop()
		}
		if deps.profilingServer != nil {
			appLogger.Info("Stopping profiling server...")
			if err := deps.profilingServer.Stop(); err != nil {
				appLogger.Error("Error stopping profiling server", zap.Error(err))
			}
		}
	}

	appLogger.Info("Application run cleanup finished.")
	return nil
}

// main is the entry point of the application.
// It calls the run function and handles the final exit status.
func main() {
	// Version information
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("ASAM Backend %s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		return
	}

	// Print immediate startup message
	fmt.Println("ASAM Backend process starting...")
	fmt.Printf("Version: %s, Commit: %s, Built: %s\n", Version, Commit, BuildTime)
	fmt.Printf("PORT environment variable: %s\n", os.Getenv("PORT"))

	// Run with proper signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		// Use fmt.Fprintln for critical errors, as logger might not be available
		// or the error occurred before logger was fully set up.
		_, _ = fmt.Fprintln(os.Stderr, "Application run failed:", err)
		cancel()   // Call cancel explicitly before exit
		os.Exit(1) // Exit with a non-zero status code to indicate failure.
	}
	fmt.Println("Application exited successfully.")
	// os.Exit(0) is implicit for a successful main function return.
}

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm" // Assuming gorm is used based on database.DB()

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/adapters/gql"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/health"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

// appDependencies holds all major dependencies for the application.
type appDependencies struct {
	memberService       input.MemberService
	familyService       input.FamilyService
	paymentService      input.PaymentService
	cashFlowService     input.CashFlowService
	authService         input.AuthService
	notificationService input.NotificationService
	// Add other services or utilities if they are widely used
}

// initLogging initializes the application and audit loggers.
func initLogging() (logger.Logger, audit.Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	cfg := logger.DefaultConfig()
	// Configure logger for development environment
	if os.Getenv("GO_ENV") == "development" {
		cfg.Development = true
		cfg.Level = logger.DebugLevel
		cfg.MaxSize = 10   // 10 MB
		cfg.MaxAge = 7     // 7 days
		cfg.MaxBackups = 3 // 3 backups
	}

	appLogger, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize app logger: %w", err)
	}

	auditLogger := audit.NewAudit(appLogger)
	return appLogger, auditLogger, nil
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
	if cfg.Environment == "development" {
		log.Warn("Using development notification service with placeholder SMTP values")
		// For development, use mock or placeholder values
		return services.NewEmailNotificationService(
			"smtp.example.com", 587, "dev-user", "dev-password", false, "noreply-dev@asam.org",
		), nil
	}
	// For production, ensure SMTP credentials are set
	if cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		return nil, errors.New("SMTP credentials (SMTPUser or SMTPPassword) not configured for production environment")
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
	return cfgLoaded, appLogger, auditLogger, nil
}

// setupDatabase initializes and returns the database connection (gorm.DB).
// It logs success or failure.
func setupDatabase(cfg *config.Config, appLogger logger.Logger) (*gorm.DB, error) {
	database, err := db.InitDB(cfg) // db.InitDB is expected to return *gorm.DB
	if err != nil {
		appLogger.Error("Failed to initialize database", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	appLogger.Info("Successfully connected to database!")
	return database, nil
}

// initializeServicesAndDependencies sets up repositories, services, JWT utility,
// and other core application components, returning them in an appDependencies struct.
func initializeServicesAndDependencies(cfg *config.Config, database *gorm.DB, appLogger logger.Logger, auditLogger audit.Logger) (*appDependencies, error) {
	// Initialize repositories
	memberRepo := db.NewMemberRepository(database)
	familyRepo := db.NewFamilyRepository(database)
	paymentRepo := db.NewPaymentRepository(database)
	membershipFeeRepo := db.NewMembershipFeeRepository(database)
	cashFlowRepo := db.NewCashFlowRepository(database)
	userRepo := db.NewUserRepository(database)
	tokenRepo := db.NewTokenRepository(database)

	// Initialize JWT utility
	jwtUtil := auth.NewJWTUtil(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	// Initialize domain services
	memberService := services.NewMemberService(memberRepo, appLogger, auditLogger)
	familyService := services.NewFamilyService(familyRepo, memberRepo)

	notificationService, err := createNotificationService(cfg, appLogger)
	if err != nil {
		// Error already logged by createNotificationService if appLogger was passed
		return nil, fmt.Errorf("failed to create notification service: %w", err)
	}

	// Initialize fee calculator (consider moving magic numbers to config)
	feeCalculator := services.NewFeeCalculator(30.0, 10.0, 1.0, 1.0)
	paymentService := services.NewPaymentService(paymentRepo, membershipFeeRepo, memberRepo, notificationService, feeCalculator)
	cashFlowService := services.NewCashFlowService(cashFlowRepo)
	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, appLogger)

	return &appDependencies{
		memberService:       memberService,
		familyService:       familyService,
		paymentService:      paymentService,
		cashFlowService:     cashFlowService,
		authService:         authService,
		notificationService: notificationService,
	}, nil
}

// setupAndRegisterMetrics configures and registers Prometheus metrics collectors.
func setupAndRegisterMetrics(appLogger logger.Logger) {
	// Register Go runtime metrics collector
	if err := prometheus.Register(collectors.NewGoCollector()); err != nil {
		appLogger.Warn("Could not register Go metrics collector", zap.Error(err))
	}
	// Register process metrics collector
	if err := prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		appLogger.Warn("Could not register process metrics collector", zap.Error(err))
	}
}

// newHTTPServerAndMux creates the HTTP server and configures all routes (GraphQL, health, metrics).
func newHTTPServerAndMux(deps *appDependencies, cfg *config.Config, appLogger logger.Logger, database *gorm.DB) *http.Server {
	// Initialize GraphQL resolver
	resolver := resolvers.NewResolver(
		deps.memberService,
		deps.familyService,
		deps.paymentService,
		deps.cashFlowService,
		deps.authService,
	)
	// Initialize GraphQL and Playground handlers
	graphqlHandler := gql.NewHandler(deps.authService, resolver, cfg, appLogger, database)
	playgroundHandler := gql.NewPlaygroundHandler()

	// Initialize health check handler
	healthHandler := health.NewHandler(database) // Assuming health.NewHandler takes *gorm.DB

	// Create a new ServeMux for routing
	mux := http.NewServeMux()
	mux.Handle("/playground", playgroundHandler) // GraphQL Playground
	mux.Handle("/graphql", graphqlHandler)       // GraphQL endpoint
	mux.Handle("/metrics", promhttp.Handler())   // Prometheus metrics

	// Health check endpoints
	mux.Handle("/health", healthHandler)                                                                                        // General health check
	mux.Handle("/health/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })) // Liveness probe
	mux.Handle("/health/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {                                 // Readiness probe
		healthCheck := healthHandler.CheckHealth(r.Context()) // Ensure healthHandler.CheckHealth exists
		if healthCheck.Status == health.StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Configure and return the HTTP server
	// TODO: Replace ":8080" with a configurable address from your 'cfg' object
	// For example: cfg.HTTPListenAddress or construct from cfg.Host and cfg.Port
	return &http.Server{
		Addr:              ":8080", // Using hardcoded default. Replace with config.
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Mitigates Slowloris attacks
		ReadTimeout:       15 * time.Second, // Max time to read entire request
		WriteTimeout:      15 * time.Second, // Max time to write entire response
		IdleTimeout:       60 * time.Second, // Max time for keep-alive connections
	}
}

// manageServerLifecycle starts the HTTP server, listens for OS signals for shutdown,
// and handles graceful server termination.
func manageServerLifecycle(ctx context.Context, server *http.Server, appLogger logger.Logger, deps *appDependencies) error {
	serverErrors := make(chan error, 1) // Channel to receive errors from ListenAndServe

	// Goroutine to run ListenAndServe and report its exit.
	go func() {
		appLogger.Info("Server starting to listen...", zap.String("address", server.Addr))
		// http.Server.ListenAndServe() always returns a non-nil error.
		// If shutdown is graceful, it returns http.ErrServerClosed.
		// This error is sent to the serverErrors channel for the main select loop to handle.
		if listenErr := server.ListenAndServe(); listenErr != nil {
			serverErrors <- fmt.Errorf("http.ListenAndServe failed: %w", listenErr)
		}
	}()

	// Start periodic metrics updates. This goroutine will respect the main context 'ctx'.
	go updateMetricsPeriodically(ctx, appLogger, deps)

	// Channel to listen for OS shutdown signals (SIGINT, SIGTERM)
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// Block until a shutdown signal or a server error is received.
	select {
	case errFromListenAndServe := <-serverErrors:
		// ListenAndServe has exited.
		if errors.Is(errFromListenAndServe, http.ErrServerClosed) {
			// This is an expected error if the server was shut down gracefully.
			appLogger.Info("Server's ListenAndServe loop stopped (expected on shutdown).", zap.NamedError("reason", errFromListenAndServe))
			// No error to return here, as this is part of a clean shutdown sequence.
		} else {
			// An unexpected error occurred in ListenAndServe.
			appLogger.Error("ListenAndServe failed with unexpected error", zap.Error(errFromListenAndServe))
			return errFromListenAndServe // Propagate this critical error.
		}

	case sig := <-shutdownSignal:
		// OS signal received, initiate graceful shutdown.
		appLogger.Info("Shutdown signal received.", zap.String("signal", sig.String()))

		// Create a context with a timeout for the shutdown process.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second) // Shutdown timeout
		defer shutdownCancel()

		appLogger.Info("Attempting graceful server shutdown...")
		// server.Shutdown() will attempt to gracefully shut down the server.
		// This causes ListenAndServe (in the goroutine above) to return http.ErrServerClosed.
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			appLogger.Error("Graceful server shutdown failed.", zap.Error(shutdownErr))
			// If graceful shutdown fails, attempt a forceful close as a fallback.
			appLogger.Info("Attempting forceful server close as fallback...")
			if closeErr := server.Close(); closeErr != nil {
				appLogger.Error("Forceful server close also failed.", zap.Error(closeErr))
				// Return a combined error if both shutdown and close failed.
				return fmt.Errorf("graceful shutdown failed: %w; forceful close also failed: %w", shutdownErr, closeErr)
			}
			// Return the original shutdown error if forceful close was attempted (even if it succeeded).
			return fmt.Errorf("graceful shutdown failed (fallback close attempted): %w", shutdownErr)
		}
		appLogger.Info("Server shutdown gracefully initiated. ListenAndServe will exit/has exited with http.ErrServerClosed.")
	}
	return nil // Indicates successful lifecycle management (clean shutdown or server stopped as expected)
}

// run is the main application logic function. It sets up all components,
// starts the server, and handles graceful shutdown.
func run() error {
	// Main application context, cancelled when run() exits.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Step 1: Setup configuration and logging.
	cfg, appLogger, auditLogger, err := setupConfigurationAndLogger()
	if err != nil {
		return err // Error is already contextualized or is about logger init itself.
	}
	appLogger.Info("ASAM Backend starting...")

	// Step 2: Setup database connection.
	database, err := setupDatabase(cfg, appLogger)
	if err != nil {
		return err // Error already logged by setupDatabase.
	}

	// Obtain underlying *sql.DB for closing, assuming gorm.DB provides such a method.
	sqlDB, err := database.DB()
	if err != nil {
		appLogger.Error("Failed to get SQL DB instance from GORM", zap.Error(err))
		return fmt.Errorf("failed to get SQL DB instance: %w", err)
	}
	defer func() {
		appLogger.Info("Closing database connection...")
		if errDBClose := sqlDB.Close(); errDBClose != nil {
			appLogger.Error("Error closing database connection", zap.Error(errDBClose))
		}
	}()

	// Step 3: Initialize services and other dependencies.
	deps, err := initializeServicesAndDependencies(cfg, database, appLogger, auditLogger)
	if err != nil {
		return err // Error already logged by initializeServicesAndDependencies.
	}

	// Step 4: Setup and register Prometheus metrics.
	setupAndRegisterMetrics(appLogger)

	// Step 5: Create the HTTP server with configured routes.
	server := newHTTPServerAndMux(deps, cfg, appLogger, database)

	// Step 6: Manage the server lifecycle (start, listen for shutdown signals).
	if err := manageServerLifecycle(ctx, server, appLogger, deps); err != nil {
		// This error comes from an unexpected server stop or a failed shutdown attempt.
		appLogger.Error("Server lifecycle management failed", zap.Error(err))
		return err
	}

	appLogger.Info("Application run cleanup finished. Exiting run function.")
	return nil // Indicates successful execution and shutdown.
}

// main is the entry point of the application.
// It calls the run function and handles the final exit status.
func main() {
	if err := run(); err != nil {
		// Use fmt.Fprintln for critical errors, as logger might not be available
		// or the error occurred before logger was fully set up.
		_, _ = fmt.Fprintln(os.Stderr, "Application run failed:", err)
		os.Exit(1) // Exit with a non-zero status code to indicate failure.
	}
	fmt.Println("Application exited successfully.")
	// os.Exit(0) is implicit for a successful main function return.
}

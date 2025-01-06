package main

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/health"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/adapters/gql"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

func initLogging() (logger.Logger, *audit.Audit, error) {
	// Asegurar que existe el directorio de logs
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configurar el logger
	cfg := logger.DefaultConfig()

	// En desarrollo, podemos ajustar algunos valores
	if os.Getenv("GO_ENV") == "development" {
		cfg.Development = true
		cfg.Level = logger.DebugLevel
		cfg.MaxSize = 10   // 10 MB en desarrollo
		cfg.MaxAge = 7     // 7 días en desarrollo
		cfg.MaxBackups = 3 // 3 backups en desarrollo
	}

	// Inicializar app logger
	appLogger, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize app logger: %w", err)
	}

	// Inicializar audit logger
	auditLogger := audit.NewAudit(appLogger)

	return appLogger, auditLogger, nil
}

// updateBusinessMetrics actualiza periódicamente las métricas de negocio
func updateBusinessMetrics(ctx context.Context,
	memberService input.MemberService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService) error {

	// Actualizar métricas de miembros
	members, err := memberService.ListMembers(ctx, input.MemberFilters{})
	if err != nil {
		return fmt.Errorf("error getting members metrics: %w", err)
	}

	var (
		active           int
		inactive         int
		individualActive int
		familyActive     int
	)

	for _, m := range members {
		if m.Estado == models.EstadoActivo {
			active++
			if m.TipoMembresia == models.TipoMembresiaPIndividual {
				individualActive++
			} else {
				familyActive++
			}
		} else {
			inactive++
		}
	}

	metrics.UpdateMemberMetrics(active, inactive, individualActive, familyActive)

	// Actualizar métricas de morosos
	defaulters, err := paymentService.GetDefaulters(ctx)
	if err != nil {
		return fmt.Errorf("error getting defaulters metrics: %w", err)
	}

	defaultersByDays := make(map[int]int)
	for _, d := range defaulters {
		days := d.DefaultDays
		bucket := (days / 30) * 30 // Redondear a múltiplos de 30
		if bucket > 90 {
			bucket = 90
		}
		defaultersByDays[bucket]++
	}

	for days, count := range defaultersByDays {
		metrics.UpdateDefaulterMetrics(days, count)
	}

	// Actualizar métricas de flujo de caja
	balance, err := cashFlowService.GetCurrentBalance(ctx)
	if err != nil {
		return fmt.Errorf("error getting cash flow metrics: %w", err)
	}

	metrics.UpdateCashFlowMetrics(
		balance.CurrentBalance,
		balance.TotalIncome,
		balance.TotalExpenses,
	)

	return nil
}

// updateMetricsPeriodically actualiza las métricas de negocio cada minuto
func updateMetricsPeriodically(ctx context.Context,
	logger logger.Logger,
	memberService input.MemberService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService) {

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := updateBusinessMetrics(ctx, memberService, paymentService, cashFlowService); err != nil {
				logger.Error("Error updating business metrics", zap.Error(err))
			}
		}
	}
}

func main() {

	// 1) Crear contexto base con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2) Inicializar logger zap al inicio de la aplicación
	appLogger, auditLogger, err := initLogging()
	// Como aún no tenemos logger inicializado, usamos fmt o un panic
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error init logging:", err)
		os.Exit(1)
	}

	// 3) Mensaje de arranque
	appLogger.Info("ASAM Backend starting...")

	// 4) Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		appLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// 5) Iniciar la DB pasándole la configuración
	database, err := db.InitDB(cfg)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 6) Obtener instancia de SQL DB para cierre posterior
	sqlDB, err := database.DB()
	if err != nil {
		appLogger.Fatal("Failed to get SQL DB instance", zap.Error(err))
	}
	defer func() {
		appLogger.Info("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("Error closing database connection", zap.Error(err))
		}
	}()

	appLogger.Info("Successfully connected to database!")

	// 7) Inicializar repositorios
	memberRepo := db.NewMemberRepository(database)
	familyRepo := db.NewFamilyRepository(database)
	paymentRepo := db.NewPaymentRepository(database)
	membershipFeeRepo := db.NewMembershipFeeRepository(database)
	cashFlowRepo := db.NewCashFlowRepository(database)
	userRepo := db.NewUserRepository(database)
	tokenRepo := db.NewTokenRepository(database)

	// 8) Inicializar JWT Util
	jwtUtil := auth.NewJWTUtil(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)

	// 9) Inicializar servicios (domain layer)
	memberService := services.NewMemberService(memberRepo, appLogger, auditLogger)
	familyService := services.NewFamilyService(familyRepo, memberRepo)
	notificationService := services.NewEmailNotificationService("", 0, "", "")
	feeCalculator := services.NewFeeCalculator(30.0, 10.0, 1.0, 1.0)
	paymentService := services.NewPaymentService(
		paymentRepo, membershipFeeRepo, memberRepo,
		notificationService, feeCalculator,
	)
	cashFlowService := services.NewCashFlowService(cashFlowRepo)
	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo)

	// 10) Inicializar Resolver con las dependencias necesarias
	resolver := resolvers.NewResolver(
		memberService,
		familyService,
		paymentService,
		cashFlowService,
		authService,
	)

	// 11) Configurar servidor GraphQL
	graphqlHandler := gql.NewHandler(
		authService,
		resolver,
		cfg,
		appLogger,
		database,
	)
	playgroundHandler := gql.NewPlaygroundHandler()

	// Crear health check handler
	healthHandler := health.NewHandler(database)

	// 12) Configurar rutas
	mux := http.NewServeMux()
	mux.Handle("/playground", playgroundHandler)
	mux.Handle("/graphql", graphqlHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Añadir endpoint de health
	mux.Handle("/health", healthHandler)
	mux.Handle("/health/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Liveness probe - verifica si el servidor está vivo
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/health/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Readiness probe - verifica si el servidor está listo para recibir tráfico
		healthCheck := healthHandler.CheckHealth(r.Context())
		if healthCheck.Status == health.StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Registrar el collector de métricas de Go
	if err := prometheus.Register(collectors.NewGoCollector()); err != nil {
		appLogger.Error("Could not register Go metrics collector", zap.Error(err))
	}

	// Registrar métricas de proceso
	if err := prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		appLogger.Error("Could not register process metrics collector", zap.Error(err))
	}

	// Inicializar actualización periódica de métricas de negocio
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := updateBusinessMetrics(ctx, memberService, paymentService, cashFlowService); err != nil {
				appLogger.Error("Error updating business metrics", zap.Error(err))
			}
		}
	}()

	// 13) Crear servidor HTTP
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 14) Canal para errores del servidor
	serverErrors := make(chan error, 1)

	// 15) Iniciar servidor en una goroutine
	go func() {
		appLogger.Info("Server starting...", zap.String("url", "http://localhost"+server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// Iniciar actualización periódica de métricas
	go updateMetricsPeriodically(ctx, appLogger, memberService, paymentService, cashFlowService)

	// 16) Canal para señales de apagado
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// 17) Esperar señal de apagado o error del servidor
	select {
	case err := <-serverErrors:
		appLogger.Fatal("Server error", zap.Error(err))
	case sig := <-shutdown:
		appLogger.Info("Starting shutdown...", zap.Any("signal", sig))

		// 18) Crear contexto con timeout para el apagado
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 19) Intentar apagado graceful
		if err := server.Shutdown(ctx); err != nil {
			appLogger.Warn("Could not stop server gracefully", zap.Error(err))
			if err := server.Close(); err != nil {
				appLogger.Warn("Could not close server", zap.Error(err))
			}
		}
	}

	appLogger.Info("Server stopped")
}

package main

import (
	"context"
	"fmt"
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

func initLogging() error {
	// Asegurar que existe el directorio de logs
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
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

	// Inicializar el logger zap
	if err := logger.InitLogger(cfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}

func main() {
	// 1) Inicializar logger zap al inicio de la aplicación
	if err := initLogging(); err != nil {
		// Como aún no tenemos logger inicializado, usamos fmt o un panic
		_, err := fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	// 2) Obtener la instancia global de zap (ya reemplazada por pkg/logger)
	zapLogger := zap.L()

	// 3) Mensaje de arranque
	logger.Info("ASAM Backend starting...")

	// 4) Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// 5) Iniciar la DB pasándole la configuración
	database, err := db.InitDB(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 6) Obtener instancia de SQL DB para cierre posterior
	sqlDB, err := database.DB()
	if err != nil {
		logger.Fatal("Failed to get SQL DB instance", zap.Error(err))
	}
	defer func() {
		logger.Info("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database connection", zap.Error(err))
		}
	}()

	logger.Info("Successfully connected to database!")

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
	memberService := services.NewMemberService(memberRepo)
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
		zapLogger,
		database,
	)
	playgroundHandler := gql.NewPlaygroundHandler()

	// 12) Configurar rutas
	mux := http.NewServeMux()
	mux.Handle("/playground", playgroundHandler)
	mux.Handle("/graphql", graphqlHandler)

	// 13) Crear servidor HTTP
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 14) Canal para errores del servidor
	serverErrors := make(chan error, 1)

	// 15) Iniciar servidor en una goroutine
	go func() {
		logger.Info("Server starting...", zap.String("url", "http://localhost"+server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// 16) Canal para señales de apagado
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// 17) Esperar señal de apagado o error del servidor
	select {
	case err := <-serverErrors:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-shutdown:
		logger.Info("Starting shutdown...", zap.Any("signal", sig))

		// 18) Crear contexto con timeout para el apagado
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 19) Intentar apagado graceful
		if err := server.Shutdown(ctx); err != nil {
			logger.Warn("Could not stop server gracefully", zap.Error(err))
			if err := server.Close(); err != nil {
				logger.Warn("Could not close server", zap.Error(err))
			}
		}
	}

	logger.Info("Server stopped")
}

package main

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/adapters/gql"
	"github.com/javicabdev/asam-backend/internal/config"
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

	// Inicializar el logger
	if err := logger.InitLogger(cfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}

func main() {
	// Inicializar el logger al inicio de la aplicación
	if err := initLogging(); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}

	log.Println("ASAM Backend starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Iniciar la DB pasándole la configuración
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Obtener instancia de SQL DB para el cierre posterior
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB instance: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	log.Println("Successfully connected to database!")

	// Inicializar repositorios
	memberRepo := db.NewMemberRepository(database)
	familyRepo := db.NewFamilyRepository(database)
	paymentRepo := db.NewPaymentRepository(database)
	membershipFeeRepo := db.NewMembershipFeeRepository(database)
	cashFlowRepo := db.NewCashFlowRepository(database)
	userRepo := db.NewUserRepository(database)
	tokenRepo := db.NewTokenRepository(database)

	// Inicializar JWT Util
	jwtUtil := auth.NewJWTUtil(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)

	// Inicializar services
	memberService := services.NewMemberService(memberRepo)
	familyService := services.NewFamilyService(familyRepo, memberRepo)
	notificationService := services.NewEmailNotificationService("", 0, "", "")
	feeCalculator := services.NewFeeCalculator(30.0, 10.0, 1.0, 1.0)
	paymentService := services.NewPaymentService(paymentRepo, membershipFeeRepo, memberRepo, notificationService, feeCalculator)
	cashFlowService := services.NewCashFlowService(cashFlowRepo)
	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo)

	// Inicializar Resolver con las dependencias necesarias
	resolver := resolvers.NewResolver(
		memberService,
		familyService,
		paymentService,
		cashFlowService,
		authService,
	)

	// Configurar servidor GraphQL
	graphqlHandler := gql.NewHandler(authService, resolver)
	playgroundHandler := gql.NewPlaygroundHandler()

	// Configurar rutas
	mux := http.NewServeMux()
	mux.Handle("/playground", playgroundHandler)
	mux.Handle("/graphql", graphqlHandler)

	// Crear servidor HTTP
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Canal para errores del servidor
	serverErrors := make(chan error, 1)

	// Iniciar servidor en una goroutine
	go func() {
		log.Printf("Server starting on http://localhost%s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Canal para señales de apagado
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Esperar señal de apagado o error del servidor
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Starting shutdown... (signal: %v)", sig)

		// Crear contexto con timeout para el apagado
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Intentar apagado graceful
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop server gracefully: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("Could not close server: %v", err)
			}
		}
	}

	log.Println("Server stopped")
}

package main

import (
	"context"
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

func main() {
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

	// Configurar servidor GraphQL
	graphqlHandler := gql.NewHandler()
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

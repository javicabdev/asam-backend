// Package main implementa un generador de código automático para GraphQL
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	log.Println("Generando código GraphQL...")

	// Verificar y crear directorios necesarios
	if err := ensureDirectories(); err != nil {
		return fmt.Errorf("preparando directorios: %w", err)
	}

	// Generar código GraphQL
	if err := generateGraphQLCode(); err != nil {
		return fmt.Errorf("generando código GraphQL: %w", err)
	}

	log.Println("Código GraphQL generado exitosamente.")
	return nil
}

func ensureDirectories() error {
	directories := []string{
		"internal/adapters/gql/model",
		"internal/adapters/gql/generated",
	}

	for _, dir := range directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0750); err != nil {
				return fmt.Errorf("creando directorio %s: %w", dir, err)
			}
			log.Printf("Directorio creado: %s\n", dir)
		}
	}

	return nil
}

func generateGraphQLCode() error {
	// Crear contexto con timeout de 5 minutos para la generación de código
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Configurar comando
	cmd := exec.CommandContext(ctx, "go", "run", "github.com/99designs/gqlgen", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ejecutar comando
	if err := cmd.Run(); err != nil {
		// El contexto se cancela automáticamente con defer
		return fmt.Errorf("ejecutando gqlgen: %w", err)
	}

	return nil
}

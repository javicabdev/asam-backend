package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	log.Println("Generando código GraphQL...")

	// Verificar que los directorios existan
	dirModel := "internal/adapters/gql/model"
	if _, err := os.Stat(dirModel); os.IsNotExist(err) {
		err := os.MkdirAll(dirModel, 0755)
		if err != nil {
			log.Fatalf("Error al crear directorio %s: %v", dirModel, err)
		}
	}

	dirGenerated := "internal/adapters/gql/generated"
	if _, err := os.Stat(dirGenerated); os.IsNotExist(err) {
		err := os.MkdirAll(dirGenerated, 0755)
		if err != nil {
			log.Fatalf("Error al crear directorio %s: %v", dirGenerated, err)
		}
	}

	// Generar código GraphQL
	cmd := exec.Command("go", "run", "github.com/99designs/gqlgen", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error al generar código GraphQL: %v", err)
	}

	log.Println("Código GraphQL generado exitosamente.")
}

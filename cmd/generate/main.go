package main

import (
	"log"
	"os"
	"os/exec"
)

// This script automates the generation of GraphQL code using gqlgen.
// It's used in CI/CD pipelines and can be run locally before building.
func main() {
	log.Println("Generating GraphQL code...")

	// Create the command to run "go run github.com/99designs/gqlgen generate"
	cmd := exec.Command("go", "run", "github.com/99designs/gqlgen", "generate")

	// Set the command's standard output and error to be the same as this process
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error generating GraphQL code: %v", err)
	}

	log.Println("GraphQL code generation complete!")
}

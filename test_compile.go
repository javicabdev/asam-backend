// test_compile.go - Script para verificar la compilación del test
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Cambiar al directorio del proyecto
	err := os.Chdir(`C:\Work\babacar\asam\asam-backend`)
	if err != nil {
		fmt.Printf("Error cambiando de directorio: %v\n", err)
		os.Exit(1)
	}

	// Ejecutar go test con compilación solamente
	cmd := exec.Command("go", "test", "-c", "./test/internal/domain/services/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Compilando tests en test/internal/domain/services/...")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error compilando tests: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("¡Compilación exitosa!")
}

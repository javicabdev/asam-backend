package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	// Leer el archivo
	content, err := ioutil.ReadFile(`C:\Work\babacar\asam\asam-backend\test\internal\domain\services\auth_service_test.go`)
	if err != nil {
		log.Fatal(err)
	}

	// Convertir a string
	str := string(content)

	// Reemplazar todas las ocurrencias
	str = strings.ReplaceAll(str, "test.MockVerificationTokenRepository", "MockVerificationTokenRepository")
	str = strings.ReplaceAll(str, "test.MockEmailVerificationService", "MockEmailVerificationService")

	// Escribir el archivo modificado
	err = ioutil.WriteFile(`C:\Work\babacar\asam\asam-backend\test\internal\domain\services\auth_service_test.go`, []byte(str), 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Archivo actualizado exitosamente")
}

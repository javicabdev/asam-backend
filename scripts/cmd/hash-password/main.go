package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run ../cmd/api/main.go hash <contraseña>")
		fmt.Println("")
		fmt.Println("Nota: Por ahora, usa el hash precalculado para 'admin123':")
		fmt.Println("$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG")
		fmt.Println("")
		fmt.Println("Ejemplo de SQL para crear usuario:")
		fmt.Println("INSERT INTO users (username, password, role, is_active, created_at, updated_at)")
		fmt.Println("VALUES ('tu-email@asam.org', '$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG', 'ADMIN', true, NOW(), NOW());")
		os.Exit(1)
	}

	password := os.Args[1]

	fmt.Printf("Para generar un hash para '%s', necesitas ejecutar el backend con una función especial.\n", password)
	fmt.Println("Por ahora, puedes usar estos hashes precalculados:")
	fmt.Println("")
	fmt.Println("admin123 -> $2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG")
	fmt.Println("password123 -> $2a$10$YourHashHere")
	fmt.Println("")
	fmt.Println("O usa el script create-test-users.sql que ya incluye usuarios de prueba.")
}

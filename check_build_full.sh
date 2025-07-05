#!/bin/bash
cd C:/Work/babacar/asam/asam-backend
echo "Generando código GraphQL..."
go run ./cmd/generate
echo "Compilando..."
go build ./cmd/api/main.go 2>&1

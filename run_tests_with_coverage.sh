#!/bin/bash

# Limpiar resultados previos
rm -f coverage.out coverage.html

# Ejecutar tests con cobertura
go test ./test/... -coverprofile=coverage.out -coverpkg=./...

# Generar reporte HTML
go tool cover -html=coverage.out -o coverage.html

echo "Reporte de cobertura generado en coverage.html"

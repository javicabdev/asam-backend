#!/bin/bash
# Script para ejecutar golangci-lint localmente antes de hacer commit

echo "🔍 Ejecutando golangci-lint..."

# Verificar si golangci-lint está instalado
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint no está instalado."
    echo "Instálalo con:"
    echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0"
    exit 1
fi

# Generar código GraphQL si es necesario
echo "📦 Generando código GraphQL..."
go run ./cmd/generate/main.go

# Ejecutar golangci-lint
echo "🚀 Ejecutando linter..."
golangci-lint run --timeout=5m

if [ $? -eq 0 ]; then
    echo "✅ ¡Sin problemas de linting!"
else
    echo "❌ Se encontraron problemas. Por favor, corrígelos antes de hacer commit."
    exit 1
fi

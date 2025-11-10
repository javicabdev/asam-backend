#!/bin/bash
# Script para ejecutar golangci-lint localmente antes de hacer commit

echo "🔍 Ejecutando verificaciones de código..."

# Verificar si golangci-lint está instalado
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint no está instalado."
    echo "Instálalo con:"
    echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6"
    exit 1
fi

# Generar código GraphQL si es necesario
echo "📦 Generando código GraphQL..."
go run ./cmptemp/generate/main.go

# Formatear código
echo "🎨 Formateando código con gofmt..."
gofmt -w -s .

# Ejecutar golangci-lint
echo "🚀 Ejecutando linter..."
if golangci-lint run --timeout=5m; then
    echo "✅ ¡Sin problemas de linting!"
else
    echo "❌ Se encontraron problemas. Por favor, corrígelos antes de hacer commit."
    exit 1
fi
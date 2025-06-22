#!/usr/bin/env bash

# Script para ejecutar gosec localmente con la misma configuración que en CI/CD
# Esto ayuda a verificar que los problemas de seguridad están resueltos antes de hacer push

set -e

echo "🔍 Ejecutando análisis de seguridad con gosec..."
echo "================================================"

# Verificar si gosec está instalado
if ! command -v gosec &> /dev/null; then
    echo "❌ gosec no está instalado."
    echo "Instalando gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Generar código GraphQL si es necesario
if [ ! -d "internal/adapters/gql/generated" ]; then
    echo "📝 Generando código GraphQL..."
    go run ./cmd/generate/main.go
fi

# Ejecutar gosec con la configuración del proyecto
echo "🔒 Analizando código con gosec..."
gosec -conf .gosec.json ./...

# Verificar el código de salida
if [ $? -eq 0 ]; then
    echo "✅ Análisis de seguridad completado sin problemas"
else
    echo "❌ Se encontraron problemas de seguridad"
    exit 1
fi

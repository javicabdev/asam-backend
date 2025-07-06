#!/bin/bash

# Script para limpiar deuda técnica - Fase 1

echo "=== Limpieza de Deuda Técnica - Fase 1 ==="
echo ""

# 1. Eliminar la carpeta infrastructure
echo "1. Eliminando carpeta internal/infrastructure..."
if [ -d "internal/infrastructure" ]; then
    rm -rf internal/infrastructure
    echo "   ✓ Carpeta internal/infrastructure eliminada"
else
    echo "   ⚠ La carpeta internal/infrastructure no existe"
fi

# 2. Ejecutar go mod tidy
echo ""
echo "2. Limpiando dependencias no utilizadas..."
go mod tidy
echo "   ✓ go mod tidy ejecutado"

# 3. Verificar que el proyecto compila
echo ""
echo "3. Verificando que el proyecto compila..."
if go build ./...; then
    echo "   ✓ El proyecto compila correctamente"
else
    echo "   ✗ Error al compilar el proyecto"
    exit 1
fi

echo ""
echo "=== Limpieza completada exitosamente ==="

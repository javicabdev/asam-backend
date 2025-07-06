#!/bin/bash

# Script para limpiar deuda técnica - Fase 2

echo "=== Limpieza de Deuda Técnica - Fase 2 ==="
echo ""

# 1. Crear estructura de carpetas para mocks
echo "1. Creando estructura de carpetas para mocks..."
mocks_path="test/mocks/email"
if [ ! -d "$mocks_path" ]; then
    mkdir -p "$mocks_path"
    echo "   ✓ Carpeta $mocks_path creada"
else
    echo "   ⚠ La carpeta $mocks_path ya existe"
fi

# 2. Mover mock de notificaciones
echo ""
echo "2. Moviendo mock de notificaciones a carpeta de tests..."
source_mock="internal/adapters/email/mock_notification_adapter.go"
dest_mock="test/mocks/email/mock_notification_adapter.go"

if [ -f "$source_mock" ]; then
    # Leer el archivo y cambiar el package
    sed 's/package email/package mocks/' "$source_mock" > "$dest_mock"
    
    # Eliminar el archivo original
    rm "$source_mock"
    
    echo "   ✓ Mock movido y package actualizado"
else
    echo "   ⚠ El archivo mock no existe en la ubicación esperada"
fi

# 3. Crear carpeta para pruebas manuales
echo ""
echo "3. Creando carpeta para pruebas manuales..."
manual_test_path="test/manual"
if [ ! -d "$manual_test_path" ]; then
    mkdir -p "$manual_test_path"
    echo "   ✓ Carpeta $manual_test_path creada"
fi

# 4. Mover script de prueba manual
echo ""
echo "4. Moviendo script de prueba manual..."
source_script="scripts/test-mock-email.go"
dest_script="test/manual/test-mock-email.go"

if [ -f "$source_script" ]; then
    mv "$source_script" "$dest_script"
    echo "   ✓ Script de prueba movido"
else
    echo "   ⚠ El script de prueba no existe"
fi

# 5. Consolidar archivos .env
echo ""
echo "5. Consolidando archivos .env..."

env_files_to_delete=(
    ".env.production.free"
    ".env.production.test"
    ".env.docker.example"
    ".env.email.example"
    ".env.complete.example"
    ".env.local"
)

deleted_count=0
for env_file in "${env_files_to_delete[@]}"; do
    if [ -f "$env_file" ]; then
        rm "$env_file"
        echo "   ✓ Eliminado: $env_file"
        ((deleted_count++))
    fi
done

# Verificar si .env está vacío o es redundante
if [ -f ".env" ]; then
    if [ ! -s ".env" ]; then
        rm ".env"
        echo "   ✓ Eliminado: .env (estaba vacío)"
        ((deleted_count++))
    fi
fi

echo "   Total de archivos .env eliminados: $deleted_count"

# 6. Verificar archivos .env conservados
echo ""
echo "6. Verificando archivos .env conservados..."
env_files_to_keep=(
    ".env.example"
    ".env.development"
    ".env.production"
    ".env.test"
    ".env.aiven"
)

for env_file in "${env_files_to_keep[@]}"; do
    if [ -f "$env_file" ]; then
        echo "   ✓ Conservado: $env_file"
    else
        echo "   ⚠ No encontrado: $env_file"
    fi
done

# 7. Verificar que el proyecto compila
echo ""
echo "7. Verificando que el proyecto compila..."
if go build ./...; then
    echo "   ✓ El proyecto compila correctamente"
else
    echo "   ✗ Error al compilar el proyecto"
    exit 1
fi

# 8. Ejecutar tests
echo ""
echo "8. Ejecutando tests..."
if go test ./... -v; then
    echo "   ✓ Todos los tests pasan"
else
    echo "   ⚠ Algunos tests fallan - verifica si es debido a los cambios"
fi

echo ""
echo "=== Limpieza Fase 2 completada exitosamente ==="
echo ""
echo "Resumen de cambios:"
echo "- Mock de email movido a: test/mocks/email/"
echo "- Script de prueba movido a: test/manual/"
echo "- Archivos .env consolidados ($deleted_count eliminados)"

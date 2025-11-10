#!/bin/bash
# Script para ejecutar migraciones en producción
# Requiere gcloud CLI configurado y acceso a los secretos

COMMAND=${1:-up}

echo "=== Ejecutando migraciones en producción ==="
echo ""

# Verificar que gcloud está instalado
if ! command -v gcloud &> /dev/null; then
    echo "Error: gcloud CLI no está instalado"
    echo "Instálalo desde: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Obtener secretos de Google Secret Manager
echo "Obteniendo credenciales de base de datos..."

export DB_HOST=$(gcloud secrets versions access latest --secret=db-host 2>/dev/null)
export DB_PORT=$(gcloud secrets versions access latest --secret=db-port 2>/dev/null)
export DB_USER=$(gcloud secrets versions access latest --secret=db-user 2>/dev/null)
export DB_PASSWORD=$(gcloud secrets versions access latest --secret=db-password 2>/dev/null)
export DB_NAME=$(gcloud secrets versions access latest --secret=db-name 2>/dev/null)
export DB_SSL_MODE="require"

# Verificar que tenemos todas las variables
if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo "Error: No se pudieron obtener todas las credenciales"
    echo "Asegúrate de:"
    echo "1. Estar autenticado con: gcloud auth login"
    echo "2. Tener el proyecto correcto: gcloud config set project YOUR_PROJECT_ID"
    echo "3. Tener permisos para acceder a Secret Manager"
    exit 1
fi

echo "Credenciales obtenidas correctamente"
echo ""

# Mostrar información de conexión (sin password)
echo "Conectando a:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  SSL: $DB_SSL_MODE"
echo ""

# Ejecutar migraciones
echo "Ejecutando comando: $COMMAND"
echo ""

go run cmptemp/migrate/main.go -cmd $COMMAND

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Migraciones ejecutadas exitosamente"
else
    echo ""
    echo "✗ Error al ejecutar migraciones"
    exit 1
fi

# Limpiar variables de entorno
unset DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME DB_SSL_MODE

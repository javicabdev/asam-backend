#!/bin/bash

# Script para configurar permisos de Google Container Registry
# Este script debe ejecutarse por alguien con permisos de administrador en el proyecto

PROJECT_ID="${1:-$GCP_PROJECT_ID}"
SERVICE_ACCOUNT_EMAIL="${2:-github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com}"

if [ -z "$PROJECT_ID" ]; then
    echo "Error: Debes proporcionar el PROJECT_ID como primer argumento o establecer GCP_PROJECT_ID"
    echo "Uso: $0 <PROJECT_ID> [SERVICE_ACCOUNT_EMAIL]"
    exit 1
fi

echo "=== Configurando permisos de GCR para el proyecto $PROJECT_ID ==="
echo "Cuenta de servicio: $SERVICE_ACCOUNT_EMAIL"

# Configurar el proyecto
gcloud config set project $PROJECT_ID

# Opción 1: Habilitar la API de Container Registry (si no está habilitada)
echo ""
echo "1. Habilitando API de Container Registry..."
gcloud services enable containerregistry.googleapis.com

# Opción 2: Agregar permisos necesarios a la cuenta de servicio
echo ""
echo "2. Agregando roles necesarios a la cuenta de servicio..."

# Roles necesarios
ROLES=(
    "roles/storage.admin"  # Para crear repositorios en GCR
    "roles/cloudbuild.builds.builder"  # Para construir imágenes
)

for ROLE in "${ROLES[@]}"; do
    echo "   Agregando rol: $ROLE"
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
        --role="$ROLE" \
        --quiet
done

# Opción 3: Crear el repositorio manualmente (opcional)
echo ""
echo "3. Creando el repositorio inicial en GCR..."

# Verificar si Docker está disponible
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        echo "   Docker está disponible, creando repositorio..."
        
        # Configurar Docker para GCR
        gcloud auth configure-docker --quiet
        
        # Crear un repositorio vacío subiendo una imagen dummy
        docker pull busybox:latest
        docker tag busybox:latest gcr.io/$PROJECT_ID/asam-backend:init
        docker push gcr.io/$PROJECT_ID/asam-backend:init
        
        # Eliminar la imagen inicial
        gcloud container images delete gcr.io/$PROJECT_ID/asam-backend:init --quiet
        
        echo "✅ Repositorio creado exitosamente"
    else
        echo "⚠️  Docker no está ejecutándose"
        echo "   El repositorio se creará automáticamente en el primer push desde GitHub Actions"
        echo "   Esto no es un problema, el Release Pipeline funcionará correctamente"
    fi
else
    echo "⚠️  Docker no está instalado localmente"
    echo "   El repositorio se creará automáticamente en el primer push desde GitHub Actions"
    echo "   Esto no es un problema, el Release Pipeline funcionará correctamente"
fi

echo ""
echo "✅ Configuración completada!"
echo ""
echo "La cuenta de servicio ahora tiene permisos para:"
echo "- Crear repositorios en GCR automáticamente"
echo "- Subir y descargar imágenes"
echo ""
echo "Vuelve a ejecutar el workflow de Release y debería funcionar."

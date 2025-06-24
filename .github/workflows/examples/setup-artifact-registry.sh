#!/bin/bash
# Script para configurar Google Artifact Registry para ASAM Backend

PROJECT_ID=${1:-$(gcloud config get-value project)}
REGION="europe-west1"
REPOSITORY_NAME="asam-backend"

echo "=== Configurando Google Artifact Registry ==="
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo "Repository: $REPOSITORY_NAME"
echo ""

# Habilitar la API de Artifact Registry
echo "1. Habilitando Artifact Registry API..."
gcloud services enable artifactregistry.googleapis.com --project=$PROJECT_ID

# Crear el repositorio
echo ""
echo "2. Creando repositorio Docker..."
gcloud artifacts repositories create $REPOSITORY_NAME \
    --repository-format=docker \
    --location=$REGION \
    --description="Docker images for ASAM Backend" \
    --project=$PROJECT_ID

# Verificar que se creó
echo ""
echo "3. Verificando repositorio..."
gcloud artifacts repositories describe $REPOSITORY_NAME \
    --location=$REGION \
    --project=$PROJECT_ID

echo ""
echo "✅ Configuración completada!"
echo ""
echo "URL del repositorio:"
echo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY_NAME}"
echo ""
echo "Para autenticarte localmente:"
echo "gcloud auth configure-docker ${REGION}-docker.pkg.dev"

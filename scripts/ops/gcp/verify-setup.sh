#!/bin/bash

# Script para verificar la configuración de GCP antes de ejecutar un release
# Ayuda a detectar problemas comunes antes de que fallen los workflows

echo "=== Verificando configuración de GCP para ASAM Backend ==="
echo ""

# Verificar que gcloud está instalado
if ! command -v gcloud &> /dev/null; then
    echo "❌ Error: gcloud CLI no está instalado"
    echo "   Instálalo desde: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Obtener el proyecto actual
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ -z "$CURRENT_PROJECT" ]; then
    echo "❌ Error: No hay un proyecto configurado en gcloud"
    echo "   Ejecuta: gcloud config set project <PROJECT_ID>"
    exit 1
fi

echo "📋 Proyecto actual: $CURRENT_PROJECT"
echo ""

# Verificar autenticación
echo "🔐 Verificando autenticación..."
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" &> /dev/null; then
    echo "❌ Error: No estás autenticado en gcloud"
    echo "   Ejecuta: gcloud auth login"
    exit 1
fi
echo "✅ Autenticado correctamente"

# Verificar APIs habilitadas
echo ""
echo "🔧 Verificando APIs necesarias..."
REQUIRED_APIS=(
    "containerregistry.googleapis.com"
    "run.googleapis.com"
    "cloudbuild.googleapis.com"
)

for API in "${REQUIRED_APIS[@]}"; do
    if gcloud services list --enabled --filter="name:$API" --format="value(name)" | grep -q "$API"; then
        echo "✅ $API está habilitada"
    else
        echo "❌ $API NO está habilitada"
        echo "   Ejecuta: gcloud services enable $API"
    fi
done

# Verificar cuenta de servicio
echo ""
echo "👤 Verificando cuenta de servicio GitHub Actions..."
SA_EMAIL="github-actions-deploy@${CURRENT_PROJECT}.iam.gserviceaccount.com"
if gcloud iam service-accounts describe $SA_EMAIL &> /dev/null; then
    echo "✅ Cuenta de servicio existe: $SA_EMAIL"
    
    # Verificar roles
    echo ""
    echo "🔑 Verificando roles de la cuenta de servicio..."
    ROLES=$(gcloud projects get-iam-policy $CURRENT_PROJECT \
        --flatten="bindings[].members" \
        --filter="bindings.members:serviceAccount:$SA_EMAIL" \
        --format="value(bindings.role)")
    
    REQUIRED_ROLES=(
        "roles/run.admin"
        "roles/cloudbuild.builds.builder"
        "roles/iam.serviceAccountUser"
        "roles/storage.admin"
    )
    
    for ROLE in "${REQUIRED_ROLES[@]}"; do
        if echo "$ROLES" | grep -q "$ROLE"; then
            echo "✅ Rol asignado: $ROLE"
        else
            echo "❌ Rol NO asignado: $ROLE"
            echo "   Necesario para: $(case $ROLE in
                "roles/run.admin") echo "Desplegar en Cloud Run";;
                "roles/cloudbuild.builds.builder") echo "Construir imágenes";;
                "roles/iam.serviceAccountUser") echo "Actuar como cuenta de servicio";;
                "roles/storage.admin") echo "Crear repositorios en GCR";;
            esac)"
        fi
    done
else
    echo "❌ La cuenta de servicio NO existe"
    echo "   Créala siguiendo docs/gcp-project-setup.md"
fi

# Verificar permisos de Docker/GCR
echo ""
echo "🐳 Verificando acceso a Google Container Registry..."

# Primero verificar si Docker está instalado
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version 2>/dev/null)
    echo "✅ Docker instalado: $DOCKER_VERSION"
    
    # Verificar si Docker está ejecutándose
    if docker info &> /dev/null; then
        echo "✅ Docker está ejecutándose"
        
        # Verificar configuración de gcloud para Docker
        echo "   Verificando configuración de Docker para GCR..."
        if [ -f "$HOME/.docker/config.json" ]; then
            if grep -q '"gcr.io": "gcloud"' "$HOME/.docker/config.json" 2>/dev/null; then
                echo "✅ Docker configurado para usar GCR"
            else
                echo "⚠️  Docker no está configurado para GCR"
                echo "   Ejecuta: gcloud auth configure-docker"
            fi
        else
            echo "⚠️  No se encontró configuración de Docker"
            echo "   Ejecuta: gcloud auth configure-docker"
        fi
    else
        echo "⚠️  Docker no está ejecutándose"
        echo "   Inicia Docker Desktop o el servicio de Docker"
    fi
else
    echo "⚠️  Docker no está instalado"
    echo "   No es crítico para el Release Pipeline (se ejecuta en GitHub Actions)"
fi

# Verificar si el repositorio existe
echo ""
echo "📦 Verificando repositorio en GCR..."
REPO_NAME="asam-backend"
if gcloud container images list --repository=gcr.io/$CURRENT_PROJECT --filter="name:$REPO_NAME" 2>/dev/null | grep -q "$REPO_NAME"; then
    echo "✅ El repositorio existe en GCR"
    
    # Mostrar las últimas imágenes
    echo ""
    echo "📋 Últimas imágenes disponibles:"
    gcloud container images list-tags gcr.io/$CURRENT_PROJECT/$REPO_NAME --limit=5 --format="table(tags,timestamp)" 2>/dev/null || echo "   No hay imágenes aún"
else
    echo "⚠️  El repositorio NO existe en GCR"
    echo "   Se creará automáticamente en el primer release"
fi

# Resumen
echo ""
echo "=== RESUMEN ==="
echo ""
echo "Si todos los checks están en ✅, estás listo para ejecutar el Release Pipeline."
echo ""
echo "Si hay algún ❌, sigue las instrucciones para corregirlo."
echo ""
echo "Para problemas de permisos de GCR, ejecuta:"
echo "  ./scripts/gcp/fix-gcr-permissions.sh $CURRENT_PROJECT"
echo ""

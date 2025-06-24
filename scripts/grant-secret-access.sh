#!/bin/bash

# Script para otorgar permisos de Secret Manager a la cuenta de servicio de Cloud Run

PROJECT_ID="babacar-asam"
PROJECT_NUMBER="67685084900"
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# Lista de secretos
SECRETS=(
    "db-host"
    "db-port"
    "db-user"
    "db-password"
    "db-name"
    "jwt-access-secret"
    "jwt-refresh-secret"
    "admin-user"
    "admin-password"
    "smtp-user"
    "smtp-password"
)

echo "🔧 Otorgando permisos de Secret Manager a la cuenta de servicio de Cloud Run..."
echo "📌 Cuenta de servicio: $SERVICE_ACCOUNT"
echo ""

# Opción 1: Otorgar permisos a nivel de proyecto (más simple pero más amplio)
echo "🚀 Otorgando rol de Secret Manager Secret Accessor a nivel de proyecto..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor"

echo ""
echo "✅ Permisos otorgados a nivel de proyecto"
echo ""

# Opción 2: Si prefieres otorgar permisos a nivel de secreto individual (más granular)
# Descomenta las siguientes líneas si prefieres este enfoque:

# for SECRET in "${SECRETS[@]}"; do
#     echo "🔐 Otorgando acceso al secreto: $SECRET"
#     gcloud secrets add-iam-policy-binding $SECRET \
#         --project=$PROJECT_ID \
#         --member="serviceAccount:${SERVICE_ACCOUNT}" \
#         --role="roles/secretmanager.secretAccessor"
# done

echo ""
echo "✅ ¡Listo! Ahora intenta desplegar de nuevo en Cloud Run."
echo ""
echo "📝 Si el problema persiste, verifica que:"
echo "   1. Los secretos existen con: gcloud secrets list --project=$PROJECT_ID"
echo "   2. Los secretos tienen valores con: gcloud secrets versions list SECRET_NAME --project=$PROJECT_ID"

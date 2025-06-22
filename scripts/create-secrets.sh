#!/bin/bash
# Script para crear secrets en Google Secret Manager
# Uso: ./create-secrets.sh

set -e

echo "🔐 Creando secrets en Google Secret Manager..."

# Función para crear o actualizar un secret
create_or_update_secret() {
    local SECRET_NAME=$1
    local SECRET_VALUE=$2
    
    # Verificar si el secret existe
    if gcloud secrets describe $SECRET_NAME --project=$PROJECT_ID &>/dev/null; then
        echo "Actualizando secret: $SECRET_NAME"
        echo -n "$SECRET_VALUE" | gcloud secrets versions add $SECRET_NAME --data-file=- --project=$PROJECT_ID
    else
        echo "Creando secret: $SECRET_NAME"
        echo -n "$SECRET_VALUE" | gcloud secrets create $SECRET_NAME --data-file=- --project=$PROJECT_ID
    fi
}

# Verificar que PROJECT_ID esté configurado
if [ -z "$PROJECT_ID" ]; then
    echo "❌ Error: PROJECT_ID no está configurado"
    echo "Ejecuta: export PROJECT_ID=tu-proyecto-id"
    exit 1
fi

echo "📋 Proyecto: $PROJECT_ID"

# Habilitar Secret Manager API si no está habilitada
echo "Habilitando Secret Manager API..."
gcloud services enable secretmanager.googleapis.com --project=$PROJECT_ID

# Database secrets
echo -e "\n🗄️  Configurando secrets de base de datos..."
create_or_update_secret "db-host" "pg-asam-asam-backend-db.l.aivencloud.com"
create_or_update_secret "db-port" "14276"
create_or_update_secret "db-user" "avnadmin"
echo -n "Ingresa DB_PASSWORD: " && read -s DB_PASSWORD && echo
create_or_update_secret "db-password" "$DB_PASSWORD"
create_or_update_secret "db-name" "asam-backend-db"

# JWT secrets
echo -e "\n🔑 Configurando secrets JWT..."
echo "Generando JWT secrets aleatorios..."
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)
create_or_update_secret "jwt-access-secret" "$JWT_ACCESS_SECRET"
create_or_update_secret "jwt-refresh-secret" "$JWT_REFRESH_SECRET"

# Admin credentials
echo -e "\n👤 Configurando credenciales de admin..."
echo -n "Ingresa ADMIN_USER [admin]: " && read ADMIN_USER
ADMIN_USER=${ADMIN_USER:-admin}
echo -n "Ingresa ADMIN_PASSWORD: " && read -s ADMIN_PASSWORD && echo
create_or_update_secret "admin-user" "$ADMIN_USER"
create_or_update_secret "admin-password" "$ADMIN_PASSWORD"

# SMTP (opcional)
echo -e "\n📧 ¿Configurar SMTP? (s/n): " && read CONFIGURE_SMTP
if [ "$CONFIGURE_SMTP" = "s" ]; then
    echo -n "SMTP_USER: " && read SMTP_USER
    echo -n "SMTP_PASSWORD: " && read -s SMTP_PASSWORD && echo
    create_or_update_secret "smtp-user" "$SMTP_USER"
    create_or_update_secret "smtp-password" "$SMTP_PASSWORD"
fi

echo -e "\n✅ Secrets creados exitosamente!"

# Dar permisos a la cuenta de servicio de Cloud Run
echo -e "\n🔧 Configurando permisos para Cloud Run..."
SERVICE_ACCOUNT=$(gcloud iam service-accounts list --filter="displayName:Compute Engine default service account" --format="value(email)" --project=$PROJECT_ID)

if [ -n "$SERVICE_ACCOUNT" ]; then
    echo "Otorgando acceso a secrets para: $SERVICE_ACCOUNT"
    
    # Lista de secrets
    SECRETS=(
        "db-host" "db-port" "db-user" "db-password" "db-name"
        "jwt-access-secret" "jwt-refresh-secret"
        "admin-user" "admin-password"
    )
    
    # Agregar SMTP si se configuró
    if [ "$CONFIGURE_SMTP" = "s" ]; then
        SECRETS+=("smtp-user" "smtp-password")
    fi
    
    # Otorgar permisos
    for SECRET in "${SECRETS[@]}"; do
        gcloud secrets add-iam-policy-binding $SECRET \
            --member="serviceAccount:$SERVICE_ACCOUNT" \
            --role="roles/secretmanager.secretAccessor" \
            --project=$PROJECT_ID &>/dev/null
    done
    
    echo "✅ Permisos configurados"
else
    echo "⚠️  No se pudo encontrar la cuenta de servicio. Configura los permisos manualmente."
fi

echo -e "\n🎉 Configuración completada!"
echo "Los secrets están listos para usar en Cloud Run"

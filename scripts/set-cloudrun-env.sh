#!/bin/bash
# Script to set required environment variables for ASAM Backend in Cloud Run

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SERVICE_NAME="asam-backend"
REGION="europe-west1"

echo -e "${YELLOW}Configurando variables de entorno para $SERVICE_NAME en Cloud Run...${NC}"

# Check if service exists
if ! gcloud run services describe $SERVICE_NAME --region=$REGION &>/dev/null; then
    echo -e "${RED}Error: El servicio $SERVICE_NAME no existe en la región $REGION${NC}"
    exit 1
fi

# Generate secure random secrets if not provided
generate_secret() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

echo -e "${GREEN}Generando secretos seguros...${NC}"
JWT_ACCESS_SECRET=$(generate_secret)
JWT_REFRESH_SECRET=$(generate_secret)

# Prompt for required values
echo -e "${YELLOW}Por favor, proporciona los siguientes valores:${NC}"
read -p "ADMIN_USER (usuario administrador): " ADMIN_USER
read -sp "ADMIN_PASSWORD (contraseña administrador): " ADMIN_PASSWORD
echo
read -p "SMTP_USER (usuario SMTP, dejar vacío si no se usa): " SMTP_USER
read -sp "SMTP_PASSWORD (contraseña SMTP, dejar vacío si no se usa): " SMTP_PASSWORD
echo

# Set SMTP defaults if not provided
SMTP_USER=${SMTP_USER:-"noreply@asam.org"}
SMTP_PASSWORD=${SMTP_PASSWORD:-"temp-smtp-pass"}

# Update environment variables
# Note: PORT is automatically set by Cloud Run and should NOT be overridden
echo -e "\n${YELLOW}Actualizando variables de entorno...${NC}"

gcloud run services update $SERVICE_NAME \
    --region=$REGION \
    --update-env-vars \
ENVIRONMENT=production,\
ADMIN_USER="$ADMIN_USER",\
ADMIN_PASSWORD="$ADMIN_PASSWORD",\
JWT_ACCESS_SECRET="$JWT_ACCESS_SECRET",\
JWT_REFRESH_SECRET="$JWT_REFRESH_SECRET",\
SMTP_USER="$SMTP_USER",\
SMTP_PASSWORD="$SMTP_PASSWORD"

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Variables de entorno actualizadas exitosamente!${NC}"
    echo -e "${YELLOW}Nota: Guarda estos valores en un lugar seguro:${NC}"
    echo "JWT_ACCESS_SECRET=$JWT_ACCESS_SECRET"
    echo "JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET"
else
    echo -e "\n${RED}Error al actualizar las variables de entorno${NC}"
    exit 1
fi

echo -e "\n${GREEN}Configuración completada!${NC}"

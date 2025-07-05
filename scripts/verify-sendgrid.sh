#!/bin/bash
# Script para verificar API Key de SendGrid

# Cargar variables del .env
export $(grep -v '^#' .env | xargs)

echo "Verificando API Key de SendGrid..."
echo "API Key: ${SMTP_PASSWORD:0:10}..."

# Verificar API Key con SendGrid
curl -X GET "https://api.sendgrid.com/v3/scopes" \
  -H "Authorization: Bearer $SMTP_PASSWORD" \
  -H "Content-Type: application/json"

echo ""
echo ""
echo "Si la API Key es válida, verás una lista de scopes/permisos."
echo "Si es inválida, verás un error 401 Unauthorized."

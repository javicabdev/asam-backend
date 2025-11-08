#!/bin/bash

# Script para probar la generación de cuotas anuales
# Uso: ./test_fees.sh

set -e

echo "🧪 Test de Generación de Cuotas Anuales ASAM"
echo "=============================================="
echo ""

# Colores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Verificar servidor
echo -e "${BLUE}📡 Verificando servidor...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Servidor corriendo${NC}"
else
    echo -e "${RED}✗ Servidor no responde en http://localhost:8080${NC}"
    exit 1
fi
echo ""

# 2. Login como admin
echo -e "${BLUE}🔐 Haciendo login como admin...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"operationName":"Login","query":"mutation Login{login(input:{username:\"admin2@admin2.com\",password:\"password123\"}){user{email role} accessToken}}"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"accessToken":"[^"]*' | sed 's/"accessToken":"//')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Error en login${NC}"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Login exitoso${NC}"
echo "Token: ${TOKEN:0:20}..."
echo ""

# 3. Verificar estado actual
echo -e "${BLUE}📊 Verificando estado actual de la BD...${NC}"
echo "Miembros activos:"
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT membership_number, name, surnames, membership_type, state FROM members WHERE state = 'active';" 2>/dev/null || true
echo ""

echo "Cuotas de membresía existentes:"
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT id, year, base_fee_amount, family_fee_extra FROM membership_fees ORDER BY year DESC LIMIT 5;" 2>/dev/null || true
echo ""

echo "Pagos pendientes actuales:"
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT COUNT(*) as total_pending FROM payments WHERE status = 'pending';" 2>/dev/null || true
echo ""

# 4. Ejecutar mutation de generación
echo -e "${YELLOW}🚀 Generando cuotas anuales para 2025...${NC}"
MUTATION_RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"operationName":"GenerateFees","query":"mutation GenerateFees{generateAnnualFees(input:{year:2025,base_fee_amount:100.00,family_fee_extra:50.00}){year membership_fee_id payments_generated payments_existing total_members total_expected_amount details{member_number member_name amount was_created error}}}"}')

echo "Respuesta de la API:"
echo "$MUTATION_RESPONSE" | jq '.' 2>/dev/null || echo "$MUTATION_RESPONSE"
echo ""

# 5. Verificar resultados
echo -e "${BLUE}✅ Verificando resultados en BD...${NC}"
echo ""

echo "Cuota de membresía creada:"
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT id, year, base_fee_amount, family_fee_extra, due_date FROM membership_fees WHERE year = 2025;" 2>/dev/null || true
echo ""

echo "Pagos generados:"
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT p.id, m.membership_number, m.name, m.surnames, p.amount, p.status, p.notes FROM payments p JOIN members m ON p.member_id = m.id WHERE p.membership_fee_id = (SELECT id FROM membership_fees WHERE year = 2025) ORDER BY m.membership_number;" 2>/dev/null || true
echo ""

echo -e "${GREEN}✓ Test completado${NC}"
echo ""
echo "📝 Resumen:"
echo "- Para ver GraphQL Playground: http://localhost:8080/graphql"
echo "- Para probar idempotencia, ejecuta este script de nuevo"
echo "- Para limpiar los datos de prueba, elimina los payments y membership_fees generados"

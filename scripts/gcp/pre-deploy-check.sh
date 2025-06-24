#!/bin/bash

# Script de diagnóstico para verificar la configuración antes de ejecutar el workflow
# Este script ayuda a identificar problemas comunes antes de ejecutar el despliegue

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Parámetros
PROJECT_ID="${1:-$GCP_PROJECT_ID}"
HAS_ERRORS=false

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: Debes proporcionar el PROJECT_ID como parámetro o establecer GCP_PROJECT_ID${NC}"
    echo "Uso: $0 [PROJECT_ID]"
    exit 1
fi

echo -e "${CYAN}=== Pre-Deploy Check para ASAM Backend ===${NC}"
echo -e "${GREEN}Proyecto: $PROJECT_ID${NC}"
echo ""

# Función para marcar errores
mark_error() {
    HAS_ERRORS=true
}

# 1. Verificar Google Cloud SDK
echo -e "${YELLOW}1. Verificando Google Cloud SDK...${NC}"
if command -v gcloud &> /dev/null; then
    GCLOUD_VERSION=$(gcloud version --format="value(Google Cloud SDK)" 2>/dev/null)
    echo -e "   ${GREEN}✅ Google Cloud SDK instalado: $GCLOUD_VERSION${NC}"
else
    echo -e "   ${RED}❌ Google Cloud SDK no está instalado o no está en el PATH${NC}"
    echo -e "   ${CYAN}Instálalo desde: https://cloud.google.com/sdk/docs/install${NC}"
    mark_error
fi

# 2. Verificar autenticación
echo ""
echo -e "${YELLOW}2. Verificando autenticación de Google Cloud...${NC}"
ACCOUNT=$(gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null)
if [ -n "$ACCOUNT" ]; then
    echo -e "   ${GREEN}✅ Autenticado como: $ACCOUNT${NC}"
else
    echo -e "   ${RED}❌ No estás autenticado en Google Cloud${NC}"
    echo -e "   ${CYAN}Ejecuta: gcloud auth login${NC}"
    mark_error
fi

# 3. Verificar proyecto configurado
echo ""
echo -e "${YELLOW}3. Verificando proyecto configurado...${NC}"
gcloud config set project "$PROJECT_ID" 2>/dev/null
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ "$CURRENT_PROJECT" = "$PROJECT_ID" ]; then
    echo -e "   ${GREEN}✅ Proyecto configurado correctamente: $CURRENT_PROJECT${NC}"
else
    echo -e "   ${RED}❌ El proyecto no está configurado correctamente${NC}"
    echo -e "   ${CYAN}Ejecuta: gcloud config set project $PROJECT_ID${NC}"
    mark_error
fi

# 4. Verificar APIs habilitadas
echo ""
echo -e "${YELLOW}4. Verificando APIs habilitadas...${NC}"
REQUIRED_APIS=(
    "run.googleapis.com:Cloud Run API"
    "cloudbuild.googleapis.com:Cloud Build API"
    "containerregistry.googleapis.com:Container Registry API"
    "secretmanager.googleapis.com:Secret Manager API"
)

for api_info in "${REQUIRED_APIS[@]}"; do
    IFS=':' read -r service name <<< "$api_info"
    if gcloud services list --enabled --filter="name:$service" --format="value(name)" 2>/dev/null | grep -q "$service"; then
        echo -e "   ${GREEN}✅ $name habilitada${NC}"
    else
        echo -e "   ${RED}❌ $name NO está habilitada${NC}"
        echo -e "      ${CYAN}Ejecuta: gcloud services enable $service${NC}"
        mark_error
    fi
done

# 5. Verificar cuenta de servicio
echo ""
echo -e "${YELLOW}5. Verificando cuenta de servicio para GitHub Actions...${NC}"
SA_EMAIL="github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com"
if gcloud iam service-accounts describe "$SA_EMAIL" &>/dev/null; then
    echo -e "   ${GREEN}✅ Cuenta de servicio existe: $SA_EMAIL${NC}"
    
    # Verificar roles
    REQUIRED_ROLES=(
        "roles/run.admin"
        "roles/cloudbuild.builds.builder"
        "roles/iam.serviceAccountUser"
        "roles/storage.admin"
        "roles/secretmanager.secretAccessor"
    )
    
    CURRENT_ROLES=$(gcloud projects get-iam-policy "$PROJECT_ID" \
        --flatten="bindings[].members" \
        --filter="bindings.members:serviceAccount:$SA_EMAIL" \
        --format="value(bindings.role)" 2>/dev/null)
    
    echo -e "   ${YELLOW}Verificando roles:${NC}"
    for role in "${REQUIRED_ROLES[@]}"; do
        if echo "$CURRENT_ROLES" | grep -q "$role"; then
            echo -e "      ${GREEN}✅ $role${NC}"
        else
            echo -e "      ${RED}❌ $role faltante${NC}"
            mark_error
        fi
    done
else
    echo -e "   ${RED}❌ Cuenta de servicio NO existe${NC}"
    echo -e "   ${CYAN}Crea la cuenta de servicio y configura los permisos según la documentación${NC}"
    mark_error
fi

# 6. Verificar secretos de base de datos
echo ""
echo -e "${YELLOW}6. Verificando secretos de base de datos...${NC}"
DB_SECRETS=("db-host" "db-port" "db-user" "db-password" "db-name")
ALL_SECRETS_OK=true

for secret in "${DB_SECRETS[@]}"; do
    if gcloud secrets describe "$secret" &>/dev/null; then
        if gcloud secrets versions list "$secret" --limit=1 --format="value(name)" 2>/dev/null | grep -q .; then
            echo -e "   ${GREEN}✅ $secret configurado${NC}"
        else
            echo -e "   ${RED}❌ $secret existe pero no tiene versiones${NC}"
            ALL_SECRETS_OK=false
            mark_error
        fi
    else
        echo -e "   ${RED}❌ $secret NO existe${NC}"
        ALL_SECRETS_OK=false
        mark_error
    fi
done

if [ "$ALL_SECRETS_OK" = false ]; then
    echo ""
    echo -e "   ${CYAN}Para configurar los secretos, ejecuta:${NC}"
    echo -e "   ${WHITE}./verify-db-secrets.sh $PROJECT_ID --create-secrets${NC}"
fi

# 7. Verificar otros secretos necesarios
echo ""
echo -e "${YELLOW}7. Verificando otros secretos necesarios...${NC}"
OTHER_SECRETS=(
    "jwt-access-secret"
    "jwt-refresh-secret"
    "admin-user"
    "admin-password"
)

for secret in "${OTHER_SECRETS[@]}"; do
    if gcloud secrets describe "$secret" &>/dev/null; then
        echo -e "   ${GREEN}✅ $secret configurado${NC}"
    else
        echo -e "   ${YELLOW}⚠️ $secret NO existe (opcional pero recomendado)${NC}"
    fi
done

# 8. Verificar imagen Docker más reciente
echo ""
echo -e "${YELLOW}8. Verificando imágenes Docker disponibles...${NC}"
if gcloud container images list-tags "gcr.io/$PROJECT_ID/asam-backend" --limit=5 --format="table(tags,timestamp)" 2>/dev/null; then
    echo -e "   ${GREEN}✅ Repositorio de imágenes encontrado${NC}"
    echo ""
    echo -e "   ${CYAN}Últimas imágenes disponibles:${NC}"
    gcloud container images list-tags "gcr.io/$PROJECT_ID/asam-backend" --limit=5 --format="table(tags,timestamp)"
else
    echo -e "   ${YELLOW}⚠️ Repositorio de imágenes no encontrado o sin acceso${NC}"
    echo -e "   ${CYAN}Se creará automáticamente en el primer despliegue${NC}"
fi

# Resumen final
echo ""
echo -e "${CYAN}=== Resumen ===${NC}"
echo ""

if [ "$HAS_ERRORS" = true ]; then
    echo -e "${RED}❌ Se encontraron errores que deben solucionarse antes del despliegue${NC}"
    echo ""
    echo -e "${YELLOW}Revisa los errores anteriores y sigue las instrucciones para solucionarlos.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Todo está listo para el despliegue!${NC}"
    echo ""
    echo -e "${CYAN}Próximos pasos:${NC}"
    echo -e "${WHITE}1. Asegúrate de que los secretos en GitHub estén configurados${NC}"
    echo -e "${WHITE}2. Ve a GitHub Actions en tu repositorio${NC}"
    echo -e "${WHITE}3. Ejecuta el workflow 'Deploy to Google Cloud Run'${NC}"
    echo -e "${WHITE}4. Selecciona 'Run database migrations' si necesitas ejecutar migraciones${NC}"
fi

echo ""

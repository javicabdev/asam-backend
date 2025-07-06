#!/bin/bash

# Script para verificar, crear y probar secretos de base de datos en Google Secret Manager
# Este script ayuda a configurar y verificar los secretos necesarios para las migraciones

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

# Parámetros
PROJECT_ID="${1:-$GCP_PROJECT_ID}"
CREATE_SECRETS=false
TEST_CONNECTION=false

# Procesar argumentos
for arg in "$@"; do
    case $arg in
        --create-secrets)
            CREATE_SECRETS=true
            shift
            ;;
        --test-connection)
            TEST_CONNECTION=true
            shift
            ;;
        --help)
            echo "Uso: $0 [PROJECT_ID] [--create-secrets] [--test-connection]"
            echo ""
            echo "Opciones:"
            echo "  --create-secrets    Crear secretos faltantes"
            echo "  --test-connection   Probar la conexión a la base de datos"
            exit 0
            ;;
    esac
done

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: Debes proporcionar el PROJECT_ID como parámetro o establecer GCP_PROJECT_ID${NC}"
    echo "Uso: $0 [PROJECT_ID] [--create-secrets] [--test-connection]"
    exit 1
fi

echo -e "${CYAN}=== Verificando secretos de base de datos en Google Secret Manager ===${NC}"
echo -e "${GREEN}Proyecto: $PROJECT_ID${NC}"
echo ""

# Configurar el proyecto
gcloud config set project "$PROJECT_ID"

# Lista de secretos requeridos
REQUIRED_SECRETS=("db-host" "db-port" "db-user" "db-password" "db-name")
MISSING_SECRETS=()
declare -A SECRET_VALUES

echo -e "${YELLOW}Verificando secretos existentes...${NC}"
echo ""

for secret in "${REQUIRED_SECRETS[@]}"; do
    if gcloud secrets describe "$secret" &>/dev/null; then
        echo -e "${GREEN}✅ $secret existe${NC}"
        
        # Verificar si tiene versiones
        versions=$(gcloud secrets versions list "$secret" --limit=1 --format="value(name)" 2>/dev/null)
        if [ -n "$versions" ]; then
            echo -e "${GRAY}   └─ Tiene versiones activas${NC}"
            
            # Si vamos a hacer test de conexión, obtener el valor
            if [ "$TEST_CONNECTION" = true ]; then
                value=$(gcloud secrets versions access latest --secret="$secret" 2>/dev/null)
                if [ $? -eq 0 ]; then
                    SECRET_VALUES["$secret"]="$value"
                fi
            fi
        else
            echo -e "${YELLOW}   └─ ⚠️ No tiene versiones${NC}"
            MISSING_SECRETS+=("$secret")
        fi
    else
        echo -e "${RED}❌ $secret NO existe${NC}"
        MISSING_SECRETS+=("$secret")
    fi
done

echo ""

# Verificar permisos de la cuenta de servicio
echo -e "${YELLOW}Verificando permisos de la cuenta de servicio...${NC}"
SA_EMAIL="github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com"

# Verificar si la cuenta de servicio existe
if gcloud iam service-accounts describe "$SA_EMAIL" &>/dev/null; then
    # Verificar si tiene el rol secretmanager.secretAccessor
    roles=$(gcloud projects get-iam-policy "$PROJECT_ID" \
        --flatten="bindings[].members" \
        --filter="bindings.members:serviceAccount:$SA_EMAIL" \
        --format="value(bindings.role)" 2>/dev/null)
    
    if echo "$roles" | grep -q "roles/secretmanager.secretAccessor"; then
        echo -e "${GREEN}✅ La cuenta de servicio tiene acceso a los secretos${NC}"
    else
        echo -e "${YELLOW}⚠️ La cuenta de servicio NO tiene el rol secretmanager.secretAccessor${NC}"
        echo ""
        echo -e "${CYAN}Para agregar el rol, ejecuta:${NC}"
        echo "gcloud projects add-iam-policy-binding $PROJECT_ID \\"
        echo "  --member='serviceAccount:$SA_EMAIL' \\"
        echo "  --role='roles/secretmanager.secretAccessor'"
    fi
else
    echo -e "${YELLOW}⚠️ La cuenta de servicio $SA_EMAIL no existe${NC}"
    echo ""
    echo -e "${CYAN}Para crear la cuenta de servicio, ejecuta:${NC}"
    echo "gcloud iam service-accounts create github-actions-deploy \\"
    echo "  --display-name='GitHub Actions Deploy Service Account'"
fi

echo ""

if [ ${#MISSING_SECRETS[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ Todos los secretos están configurados!${NC}"
    
    if [ "$TEST_CONNECTION" = true ] && [ ${#SECRET_VALUES[@]} -eq 5 ]; then
        echo ""
        echo -e "${YELLOW}Probando conexión a la base de datos...${NC}"
        
        # Obtener valores de los secretos
        host="${SECRET_VALUES[db-host]}"
        port="${SECRET_VALUES[db-port]}"
        user="${SECRET_VALUES[db-user]}"
        password="${SECRET_VALUES[db-password]}"
        dbname="${SECRET_VALUES[db-name]}"
        
        echo -e "${GRAY}Host: $host:$port${NC}"
        echo -e "${GRAY}Database: $dbname${NC}"
        echo -e "${GRAY}User: $user${NC}"
        
        # Usar psql si está disponible
        if command -v psql &> /dev/null; then
            export PGPASSWORD="$password"
            if result=$(psql -h "$host" -p "$port" -U "$user" -d "$dbname" -c "SELECT version();" 2>&1); then
                echo -e "${GREEN}✅ Conexión exitosa!${NC}"
                echo -e "${GRAY}$result${NC}"
            else
                echo -e "${RED}❌ Error al conectar:${NC}"
                echo -e "${RED}$result${NC}"
            fi
            unset PGPASSWORD
        else
            echo -e "${YELLOW}⚠️ psql no está instalado. No se puede probar la conexión.${NC}"
            echo -e "${CYAN}Para instalar psql:${NC}"
            echo "  - Ubuntu/Debian: sudo apt-get install postgresql-client"
            echo "  - Mac: brew install postgresql"
            echo "  - RHEL/CentOS: sudo yum install postgresql"
        fi
    fi
    
else
    echo -e "${RED}❌ Faltan los siguientes secretos:${NC}"
    for secret in "${MISSING_SECRETS[@]}"; do
        echo -e "${RED}   - $secret${NC}"
    done
    
    if [ "$CREATE_SECRETS" = true ]; then
        echo ""
        echo -e "${YELLOW}Creando secretos faltantes...${NC}"
        echo ""
        echo -e "${CYAN}Por favor, ingresa los valores para cada secreto:${NC}"
        
        for secret in "${MISSING_SECRETS[@]}"; do
            echo ""
            case "$secret" in
                "db-host")
                    read -p "DB Host (ej: pg-xxx.aivencloud.com): " value
                    ;;
                "db-port")
                    read -p "DB Port (ej: 14276): " value
                    ;;
                "db-user")
                    read -p "DB User (ej: avnadmin): " value
                    ;;
                "db-password")
                    read -s -p "DB Password: " value
                    echo ""
                    ;;
                "db-name")
                    read -p "DB Name (ej: defaultdb): " value
                    ;;
            esac
            
            if [ -n "$value" ]; then
                echo -e "${YELLOW}Creando secreto $secret...${NC}"
                
                # Verificar si el secreto existe pero no tiene versiones
                if gcloud secrets describe "$secret" &>/dev/null; then
                    # El secreto existe, solo agregar una versión
                    echo -n "$value" | gcloud secrets versions add "$secret" --data-file=-
                else
                    # Crear el secreto y agregar la versión
                    echo -n "$value" | gcloud secrets create "$secret" --data-file=-
                fi
                
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✅ $secret creado exitosamente${NC}"
                else
                    echo -e "${RED}❌ Error al crear $secret${NC}"
                fi
            fi
        done
        
        echo ""
        echo -e "${GREEN}✅ Proceso completado!${NC}"
        
        # Otorgar permisos a la cuenta de servicio
        echo ""
        echo -e "${YELLOW}Otorgando permisos a la cuenta de servicio...${NC}"
        
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$SA_EMAIL" \
            --role="roles/secretmanager.secretAccessor" \
            --quiet
            
        echo -e "${GREEN}✅ Permisos otorgados${NC}"
        
    else
        echo ""
        echo -e "${CYAN}Para crear los secretos, ejecuta:${NC}"
        echo "$0 $PROJECT_ID --create-secrets"
    fi
fi

echo ""
echo -e "${CYAN}=== Resumen ===${NC}"
echo ""
echo -e "${YELLOW}Si necesitas actualizar un secreto existente:${NC}"
echo 'echo "nuevo-valor" | gcloud secrets versions add <secret-name> --data-file=-'
echo ""
echo -e "${YELLOW}Para ver el valor actual de un secreto:${NC}"
echo "gcloud secrets versions access latest --secret=<secret-name>"
echo ""
echo -e "${YELLOW}Para probar la conexión a la base de datos:${NC}"
echo "$0 $PROJECT_ID --test-connection"
echo ""

#!/bin/bash

# Script para arrancar ASAM Backend localmente
# Este script facilita el arranque del proyecto con Docker

# Colores para output
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
GRAY='\033[0;90m'
DARK_GREEN='\033[2;32m'
DARK_GRAY='\033[2;37m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
cat << "EOF"
╔═══════════════════════════════════════╗
║       ASAM Backend - Arranque Local   ║
╚═══════════════════════════════════════╝
EOF
echo -e "${NC}"

# Verificar Docker
echo -e "${YELLOW}🔍 Verificando Docker...${NC}"
if command -v docker &> /dev/null && command -v docker-compose &> /dev/null; then
    echo -e "${GREEN}✅ Docker está instalado y funcionando${NC}"
else
    echo -e "${RED}❌ Docker no está instalado o no está funcionando${NC}"
    echo -e "${YELLOW}   Por favor instala Docker Desktop desde: https://www.docker.com/products/docker-desktop${NC}"
    exit 1
fi

# Verificar Go (opcional, solo para desarrollo)
echo -e "\n${YELLOW}🔍 Verificando Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo -e "${GREEN}✅ $GO_VERSION${NC}"
else
    echo -e "${YELLOW}⚠️  Go no está instalado (opcional para solo ejecutar con Docker)${NC}"
fi

# Detener contenedores previos
echo -e "\n${YELLOW}🛑 Deteniendo contenedores previos...${NC}"
docker-compose down --remove-orphans 2>/dev/null

# Limpiar volúmenes si se especifica
if [[ " $* " == *" --clean "* ]]; then
    echo -e "\n${YELLOW}🧹 Limpieza completa del entorno...${NC}"

    # Detener y limpiar
    echo -e "${GRAY}   Deteniendo todos los contenedores...${NC}"
    docker-compose down -v --remove-orphans

    # Limpiar contenedores huérfanos adicionales
    echo -e "${GRAY}   Eliminando contenedores huérfanos...${NC}"
    docker container prune -f 2>/dev/null

    # Limpiar redes no utilizadas
    echo -e "${GRAY}   Limpiando redes no utilizadas...${NC}"
    docker network prune -f 2>/dev/null

    echo -e "${GREEN}✅ Limpieza completa finalizada${NC}"
    sleep 2
fi

# Verificar que existe el archivo de configuración
echo -e "\n${YELLOW}📋 Verificando archivo de configuración...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ Error: No se encontró el archivo .env${NC}"
    echo -e "${YELLOW}\n⚠️  El archivo .env es requerido para arrancar el proyecto.${NC}"
    echo -e "\n${CYAN}Para crear el archivo de configuración:${NC}"
    echo -e "${GRAY}1. Copia el archivo de ejemplo:${NC}"
    echo -e "   ${GREEN}cp .env.example .env${NC}"
    echo -e "\n${GRAY}2. Edita .env y configura:${NC}"
    echo -e "   - ${GREEN}MAILERSEND_API_KEY${NC} con tu API key real de MailerSend"
    echo -e "   - Otras variables según necesites"
    echo -e "\n${GRAY}3. Vuelve a ejecutar este script${NC}\n"
    exit 1
fi

echo -e "${GREEN}✅ Archivo .env encontrado${NC}"

# Verificar si hay problemas con contenedores existentes
echo -e "\n${YELLOW}🔍 Verificando estado de contenedores...${NC}"
EXISTING_CONTAINERS=$(docker ps -a --filter "name=asam" --format "{{.Names}} {{.Status}}" 2>/dev/null)
if [ -n "$EXISTING_CONTAINERS" ]; then
    if echo "$EXISTING_CONTAINERS" | grep -q "Exited\|Dead"; then
        echo -e "${YELLOW}   ⚠️  Detectados contenedores en mal estado${NC}"
        echo -e "${GRAY}   Limpiando contenedores problemáticos...${NC}"
        docker-compose down -v --remove-orphans
        sleep 2
    fi
fi

# Construir y arrancar servicios
echo -e "\n${YELLOW}🚀 Construyendo y arrancando servicios...${NC}"
docker-compose up -d --build

# Verificar si los contenedores arrancaron correctamente
sleep 3
API_STATUS=$(docker ps --filter "name=asam-backend-api" --format "{{.Status}}" 2>/dev/null)
DB_STATUS=$(docker ps --filter "name=asam-postgres" --format "{{.Status}}" 2>/dev/null)

if [ -z "$API_STATUS" ] || [ -z "$DB_STATUS" ]; then
    echo -e "${RED}❌ Error: Los contenedores no arrancaron correctamente${NC}"
    echo -e "${YELLOW}   Intenta ejecutar: ./scripts/reset-emergency.sh${NC}"
    echo -e "${YELLOW}   Y luego: ./start-docker.sh${NC}"
    exit 1
fi

# Esperar a que PostgreSQL esté listo
echo -e "\n${YELLOW}⏳ Esperando a que PostgreSQL esté listo...${NC}"
MAX_ATTEMPTS=30
ATTEMPT=0
READY=false

while [ $ATTEMPT -lt $MAX_ATTEMPTS ] && [ "$READY" = false ]; do
    ATTEMPT=$((ATTEMPT + 1))
    echo -n "."

    if docker-compose exec -T postgres pg_isready -U postgres -d asam_db &>/dev/null; then
        READY=true
    else
        sleep 1
    fi
done

echo ""
if [ "$READY" = true ]; then
    echo -e "${GREEN}✅ PostgreSQL está listo${NC}"
else
    echo -e "${RED}❌ PostgreSQL no está respondiendo${NC}"
    exit 1
fi

# Ejecutar migraciones
echo -e "\n${YELLOW}🔄 Ejecutando migraciones...${NC}"
# Esperar un poco más para asegurar que el API esté lista
sleep 3

# Siempre ejecutar migraciones para asegurar que todas estén aplicadas
echo -e "${GRAY}   Verificando y aplicando todas las migraciones...${NC}"

# Intentar ejecutar migraciones con el comando Go
echo -e "${GRAY}   Ejecutando migraciones...${NC}"
docker-compose exec -T api go run ./cmd/migrate -env local up

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Migraciones ejecutadas con éxito${NC}"
else
    echo -e "${YELLOW}⚠️  Error al ejecutar migraciones con Go${NC}"

    # Como respaldo, intentar ejecutar las migraciones SQL directamente
    echo -e "${GRAY}   Intentando ejecutar migraciones SQL directamente...${NC}"

    # Obtener todos los archivos de migración .up.sql ordenados
    if [ -d "migrations" ]; then
        for MIGRATION_FILE in $(ls migrations/*.up.sql 2>/dev/null | sort); do
            MIGRATION_NAME=$(basename "$MIGRATION_FILE")
            echo -e "${GRAY}   Aplicando: $MIGRATION_NAME${NC}"

            # Ejecutar la migración
            if docker-compose exec -T postgres psql -U postgres -d asam_db < "$MIGRATION_FILE" &>/dev/null; then
                echo -e "${DARK_GREEN}   ✓ $MIGRATION_NAME aplicada${NC}"
            else
                # Ignorar errores de "already exists" ya que es esperado
                echo -e "${DARK_GRAY}   ~ $MIGRATION_NAME (ya aplicada o error menor)${NC}"
            fi
        done
    fi

    echo -e "${GREEN}✅ Proceso de migraciones completado${NC}"
fi

# Verificar que las tablas principales existen
sleep 2
VERIFY_TABLES=$(docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'members', 'families', 'payments', 'cash_flows');" 2>/dev/null | tr -d ' ')

if [ -n "$VERIFY_TABLES" ]; then
    VERIFY_COUNT=$VERIFY_TABLES
    if [ "$VERIFY_COUNT" -eq 5 ]; then
        echo -e "${GREEN}✅ Verificado: Todas las tablas principales existen${NC}"
    elif [ "$VERIFY_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Solo $VERIFY_COUNT de 5 tablas principales fueron creadas${NC}"
    else
        echo -e "${RED}❌ Error: No se crearon las tablas principales${NC}"
        echo -e "${YELLOW}   Intenta ejecutar: ./scripts/reset-emergency.sh${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  No se pudo verificar la creación de tablas${NC}"
fi

# Crear usuarios de prueba usando la herramienta de gestión de usuarios
echo -e "\n${YELLOW}👥 Creando usuarios de prueba...${NC}"

# Verificar si ya existen usuarios
USER_COUNT=$(docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null | tr -d ' ')

if [ $? -ne 0 ] || [ -z "$USER_COUNT" ]; then
    echo -e "${YELLOW}   La tabla users no existe, necesita ejecutar migraciones primero${NC}"
    USER_COUNT_INT=0
else
    USER_COUNT_INT=$USER_COUNT
fi

if [ "$USER_COUNT_INT" -eq 0 ]; then
    echo -e "${GRAY}   No hay usuarios, creando usuarios de prueba...${NC}"
    # Esperar un poco para asegurar que el API esté completamente lista
    sleep 2

    # Usar el script automatizado que no requiere interacción
    if docker-compose exec -T api go run scripts/user-management/auto-create-test-users/auto-create-test-users.go; then
        echo -e "${GREEN}✅ Usuarios de prueba creados correctamente${NC}"

        # Verificar que los usuarios se crearon
        NEW_USER_COUNT=$(docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null | tr -d ' ')
        if [ -n "$NEW_USER_COUNT" ]; then
            echo -e "${GRAY}   Total de usuarios en la base de datos: $NEW_USER_COUNT${NC}"
        else
            echo -e "${YELLOW}   No se pudo verificar el número de usuarios creados${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Error al crear usuarios con el script${NC}"
        echo -e "${YELLOW}   Intenta ejecutar manualmente: make db-seed${NC}"
    fi
else
    echo -e "${GREEN}✅ Ya existen $USER_COUNT_INT usuarios en la base de datos${NC}"

    # Mostrar los usuarios existentes
    echo -e "${GRAY}   Usuarios existentes:${NC}"
    docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT username, role FROM users;" | while read -r line; do
        if [ -n "$(echo "$line" | tr -d ' ')" ]; then
            echo -e "${DARK_GRAY}   - $line${NC}"
        fi
    done
fi

# Verificación final antes de mostrar logs
FINAL_USER_CHECK=$(docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users WHERE username IN ('admin', 'user');" 2>/dev/null | tr -d ' ')

if [ $? -ne 0 ] || [ -z "$FINAL_USER_CHECK" ]; then
    FINAL_USER_COUNT=0
else
    FINAL_USER_COUNT=$FINAL_USER_CHECK
fi

if [ "$FINAL_USER_COUNT" -lt 2 ]; then
    echo -e "\n${YELLOW}⚠️  ADVERTENCIA: Los usuarios de prueba no se crearon correctamente${NC}"
    echo -e "${CYAN}   Solución rápida: ./scripts/auto-fix.sh${NC}"
    echo -e "${YELLOW}   O manualmente: docker-compose exec api go run scripts/user-management/auto-create-test-users.go${NC}"
    echo ""
    echo -e "${GRAY}   Para diagnóstico completo: ./scripts/diagnostico.sh${NC}"
fi

# Mostrar logs en tiempo real
echo -e "\n${YELLOW}📜 Mostrando logs de la aplicación...${NC}"
echo -e "${GRAY}   (Presiona Ctrl+C para detener los logs)${NC}"
echo ""

# Mostrar información de acceso
echo -e "${CYAN}"
cat << 'EOF'

╔════════════════════════════════════════════════════════════╗
║                    ASAM Backend Activo                     ║
╠════════════════════════════════════════════════════════════╣
║  🌐 GraphQL Playground: http://localhost:8080/playground   ║
║  🔧 API Endpoint:      http://localhost:8080/graphql      ║
║  ❤️  Health Check:     http://localhost:8080/health       ║
║  📊 Metrics:          http://localhost:8080/metrics       ║
╠════════════════════════════════════════════════════════════╣
║                  Usuarios de Prueba:                       ║
║  👤 Admin:     admin / AsamAdmin2025!                     ║
║  👤 Usuario:   user  / AsamUser2025!                      ║
╠════════════════════════════════════════════════════════════╣
║  🛑 Para detener: docker-compose down                      ║
║  🧹 Limpiar todo: ./start-docker.sh --clean               ║
╠════════════════════════════════════════════════════════════╣
║  🔧 ¿Problemas? Ejecuta: ./scripts/auto-fix.sh            ║
║  📊 Diagnóstico: ./scripts/diagnostico.sh                 ║
║  ❓ Ver ayuda: ./scripts/help.sh                           ║
╚════════════════════════════════════════════════════════════╝

EOF
echo -e "${NC}"

# Seguir logs
docker-compose logs -f api

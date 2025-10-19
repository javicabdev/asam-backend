#!/bin/bash

# Script de emergencia para resetear completamente el entorno Docker
# Usa este script cuando tengas problemas con contenedores o puertos

CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}"
cat << "EOF"
╔═══════════════════════════════════════╗
║    ASAM Backend - Reset de Emergencia ║
╚═══════════════════════════════════════╝
EOF
echo -e "${NC}"

echo -e "${YELLOW}⚠️  Este script va a:${NC}"
echo -e "   1. Detener TODOS los contenedores de Docker"
echo -e "   2. Eliminar volúmenes de datos (SE PERDERÁN DATOS)"
echo -e "   3. Limpiar redes y contenedores huérfanos"
echo -e "   4. Resetear configuración de puertos si hay conflictos"
echo ""
read -p "$(echo -e ${YELLOW}"¿Continuar? (y/n): "${NC})" -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${CYAN}Operación cancelada${NC}"
    exit 0
fi

# Detener todos los contenedores ASAM
echo -e "\n${YELLOW}🛑 Deteniendo contenedores ASAM...${NC}"
docker-compose down -v --remove-orphans

# Forzar eliminación de contenedores específicos si aún existen
echo -e "\n${YELLOW}🗑️  Eliminando contenedores específicos...${NC}"
docker rm -f asam-backend-api asam-postgres 2>/dev/null || true

# Limpiar contenedores detenidos
echo -e "\n${YELLOW}🧹 Limpiando contenedores detenidos...${NC}"
docker container prune -f

# Limpiar volúmenes no utilizados
echo -e "\n${YELLOW}🧹 Limpiando volúmenes...${NC}"
docker volume prune -f

# Limpiar redes no utilizadas
echo -e "\n${YELLOW}🧹 Limpiando redes...${NC}"
docker network prune -f

# Verificar si el puerto 5432 sigue ocupado
echo -e "\n${YELLOW}🔍 Verificando puerto 5432...${NC}"
if lsof -Pi :5432 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${RED}❌ Puerto 5432 aún está ocupado por otro proceso${NC}"
    echo -e "${CYAN}Procesos usando el puerto 5432:${NC}"
    lsof -i :5432
    echo ""
    echo -e "${YELLOW}OPCIONES:${NC}"
    echo -e "  1. Cambiar puerto de Docker a 5433 (recomendado)"
    echo -e "  2. Detener PostgreSQL local: ${CYAN}brew services stop postgresql${NC}"
    echo ""
    read -p "$(echo -e ${YELLOW}"¿Cambiar puerto de Docker a 5433? (y/n): "${NC})" -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Actualizar docker-compose.yml para usar puerto alternativo
        if [ -f "docker-compose.yml" ]; then
            echo -e "${YELLOW}📝 Actualizando puerto en docker-compose.yml...${NC}"
            sed -i.bak 's/\${DB_PORT:-5432}:5432/${DB_PORT:-5433}:5432/' docker-compose.yml
            rm -f docker-compose.yml.bak
            
            # Actualizar .env si existe
            if [ -f ".env" ]; then
                if grep -q "^DB_PORT=" .env; then
                    sed -i.bak 's/^DB_PORT=.*/DB_PORT=5433/' .env
                else
                    echo "DB_PORT=5433" >> .env
                fi
                rm -f .env.bak
            fi
            
            echo -e "${GREEN}✅ Puerto actualizado a 5433${NC}"
            echo -e "${CYAN}   PostgreSQL de Docker estará en: localhost:5433${NC}"
            echo -e "${CYAN}   PostgreSQL local sigue en: localhost:5432${NC}"
        fi
    fi
else
    echo -e "${GREEN}✅ Puerto 5432 está disponible${NC}"
fi

# Verificar puerto 8080
echo -e "\n${YELLOW}🔍 Verificando puerto 8080...${NC}"
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${RED}❌ Puerto 8080 está ocupado${NC}"
    echo -e "${CYAN}Procesos usando el puerto 8080:${NC}"
    lsof -i :8080
else
    echo -e "${GREEN}✅ Puerto 8080 está disponible${NC}"
fi

echo -e "\n${GREEN}✅ Reset de emergencia completado${NC}"
echo -e "${CYAN}"
echo "═════════════════════════════════════════════"
echo "Siguiente paso: ejecutar el arranque normal"
echo "  ./start-docker.sh"
echo "═════════════════════════════════════════════"
echo -e "${NC}"

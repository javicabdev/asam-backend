#!/bin/bash

# Script de verificación de configuración de seguridad
# Verifica que SAST (gosec) esté correctamente configurado

set -e

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔍 Verificando configuración de seguridad SAST..."
echo ""

# 1. Verificar que gosec está instalado
echo -n "1. Verificando instalación de gosec... "
if command -v gosec &> /dev/null; then
    VERSION=$(gosec -version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} Instalado ($VERSION)"
else
    echo -e "${RED}✗${NC} No encontrado"
    echo -e "${YELLOW}   Ejecuta: make tools${NC}"
    exit 1
fi

# 2. Verificar archivo de configuración .gosec.json
echo -n "2. Verificando archivo .gosec.json... "
if [ -f ".gosec.json" ]; then
    echo -e "${GREEN}✓${NC} Existe"
    # Validar que sea JSON válido
    if jq empty .gosec.json 2>/dev/null; then
        echo -e "${GREEN}   ✓${NC} JSON válido"
    else
        echo -e "${YELLOW}   ⚠${NC} JSON podría tener errores"
    fi
else
    echo -e "${RED}✗${NC} No encontrado"
    exit 1
fi

# 3. Verificar workflow de CI
echo -n "3. Verificando workflow de CI... "
if [ -f ".github/workflows/ci.yml" ]; then
    echo -e "${GREEN}✓${NC} Existe"

    # Verificar que tiene el job de security
    if grep -q "job.*security" .github/workflows/ci.yml || grep -q "Security Scan" .github/workflows/ci.yml; then
        echo -e "${GREEN}   ✓${NC} Job 'security' encontrado"
    else
        echo -e "${RED}   ✗${NC} Job 'security' no encontrado"
    fi

    # Verificar permisos de security-events
    if grep -q "security-events: write" .github/workflows/ci.yml; then
        echo -e "${GREEN}   ✓${NC} Permisos 'security-events: write' configurados"
    else
        echo -e "${YELLOW}   ⚠${NC} Falta permiso 'security-events: write'"
    fi

    # Verificar upload SARIF
    if grep -q "upload-sarif" .github/workflows/ci.yml; then
        echo -e "${GREEN}   ✓${NC} Upload SARIF configurado"
    else
        echo -e "${YELLOW}   ⚠${NC} Upload SARIF no encontrado"
    fi
else
    echo -e "${RED}✗${NC} No encontrado"
    exit 1
fi

# 4. Verificar comandos en Makefile
echo -n "4. Verificando comandos en Makefile... "
if [ -f "Makefile" ]; then
    echo -e "${GREEN}✓${NC} Existe"

    if grep -q "^security:" Makefile; then
        echo -e "${GREEN}   ✓${NC} Comando 'make security' disponible"
    else
        echo -e "${RED}   ✗${NC} Comando 'make security' no encontrado"
    fi

    if grep -q "^security-ci:" Makefile; then
        echo -e "${GREEN}   ✓${NC} Comando 'make security-ci' disponible"
    else
        echo -e "${YELLOW}   ⚠${NC} Comando 'make security-ci' no encontrado"
    fi
else
    echo -e "${RED}✗${NC} No encontrado"
fi

# 5. Verificar que gosec-report.json está en .gitignore
echo -n "5. Verificando .gitignore... "
if [ -f ".gitignore" ]; then
    if grep -q "gosec-report.json" .gitignore && grep -q "gosec-results.sarif" .gitignore; then
        echo -e "${GREEN}✓${NC} Archivos de reporte excluidos"
    else
        echo -e "${YELLOW}⚠${NC} Algunos archivos de reporte no están en .gitignore"
    fi
else
    echo -e "${YELLOW}⚠${NC} .gitignore no encontrado"
fi

# 6. Ejecutar un análisis de prueba
echo ""
echo "6. Ejecutando análisis de prueba..."
if gosec -conf .gosec.json -fmt json -out /tmp/gosec-test.json ./... 2>&1 | tail -5; then
    echo -e "${GREEN}✓${NC} Análisis completado"

    # Mostrar estadísticas
    if [ -f "/tmp/gosec-test.json" ]; then
        if command -v jq &> /dev/null; then
            FILES=$(jq -r '.Stats.files // 0' /tmp/gosec-test.json)
            LINES=$(jq -r '.Stats.lines // 0' /tmp/gosec-test.json)
            ISSUES=$(jq -r '.Stats.found // 0' /tmp/gosec-test.json)

            echo ""
            echo "   📊 Estadísticas del análisis:"
            echo "   - Archivos analizados: $FILES"
            echo "   - Líneas de código: $LINES"
            echo "   - Issues encontrados: $ISSUES"

            if [ "$ISSUES" -eq 0 ]; then
                echo -e "   ${GREEN}🎉 No se encontraron vulnerabilidades!${NC}"
            elif [ "$ISSUES" -lt 5 ]; then
                echo -e "   ${YELLOW}⚠️  Pocas vulnerabilidades encontradas${NC}"
            else
                echo -e "   ${YELLOW}⚠️  Múltiples vulnerabilidades encontradas${NC}"
            fi
        fi
        rm -f /tmp/gosec-test.json
    fi
else
    echo -e "${RED}✗${NC} Error en el análisis"
    exit 1
fi

# 7. Verificar documentación
echo ""
echo -n "7. Verificando documentación... "
DOCS_FOUND=0
if [ -f "SECURITY.md" ]; then
    echo -e "${GREEN}✓${NC} SECURITY.md existe"
    ((DOCS_FOUND++))
else
    echo -e "${YELLOW}⚠${NC} SECURITY.md no encontrado"
fi

if [ -f "docs/SAST-DAST-GUIDE.md" ]; then
    echo -e "${GREEN}   ✓${NC} SAST-DAST-GUIDE.md existe"
    ((DOCS_FOUND++))
else
    echo -e "${YELLOW}   ⚠${NC} SAST-DAST-GUIDE.md no encontrado"
fi

if [ -f "docs/GITHUB-SECURITY-SETUP.md" ]; then
    echo -e "${GREEN}   ✓${NC} GITHUB-SECURITY-SETUP.md existe"
    ((DOCS_FOUND++))
else
    echo -e "${YELLOW}   ⚠${NC} GITHUB-SECURITY-SETUP.md no encontrado"
fi

# Resumen final
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Verificación completada${NC}"
echo ""
echo "Próximos pasos:"
echo "1. Ejecuta 'make security' para análisis local completo"
echo "2. Haz commit y push para activar GitHub Security"
echo "3. Ve a https://github.com/[tu-repo]/security/code-scanning"
echo "4. Configura branch protection en Settings > Branches"
echo ""
echo "Documentación:"
echo "- SECURITY.md - Política de seguridad"
echo "- docs/SAST-DAST-GUIDE.md - Guía de SAST/DAST"
echo "- docs/GITHUB-SECURITY-SETUP.md - Setup de GitHub Security"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

#!/bin/bash
# Script para instalar hooks de git en sistemas Unix

echo "Instalando hooks de git..."

# Crear directorio de hooks si no existe
mkdir -p .git/hooks

# Copiar el hook de pre-commit
cp scripts/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

echo "Hooks instalados correctamente."

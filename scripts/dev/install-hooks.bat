@echo off
:: Script para instalar hooks de git en Windows
echo Instalando hooks de git...

:: Crear directorio de hooks si no existe
if not exist .git\hooks mkdir .git\hooks

:: Copiar el hook de pre-commit
copy /Y scripts\pre-commit .git\hooks\pre-commit

echo Hooks instalados correctamente.
echo Para utilizar los hooks en Windows, necesitas Git Bash o un entorno similar.

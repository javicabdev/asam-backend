@echo off
echo =============================
echo Gestion de Usuarios ASAM
echo =============================
echo.

cd /d "%~dp0"
cd ../..

echo Instalando dependencias...
go mod download

echo.
echo Ejecutando gestor de usuarios...
go run scripts/user-management/manage_users.go

pause
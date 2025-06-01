@echo off
echo ====================================
echo Ejecutando pruebas de base de datos
echo en produccion
echo ====================================
echo.

cd /d "%~dp0"
cd ../..

echo Instalando dependencias...
go mod download

echo.
echo Ejecutando pruebas...
go run scripts/prod-tests/test_database_operations.go

echo.
echo ====================================
echo Pruebas completadas
echo ====================================
pause

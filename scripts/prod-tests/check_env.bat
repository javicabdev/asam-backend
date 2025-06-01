@echo off
echo ====================================
echo Verificando variables de entorno
echo ====================================
echo.

cd /d "%~dp0"
cd ../..

go run scripts/prod-tests/check-env/main.go

echo.
pause

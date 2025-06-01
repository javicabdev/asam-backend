@echo off
echo ====================================
echo Prueba rapida de conexion a BD
echo ====================================
echo.

cd /d "%~dp0"
cd ../..

go run scripts/prod-tests/quick_connection_check.go

echo.
pause

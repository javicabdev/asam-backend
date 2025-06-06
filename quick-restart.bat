@echo off
echo 🔄 Aplicando nueva configuración y reiniciando API...
echo.

echo 🛑 Deteniendo API...
docker-compose stop api

echo.
echo 🚀 Iniciando API con nueva configuración...
docker-compose up -d api

echo.
echo ⏳ Esperando 10 segundos...
timeout /t 10 /nobreak > nul

echo.
echo 📋 Mostrando logs del API:
docker-compose logs --tail=50 api

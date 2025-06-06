@echo off
echo 🚀 Reiniciando API con JWT corregido...
docker-compose restart api
echo.
echo ⏳ Esperando 10 segundos...
timeout /t 10 /nobreak > nul
echo.
echo 📋 Estado:
docker-compose ps api
echo.
echo 🔍 Para ver logs: docker-compose logs -f api

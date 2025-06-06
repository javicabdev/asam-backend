@echo off
echo 🔍 Verificando versiones...
echo.

echo Docker version:
docker --version

echo.
echo Docker Compose version:
docker-compose --version

echo.
echo Go version en go.mod:
findstr "^go " go.mod

echo.
echo Dockerfile Go version:
findstr "FROM golang:" Dockerfile.dev | findstr -v "#"

echo.
echo Variables de entorno necesarias en .env:
findstr "POSTGRES_DB=" .env
findstr "DB_HOST=" .env
findstr "JWT_SECRET=" .env

echo.
echo ✅ Si todo se ve bien, ejecuta:
echo    .\clean-start-docker.bat
echo.
pause

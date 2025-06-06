@echo off
echo 🔄 Aplicando corrección de JWT y reiniciando...
echo.

echo 🛑 Deteniendo API...
docker-compose stop api

echo.
echo ✅ JWT_REFRESH_TTL corregido a 168h (7 días)

echo.
echo 🚀 Iniciando API...
docker-compose up -d api

echo.
echo ⏳ Esperando 15 segundos para que el API inicie...
timeout /t 15 /nobreak > nul

echo.
echo 📊 Ejecutando migraciones...
docker-compose exec api go run cmd/migrate/main.go -env=local -cmd=up
if %errorlevel% neq 0 (
    echo ⚠️  Las migraciones fallaron, intentando de nuevo...
    timeout /t 5 /nobreak > nul
    docker-compose exec api go run cmd/migrate/main.go -env=local -cmd=up
)

echo.
echo 🌱 Ejecutando seed...
docker-compose exec api go run cmd/seed/main.go -env=local -type=minimal

echo.
echo 👤 Creando usuarios de prueba...
docker-compose exec -T postgres psql -U postgres -d asam_db < scripts/create-test-users.sql

echo.
echo 🔍 Verificando estado...
docker-compose ps

echo.
echo 📋 Últimos logs del API:
docker-compose logs --tail=20 api | findstr /I "successfully server listening"

echo.
echo ✅ ¡Proceso completado!
echo.
echo 🌐 Prueba el GraphQL Playground: http://localhost:8080/playground
echo 🚀 Prueba el login: http://localhost:5173
echo.
pause

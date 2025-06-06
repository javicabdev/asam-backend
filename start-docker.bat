@echo off
echo 🐳 Iniciando ASAM Backend con Docker...
echo.

REM Detener contenedores existentes
echo 🛑 Deteniendo contenedores existentes...
docker-compose down

echo.
echo 🏗️ Construyendo e iniciando servicios...
docker-compose up -d --build

echo.
echo ⏳ Esperando a que PostgreSQL esté listo...
timeout /t 10 /nobreak > nul

echo.
echo 📊 Ejecutando migraciones...
docker-compose exec api go run cmd/migrate/main.go -env=local -cmd=up

echo.
echo 🌱 Ejecutando seed con datos de prueba...
docker-compose exec api go run cmd/seed/main.go -env=local -type=minimal

echo.
echo 👤 Creando usuarios de prueba...
docker-compose exec -T postgres psql -U postgres -d asam_db < scripts/create-test-users.sql

echo.
echo ✅ ¡ASAM Backend está corriendo!
echo.
echo 📍 URLs disponibles:
echo    - API: http://localhost:8080
echo    - GraphQL Playground: http://localhost:8080/playground
echo    - PostgreSQL: localhost:5432
echo.
echo 🔐 Credenciales de prueba:
echo    - Email: admin@asam.org
echo    - Password: admin123
echo.
echo 📋 Comandos útiles:
echo    - Ver logs: docker-compose logs -f api
echo    - Detener: docker-compose down
echo    - Reiniciar: docker-compose restart
echo.
pause

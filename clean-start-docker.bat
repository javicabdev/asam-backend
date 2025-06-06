@echo off
echo 🧹 Limpiando y reiniciando ASAM Backend con Docker...
echo.

echo 🛑 Deteniendo todos los contenedores...
docker-compose down -v

echo.
echo 🗑️ Eliminando imágenes antiguas...
docker rmi asam-backend-api 2>nul

echo.
echo 🏗️ Construyendo e iniciando servicios...
docker-compose up -d --build

echo.
echo ⏳ Esperando a que PostgreSQL esté listo...
timeout /t 15 /nobreak > nul

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
echo    - Ver logs: docker-compose logs -f
echo    - Ver logs del API: docker-compose logs -f api
echo    - Ver logs de PostgreSQL: docker-compose logs -f postgres
echo    - Detener: docker-compose down
echo    - Reiniciar: docker-compose restart
echo.
pause

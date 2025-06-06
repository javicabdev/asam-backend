@echo off
echo 🔄 Reiniciando solo el contenedor del API...
echo.

echo 📊 Aplicando configuración actualizada...
docker-compose restart api

echo.
echo ⏳ Esperando a que el API esté listo...
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
echo ✅ ¡API reiniciado!
echo.
echo 📋 Ver logs del API:
docker-compose logs -f api

@echo off
echo 📊 Ejecutando migraciones directamente en el contenedor...
echo.

echo 🔍 Entrando al contenedor para ejecutar migraciones...
docker-compose exec api sh -c "cd /app && ./asam-backend migrate up"
if %errorlevel% neq 0 (
    echo ⚠️  Intentando método alternativo...
    docker-compose exec postgres psql -U postgres -d asam_db -f /docker-entrypoint-initdb.d/init.sql 2>nul
    if %errorlevel% neq 0 (
        echo ❌ No se pudieron ejecutar las migraciones automáticamente
        echo.
        echo 💡 Ejecuta las migraciones manualmente:
        echo    1. docker-compose exec api sh
        echo    2. ./asam-backend migrate up
    )
)

echo.
echo 👤 Creando usuarios de prueba...
docker-compose exec -T postgres psql -U postgres -d asam_db < scripts/create-test-users.sql

echo.
echo ✅ Proceso completado!
pause

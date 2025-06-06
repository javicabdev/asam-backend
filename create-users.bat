@echo off
echo 👤 Creando usuarios de prueba en ASAM Backend...
echo.

REM Verificar que los contenedores estén corriendo
docker ps --filter "name=asam-postgres" --format "{{.Names}}" | findstr "asam-postgres" > nul
if errorlevel 1 (
    echo ❌ El contenedor de PostgreSQL no está corriendo.
    echo    Ejecuta 'start-docker.bat' primero.
    exit /b 1
)

echo 📝 Creando usuarios de prueba en la base de datos...
docker-compose exec -T postgres psql -U postgres -d asam_db < scripts/create-test-users.sql

echo.
echo ✅ Usuarios creados exitosamente!
echo.
echo 🔐 Credenciales de prueba:
echo    Administrador:
echo    - Email: admin@asam.org
echo    - Password: admin123
echo.
echo    Usuario regular:
echo    - Email: user@asam.org
echo    - Password: admin123
echo.
echo 🚀 Puedes probar el login en:
echo    - Frontend: http://localhost:5173
echo    - GraphQL Playground: http://localhost:8080/playground
echo.
pause

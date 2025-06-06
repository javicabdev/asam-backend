@echo off
echo 🔍 Verificando estado completo de ASAM...
echo.

echo 📦 Estado de contenedores:
docker-compose ps

echo.
echo 🔐 Variables JWT en el contenedor:
docker-compose exec api sh -c "env | grep JWT" 2>nul
if %errorlevel% neq 0 (
    echo ❌ No se puede acceder al contenedor del API
) else (
    echo ✅ Variables JWT configuradas
)

echo.
echo 🌐 Verificando GraphQL Playground...
curl -s -o nul -w "HTTP Status: %%{http_code}\n" http://localhost:8080/playground
if %errorlevel% equ 0 (
    echo ✅ GraphQL Playground accesible
) else (
    echo ❌ GraphQL Playground NO accesible
)

echo.
echo 📊 Verificando base de datos...
docker-compose exec postgres psql -U postgres -d asam_db -c "SELECT COUNT(*) FROM users;" 2>nul
if %errorlevel% neq 0 (
    echo ⚠️  Tabla users no existe o no se puede acceder
)

echo.
echo 📋 Resumen:
echo    - API: http://localhost:8080
echo    - GraphQL: http://localhost:8080/playground
echo    - Frontend: http://localhost:5173
echo.
echo 🔐 Credenciales:
echo    - admin@asam.org / admin123
echo.
pause

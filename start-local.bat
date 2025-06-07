@echo off
setlocal enabledelayedexpansion

echo.
echo =========================================
echo        ASAM Backend - Arranque Local     
echo =========================================
echo.

:: Verificar Docker
echo [1/7] Verificando Docker...
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker no esta instalado o no esta funcionando
    echo         Por favor instala Docker Desktop desde: https://www.docker.com/products/docker-desktop
    exit /b 1
)
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker Compose no esta instalado
    exit /b 1
)
echo [OK] Docker esta instalado y funcionando
echo.

:: Verificar Go (opcional)
echo [2/7] Verificando Go (opcional)...
go version >nul 2>&1
if errorlevel 1 (
    echo [AVISO] Go no esta instalado (opcional para solo ejecutar con Docker)
) else (
    for /f "tokens=*" %%i in ('go version') do echo [OK] %%i
)
echo.

:: Configurar archivo de entorno
echo [3/7] Configurando archivo de entorno...
if not exist ".env" (
    if exist ".env.development.example" (
        copy ".env.development.example" ".env" >nul
        echo [OK] Archivo .env creado
    ) else (
        echo [ERROR] No se encontro .env.development.example
        exit /b 1
    )
) else (
    echo [OK] Archivo .env ya existe
)
echo.

:: Detener contenedores previos
echo [4/7] Deteniendo contenedores previos...
docker-compose down >nul 2>&1
echo [OK] Contenedores detenidos
echo.

:: Verificar si se solicita limpieza
if "%1"=="--clean" (
    echo [4.5/7] Limpiando volumenes de datos...
    docker-compose down -v >nul 2>&1
    echo [OK] Volumenes eliminados
    echo.
)

:: Construir y arrancar servicios
echo [5/7] Construyendo y arrancando servicios...
docker-compose up -d --build
if errorlevel 1 (
    echo [ERROR] Fallo al arrancar servicios
    exit /b 1
)
echo [OK] Servicios arrancados
echo.

:: Esperar a que PostgreSQL este listo
echo [6/7] Esperando a que PostgreSQL este listo...
set /a attempts=0
set /a max_attempts=30

:wait_postgres
set /a attempts+=1
if !attempts! gtr !max_attempts! (
    echo [ERROR] PostgreSQL no esta respondiendo despues de 30 segundos
    exit /b 1
)

docker-compose exec -T postgres pg_isready -U postgres -d asam_db >nul 2>&1
if errorlevel 1 (
    <nul set /p =.
    timeout /t 1 /nobreak >nul
    goto wait_postgres
)
echo.
echo [OK] PostgreSQL esta listo
echo.

:: Ejecutar migraciones
echo [7/7] Ejecutando migraciones y creando usuarios...
docker-compose exec -T api go run ./cmd/migrate up >nul 2>&1
if errorlevel 1 (
    echo [AVISO] Error al ejecutar migraciones
) else (
    echo [OK] Migraciones ejecutadas
)

:: Crear usuarios de prueba
type scripts\create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db >nul 2>&1
if errorlevel 1 (
    echo [AVISO] Error al crear usuarios (puede que ya existan)
) else (
    echo [OK] Usuarios de prueba creados
)
echo.

:: Mostrar información de acceso
echo ============================================================
echo                     ASAM Backend Activo
echo ============================================================
echo.
echo   GraphQL Playground: http://localhost:8080/playground
echo   API Endpoint:      http://localhost:8080/graphql
echo   Health Check:      http://localhost:8080/health
echo   Metrics:           http://localhost:8080/metrics
echo.
echo   Usuarios de Prueba:
echo   - Admin:    admin@asam.org / admin123
echo   - Usuario:  user@asam.org  / admin123
echo.
echo   Para detener: docker-compose down
echo   Limpiar todo: start-local.bat --clean
echo.
echo ============================================================
echo.
echo Mostrando logs (Ctrl+C para detener)...
echo.

:: Seguir logs
docker-compose logs -f api

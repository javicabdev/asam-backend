# Script para arrancar ASAM Backend localmente
# Este script facilita el arranque del proyecto con Docker

Write-Host @"
╔═══════════════════════════════════════╗
║       ASAM Backend - Arranque Local   ║
╚═══════════════════════════════════════╝
"@ -ForegroundColor Cyan

# Verificar Docker
Write-Host "🔍 Verificando Docker..." -ForegroundColor Yellow
try {
    docker --version | Out-Null
    docker-compose --version | Out-Null
    Write-Host "✅ Docker está instalado y funcionando" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker no está instalado o no está funcionando" -ForegroundColor Red
    Write-Host "   Por favor instala Docker Desktop desde: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Verificar Go (opcional, solo para desarrollo)
Write-Host "`n🔍 Verificando Go..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Go no está instalado (opcional para solo ejecutar con Docker)" -ForegroundColor Yellow
}

# Detener contenedores previos
Write-Host "`n🛑 Deteniendo contenedores previos..." -ForegroundColor Yellow
docker-compose down --remove-orphans 2>$null

# Limpiar volúmenes si se especifica
if ($args -contains "--clean") {
    Write-Host "`n🧹 Limpieza completa del entorno..." -ForegroundColor Yellow
    
    # Detener y limpiar todo
    Write-Host "   Deteniendo todos los contenedores..." -ForegroundColor Gray
    docker-compose down -v --remove-orphans
    
    # Limpiar contenedores huérfanos adicionales
    Write-Host "   Eliminando contenedores huérfanos..." -ForegroundColor Gray
    docker container prune -f 2>$null
    
    # Limpiar redes no utilizadas
    Write-Host "   Limpiando redes no utilizadas..." -ForegroundColor Gray
    docker network prune -f 2>$null
    
    # Eliminar el archivo .env para empezar limpio
    if (Test-Path ".env") {
        Write-Host "   Eliminando archivo .env existente..." -ForegroundColor Gray
        Remove-Item ".env" -Force
    }
    
    Write-Host "✅ Limpieza completa finalizada" -ForegroundColor Green
    Start-Sleep -Seconds 2
}

# Siempre verificar/crear archivo de entorno (especialmente después de --clean)
Write-Host "`n📋 Configurando archivo de entorno..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    if (Test-Path ".env.development.example") {
        Copy-Item ".env.development.example" ".env"
        Write-Host "✅ Archivo .env creado desde .env.development.example" -ForegroundColor Green
    } elseif (Test-Path ".env.development") {
        Copy-Item ".env.development" ".env"
        Write-Host "✅ Archivo .env creado desde .env.development" -ForegroundColor Green
    } else {
        Write-Host "❌ No se encontró archivo de configuración de ejemplo" -ForegroundColor Red
        Write-Host "   Creando archivo .env mínimo..." -ForegroundColor Yellow
        
        # Crear un .env mínimo para desarrollo
        @"
# Database configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=asam_db

# API configuration
API_PORT=8080
ENVIRONMENT=development

# JWT configuration
JWT_ACCESS_SECRET=dev-access-secret-change-in-production
JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Admin user (for monitoring endpoints)
ADMIN_USER=admin
ADMIN_PASSWORD=admin123
"@ | Out-File -FilePath ".env" -Encoding utf8
        Write-Host "✅ Archivo .env creado con configuración mínima" -ForegroundColor Green
    }
} else {
    Write-Host "✅ Archivo .env ya existe" -ForegroundColor Green
}

# Verificar y corregir DB_HOST para Docker
Write-Host "`n🔧 Verificando configuración de base de datos..." -ForegroundColor Yellow
$envContent = Get-Content ".env" -Raw

# Si DB_HOST es localhost, cambiarlo a postgres para Docker
if ($envContent -match "DB_HOST=localhost") {
    Write-Host "   Corrigiendo DB_HOST de localhost a postgres para Docker..." -ForegroundColor Gray
    $envContent = $envContent -replace "DB_HOST=localhost", "DB_HOST=postgres"
    Set-Content ".env" -Value $envContent -NoNewline
    Write-Host "✅ DB_HOST actualizado para Docker" -ForegroundColor Green
} elseif ($envContent -match "DB_HOST=postgres") {
    Write-Host "✅ DB_HOST ya configurado correctamente para Docker" -ForegroundColor Green
} else {
    Write-Host "⚠️  DB_HOST tiene un valor personalizado, verificar configuración" -ForegroundColor Yellow
}

# Verificar y agregar variables JWT si no existen
Write-Host "`n🔐 Verificando configuración de JWT..." -ForegroundColor Yellow
$envContent = Get-Content ".env" -Raw

$jwtConfigAdded = $false
if ($envContent -notmatch "JWT_ACCESS_SECRET") {
    Write-Host "   Agregando JWT_ACCESS_SECRET..." -ForegroundColor Gray
    Add-Content ".env" "`n# JWT Configuration (added by start-docker.ps1)"
    Add-Content ".env" "JWT_ACCESS_SECRET=dev-access-secret-change-in-production"
    $jwtConfigAdded = $true
}

if ($envContent -notmatch "JWT_REFRESH_SECRET") {
    Write-Host "   Agregando JWT_REFRESH_SECRET..." -ForegroundColor Gray
    if (-not $jwtConfigAdded) {
        Add-Content ".env" "`n# JWT Configuration (added by start-docker.ps1)"
    }
    Add-Content ".env" "JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production"
    $jwtConfigAdded = $true
}

if ($envContent -notmatch "JWT_ACCESS_TTL") {
    Write-Host "   Agregando JWT_ACCESS_TTL..." -ForegroundColor Gray
    Add-Content ".env" "JWT_ACCESS_TTL=15m"
    $jwtConfigAdded = $true
}

if ($envContent -notmatch "JWT_REFRESH_TTL") {
    Write-Host "   Agregando JWT_REFRESH_TTL..." -ForegroundColor Gray
    Add-Content ".env" "JWT_REFRESH_TTL=168h"
    $jwtConfigAdded = $true
}

if ($jwtConfigAdded) {
    Write-Host "✅ Configuración JWT agregada al archivo .env" -ForegroundColor Green
} else {
    Write-Host "✅ Configuración JWT ya existe" -ForegroundColor Green
}

# Verificar si hay problemas con contenedores existentes
Write-Host "`n🔍 Verificando estado de contenedores..." -ForegroundColor Yellow
$existingContainers = docker ps -a --filter "name=asam" --format "{{.Names}} {{.Status}}" 2>$null
if ($existingContainers) {
    $deadContainers = $existingContainers | Where-Object { $_ -match "Exited" -or $_ -match "Dead" }
    if ($deadContainers) {
        Write-Host "   ⚠️  Detectados contenedores en mal estado" -ForegroundColor Yellow
        Write-Host "   Limpiando contenedores problemáticos..." -ForegroundColor Gray
        docker-compose down -v --remove-orphans
        Start-Sleep -Seconds 2
    }
}

# Construir y arrancar servicios
Write-Host "`n🚀 Construyendo y arrancando servicios..." -ForegroundColor Yellow
docker-compose up -d --build

# Verificar si los contenedores arrancaron correctamente
Start-Sleep -Seconds 3
$apiStatus = docker ps --filter "name=asam-backend-api" --format "{{.Status}}" 2>$null
$dbStatus = docker ps --filter "name=asam-postgres" --format "{{.Status}}" 2>$null

if (-not $apiStatus -or -not $dbStatus) {
    Write-Host "❌ Error: Los contenedores no arrancaron correctamente" -ForegroundColor Red
    Write-Host "   Intenta ejecutar: .\scripts\reset-emergency.ps1" -ForegroundColor Yellow
    Write-Host "   Y luego: .\start-docker.ps1" -ForegroundColor Yellow
    exit 1
}

# Si agregamos configuración JWT, reiniciar el contenedor API para cargar los cambios
if ($jwtConfigAdded) {
    Write-Host "`n🔄 Reiniciando API para aplicar cambios de configuración..." -ForegroundColor Yellow
    docker-compose restart api
    Start-Sleep -Seconds 3
}

# Esperar a que PostgreSQL esté listo
Write-Host "`n⏳ Esperando a que PostgreSQL esté listo..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    $attempt++
    Write-Host -NoNewline "."
    
    try {
        $result = docker-compose exec -T postgres pg_isready -U postgres -d asam_db 2>$null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
        }
    } catch {
        # Ignorar errores, seguir esperando
    }
    
    if (-not $ready) {
        Start-Sleep -Seconds 1
    }
}

Write-Host ""
if ($ready) {
    Write-Host "✅ PostgreSQL está listo" -ForegroundColor Green
} else {
    Write-Host "❌ PostgreSQL no está respondiendo" -ForegroundColor Red
    exit 1
}

# Ejecutar migraciones
Write-Host "`n🔄 Ejecutando migraciones..." -ForegroundColor Yellow
# Esperar un poco más para asegurar que el API esté lista
Start-Sleep -Seconds 3

# Primero copiar .env a .env.development para que el comando de migración lo encuentre
docker-compose exec -T api sh -c "cp .env .env.development" 2>$null

# Siempre ejecutar migraciones para asegurar que todas estén aplicadas
Write-Host "   Verificando y aplicando todas las migraciones..." -ForegroundColor Gray

# Intentar ejecutar migraciones con el comando Go
Write-Host "   Ejecutando migraciones..." -ForegroundColor Gray
docker-compose exec -T api go run ./cmd/migrate -env local up

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Migraciones ejecutadas con éxito" -ForegroundColor Green
} else {
    Write-Host "⚠️  Error al ejecutar migraciones con Go" -ForegroundColor Yellow
    
    # Como respaldo, intentar ejecutar las migraciones SQL directamente
    Write-Host "   Intentando ejecutar migraciones SQL directamente..." -ForegroundColor Gray
    
    # Obtener todos los archivos de migración .up.sql ordenados
    $migrationFiles = Get-ChildItem -Path "migrations" -Filter "*.up.sql" | Sort-Object Name
    
    foreach ($migration in $migrationFiles) {
        Write-Host "   Aplicando: $($migration.Name)" -ForegroundColor Gray
        
        # Verificar si la migración ya fue aplicada (verificación simple basada en existencia de columnas/tablas)
        $migrationContent = Get-Content $migration.FullName -Raw
        
        # Ejecutar la migración
        $migrationContent | docker-compose exec -T postgres psql -U postgres -d asam_db 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   ✓ $($migration.Name) aplicada" -ForegroundColor DarkGreen
        } else {
            # Ignorar errores de "already exists" ya que es esperado
            Write-Host "   ~ $($migration.Name) (ya aplicada o error menor)" -ForegroundColor DarkGray
        }
    }
    
    Write-Host "✅ Proceso de migraciones completado" -ForegroundColor Green
}

# Verificar que las tablas principales existen
Start-Sleep -Seconds 2
$verifyTables = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'members', 'families', 'payments', 'cash_flows');" 2>$null
if ($verifyTables -is [array]) {
    $verifyTables = $verifyTables -join ''
}
$verifyTables = $verifyTables -replace '\s', ''
try {
    $verifyCount = [int]$verifyTables
    if ($verifyCount -eq 5) {
        Write-Host "✅ Verificado: Todas las tablas principales existen" -ForegroundColor Green
    } elseif ($verifyCount -gt 0) {
        Write-Host "⚠️  Solo $verifyCount de 5 tablas principales fueron creadas" -ForegroundColor Yellow
    } else {
        Write-Host "❌ Error: No se crearon las tablas principales" -ForegroundColor Red
        Write-Host "   Intenta ejecutar: .\scripts\reset-emergency.ps1" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "⚠️  No se pudo verificar la creación de tablas" -ForegroundColor Yellow
}

# Crear usuarios de prueba usando la herramienta de gestión de usuarios
Write-Host "`n👥 Creando usuarios de prueba..." -ForegroundColor Yellow

# Verificar si ya existen usuarios
$userCount = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users;" 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "   La tabla users no existe, necesita ejecutar migraciones primero" -ForegroundColor Yellow
    $userCountInt = 0
} else {
    if ($userCount -is [array]) {
        $userCount = $userCount -join ''
    }
    $userCount = $userCount -replace '\s', ''
    try {
        $userCountInt = [int]$userCount
    } catch {
        $userCountInt = 0
    }
}

if ($userCountInt -eq 0) {
    Write-Host "   No hay usuarios, creando usuarios de prueba..." -ForegroundColor Gray
    # Esperar un poco para asegurar que el API esté completamente lista
    Start-Sleep -Seconds 2
    
    # Usar el script automatizado que no requiere interacción
    docker-compose exec -T api go run scripts/user-management/auto-create-test-users/auto-create-test-users.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Usuarios de prueba creados correctamente" -ForegroundColor Green
        
        # Verificar que los usuarios se crearon
        $newUserCount = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users;" 2>$null
        if ($newUserCount -is [array]) {
            $newUserCount = $newUserCount -join ''
        }
        $newUserCount = $newUserCount -replace '\s', ''
        try {
            $newUserCountInt = [int]$newUserCount
            Write-Host "   Total de usuarios en la base de datos: $newUserCountInt" -ForegroundColor Gray
        } catch {
            Write-Host "   No se pudo verificar el número de usuarios creados" -ForegroundColor Yellow
        }
    } else {
        Write-Host "⚠️  Error al crear usuarios con el script" -ForegroundColor Yellow
        Write-Host "   Intenta ejecutar manualmente: make db-seed" -ForegroundColor Yellow
    }
} else {
    Write-Host "✅ Ya existen $userCountInt usuarios en la base de datos" -ForegroundColor Green
    
    # Mostrar los usuarios existentes
    Write-Host "   Usuarios existentes:" -ForegroundColor Gray
    docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT username, role FROM users;" | ForEach-Object {
        if ($_ -match '\S') {
            Write-Host "   - $_" -ForegroundColor DarkGray
        }
    }
}

# Verificación final antes de mostrar logs
$finalUserCheck = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users WHERE username IN ('admin', 'user');" 2>$null
if ($LASTEXITCODE -ne 0) {
    $finalUserCount = 0
} else {
    if ($finalUserCheck -is [array]) {
        $finalUserCheck = $finalUserCheck -join ''
    }
    $finalUserCheck = $finalUserCheck -replace '\s', ''
    try {
        $finalUserCount = [int]$finalUserCheck
    } catch {
        $finalUserCount = 0
    }
}

if ($finalUserCount -lt 2) {
    Write-Host "`n⚠️  ADVERTENCIA: Los usuarios de prueba no se crearon correctamente" -ForegroundColor Yellow
    Write-Host "   Solución rápida: .\scripts\auto-fix.ps1" -ForegroundColor Cyan
    Write-Host "   O manualmente: docker-compose exec api go run scripts/user-management/auto-create-test-users.go" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "   Para diagnóstico completo: .\scripts\diagnostico.ps1" -ForegroundColor Gray
}

# Mostrar logs en tiempo real
Write-Host "`n📜 Mostrando logs de la aplicación..." -ForegroundColor Yellow
Write-Host "   (Presiona Ctrl+C para detener los logs)" -ForegroundColor Gray
Write-Host ""

# Mostrar información de acceso
Write-Host @"

╔════════════════════════════════════════════════════════════╗
║                    ASAM Backend Activo                     ║
╠════════════════════════════════════════════════════════════╣
║  🌐 GraphQL Playground: http://localhost:8080/playground   ║
║  🔧 API Endpoint:      http://localhost:8080/graphql      ║
║  ❤️  Health Check:     http://localhost:8080/health       ║
║  📊 Metrics:          http://localhost:8080/metrics       ║
╠════════════════════════════════════════════════════════════╣
║                  Usuarios de Prueba:                       ║
║  👤 Admin:     admin / AsamAdmin2025!                     ║
║  👤 Usuario:   user  / AsamUser2025!                      ║
╠════════════════════════════════════════════════════════════╣
║  🛑 Para detener: docker-compose down                      ║
║  🧹 Limpiar todo: .\start-docker.ps1 --clean             ║
╠════════════════════════════════════════════════════════════╣
║  🔧 ¿Problemas? Ejecuta: .\scripts\auto-fix.ps1           ║
║  📊 Diagnóstico: .\scripts\diagnostico.ps1                ║
║  ❓ Ver ayuda: .\scripts\help.ps1                          ║
╚════════════════════════════════════════════════════════════╝

"@ -ForegroundColor Cyan

# Seguir logs
docker-compose logs -f api

# Script de corrección automática para ASAM Backend

Write-Host @"
╔═══════════════════════════════════════╗
║   ASAM Backend - Auto Fix             ║
╚═══════════════════════════════════════╝
"@ -ForegroundColor Cyan

Write-Host "`n🔧 Este script intentará corregir automáticamente los problemas comunes" -ForegroundColor Yellow
Write-Host "   Presiona Ctrl+C para cancelar en cualquier momento" -ForegroundColor Gray
Write-Host ""

# Paso 1: Verificar Docker
Write-Host "🔍 Verificando Docker..." -ForegroundColor Yellow
try {
    docker --version | Out-Null
    Write-Host "✅ Docker está instalado" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker no está instalado" -ForegroundColor Red
    Write-Host "   Por favor instala Docker Desktop desde: https://docker.com" -ForegroundColor Yellow
    exit 1
}

# Paso 2: Verificar y corregir archivo .env
Write-Host "`n🔧 Verificando archivo .env..." -ForegroundColor Yellow
$envFixed = $false

if (-not (Test-Path ".env")) {
    Write-Host "   Creando archivo .env..." -ForegroundColor Gray
    if (Test-Path ".env.development.example") {
        Copy-Item ".env.development.example" ".env"
        $envFixed = $true
    } else {
        Write-Host "❌ No se encontró .env.development.example" -ForegroundColor Red
        exit 1
    }
}

# Leer contenido actual
$envContent = Get-Content ".env" -Raw

# Corregir DB_HOST
if ($envContent -match "DB_HOST=localhost") {
    Write-Host "   Corrigiendo DB_HOST..." -ForegroundColor Gray
    $envContent = $envContent -replace "DB_HOST=localhost", "DB_HOST=postgres"
    $envFixed = $true
}

# Agregar JWT si falta
$jwtAdded = $false
if ($envContent -notmatch "JWT_ACCESS_SECRET") {
    Write-Host "   Agregando JWT_ACCESS_SECRET..." -ForegroundColor Gray
    $envContent += "`n# JWT Configuration (added by auto-fix.ps1)`nJWT_ACCESS_SECRET=dev-access-secret-change-in-production"
    $jwtAdded = $true
    $envFixed = $true
}
if ($envContent -notmatch "JWT_REFRESH_SECRET") {
    Write-Host "   Agregando JWT_REFRESH_SECRET..." -ForegroundColor Gray
    if (-not $jwtAdded) {
        $envContent += "`n# JWT Configuration (added by auto-fix.ps1)"
    }
    $envContent += "`nJWT_REFRESH_SECRET=dev-refresh-secret-change-in-production"
    $envFixed = $true
}
if ($envContent -notmatch "JWT_ACCESS_TTL") {
    Write-Host "   Agregando JWT_ACCESS_TTL..." -ForegroundColor Gray
    $envContent += "`nJWT_ACCESS_TTL=15m"
    $envFixed = $true
}
if ($envContent -notmatch "JWT_REFRESH_TTL") {
    Write-Host "   Agregando JWT_REFRESH_TTL..." -ForegroundColor Gray
    $envContent += "`nJWT_REFRESH_TTL=168h"
    $envFixed = $true
}

if ($envFixed) {
    Set-Content ".env" -Value $envContent -NoNewline
    Write-Host "✅ Archivo .env corregido" -ForegroundColor Green
} else {
    Write-Host "✅ Archivo .env ya está correcto" -ForegroundColor Green
}

# Paso 3: Verificar contenedores
Write-Host "`n🐳 Verificando contenedores..." -ForegroundColor Yellow
$apiRunning = docker ps --filter "name=asam-backend-api" --format "{{.Names}}" | Select-String "asam-backend-api"
$dbRunning = docker ps --filter "name=asam-postgres" --format "{{.Names}}" | Select-String "asam-postgres"

if (-not $apiRunning -or -not $dbRunning) {
    Write-Host "   Iniciando contenedores..." -ForegroundColor Gray
    docker-compose up -d
    
    # Esperar a que arranquen
    Write-Host "   Esperando a que los servicios estén listos..." -ForegroundColor Gray
    Start-Sleep -Seconds 10
    
    # Verificar nuevamente
    $apiRunning = docker ps --filter "name=asam-backend-api" --format "{{.Names}}" | Select-String "asam-backend-api"
    $dbRunning = docker ps --filter "name=asam-postgres" --format "{{.Names}}" | Select-String "asam-postgres"
    
    if ($apiRunning -and $dbRunning) {
        Write-Host "✅ Contenedores iniciados" -ForegroundColor Green
    } else {
        Write-Host "❌ No se pudieron iniciar los contenedores" -ForegroundColor Red
        Write-Host "   Intenta: docker-compose down -v && docker-compose up -d" -ForegroundColor Yellow
        exit 1
    }
} else {
    Write-Host "✅ Contenedores ya están corriendo" -ForegroundColor Green
    
    # Si modificamos el .env, reiniciar API
    if ($envFixed) {
        Write-Host "   Reiniciando API para aplicar cambios..." -ForegroundColor Gray
        docker-compose restart api
        Start-Sleep -Seconds 5
    }
}

# Paso 4: Verificar y ejecutar migraciones
Write-Host "`n📊 Verificando base de datos..." -ForegroundColor Yellow
$tableCheck = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT 1 FROM information_schema.tables WHERE table_name = 'users';" 2>$null

if ($LASTEXITCODE -ne 0 -or $tableCheck -notmatch "1") {
    Write-Host "   Las tablas no existen, ejecutando migraciones..." -ForegroundColor Gray
    
    # Copiar .env para el comando de migración
    docker-compose exec -T api sh -c "cp .env .env.development" 2>$null
    
    # Intentar con Go primero
    docker-compose exec -T api go run ./cmd/migrate -env local up 2>$null
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "   Ejecutando migraciones con SQL directo..." -ForegroundColor Gray
        
        if (Test-Path "migrations/000001_initial_schema.up.sql") {
            Get-Content "migrations/000001_initial_schema.up.sql" -Raw | docker-compose exec -T postgres psql -U postgres -d asam_db
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Migraciones ejecutadas" -ForegroundColor Green
            } else {
                Write-Host "❌ Error al ejecutar migraciones" -ForegroundColor Red
                Write-Host "   Intenta: .\scripts\manual-setup.ps1" -ForegroundColor Yellow
                exit 1
            }
        }
    } else {
        Write-Host "✅ Migraciones ejecutadas" -ForegroundColor Green
    }
} else {
    Write-Host "✅ Las tablas ya existen" -ForegroundColor Green
}

# Paso 5: Verificar y crear usuarios
Write-Host "`n👥 Verificando usuarios..." -ForegroundColor Yellow
$userCheck = docker-compose exec -T postgres psql -U postgres -d asam_db -t -c "SELECT COUNT(*) FROM users WHERE username IN ('admin@asam.org', 'user@asam.org');" 2>$null

if ($userCheck -is [array]) {
    $userCheck = $userCheck -join ''
}
$userCheck = $userCheck -replace '\s', ''

try {
    $userCount = [int]$userCheck
} catch {
    $userCount = 0
}

if ($userCount -lt 2) {
    Write-Host "   Creando usuarios de prueba..." -ForegroundColor Gray
    docker-compose exec -T api go run scripts/user-management/auto-create-test-users.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Usuarios creados" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Posible error al crear usuarios" -ForegroundColor Yellow
    }
} else {
    Write-Host "✅ Usuarios de prueba ya existen" -ForegroundColor Green
}

# Paso 6: Verificación final
Write-Host "`n🏁 Verificación final..." -ForegroundColor Yellow

# Test de salud
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -ErrorAction Stop
    Write-Host "✅ API respondiendo correctamente" -ForegroundColor Green
} catch {
    Write-Host "⚠️  API no responde en /health (puede estar iniciándose)" -ForegroundColor Yellow
}

# Resumen
Write-Host "`n" -NoNewline
Write-Host @"
╔════════════════════════════════════════════════════════════╗
║                    ✅ AUTO-FIX COMPLETADO                  ║
╠════════════════════════════════════════════════════════════╣
║  El sistema debería estar funcionando correctamente ahora  ║
║                                                            ║
║  🌐 GraphQL Playground: http://localhost:8080/playground   ║
║  👤 Admin:     admin@asam.org / admin123                  ║
║  👤 Usuario:   user@asam.org  / admin123                  ║
╠════════════════════════════════════════════════════════════╣
║  Si sigues teniendo problemas:                             ║
║  1. Ejecuta: .\scripts\diagnostico.ps1                    ║
║  2. O reinicia todo: .\start-docker.ps1 --clean          ║
╚════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan

Write-Host "`n📜 Mostrando logs (Ctrl+C para salir)..." -ForegroundColor Yellow
docker-compose logs -f api

# Script para verificar el estado de Docker y los servicios
Write-Host "🔍 Verificando estado de ASAM Backend con Docker..." -ForegroundColor Green
Write-Host ""

# Verificar si Docker está corriendo
Write-Host "🐳 Verificando Docker..." -ForegroundColor Yellow
try {
    docker version | Out-Null
    Write-Host "✅ Docker está corriendo" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker no está corriendo. Por favor inicia Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "📦 Verificando contenedores..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "🔍 Verificando servicios específicos..." -ForegroundColor Yellow

# Verificar PostgreSQL
$postgresRunning = docker-compose ps postgres | Select-String "Up"
if ($postgresRunning) {
    Write-Host "✅ PostgreSQL está corriendo" -ForegroundColor Green
    
    # Verificar conexión a la base de datos
    Write-Host "   Verificando conexión a la base de datos..." -ForegroundColor Gray
    docker-compose exec -T postgres pg_isready -U postgres -d asam_db
    
    # Verificar si las tablas existen
    Write-Host "   Verificando tablas..." -ForegroundColor Gray
    $tables = docker-compose exec -T postgres psql -U postgres -d asam_db -c "\dt" 2>$null
    if ($tables -match "users") {
        Write-Host "   ✅ Tabla 'users' existe" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  Tabla 'users' no existe - ejecuta las migraciones" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ PostgreSQL NO está corriendo" -ForegroundColor Red
}

# Verificar API
$apiRunning = docker-compose ps api | Select-String "Up"
if ($apiRunning) {
    Write-Host "✅ API está corriendo" -ForegroundColor Green
    
    # Verificar endpoint
    Write-Host "   Verificando endpoint..." -ForegroundColor Gray
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/playground" -Method HEAD -TimeoutSec 2 -ErrorAction Stop
        Write-Host "   ✅ GraphQL Playground accesible" -ForegroundColor Green
    } catch {
        Write-Host "   ⚠️  No se puede acceder al GraphQL Playground" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ API NO está corriendo" -ForegroundColor Red
}

Write-Host ""
Write-Host "📊 Resumen:" -ForegroundColor Cyan
Write-Host "   - PostgreSQL: " -NoNewline
if ($postgresRunning) { 
    Write-Host "✅ Corriendo" -ForegroundColor Green 
} else { 
    Write-Host "❌ No corriendo" -ForegroundColor Red 
}

Write-Host "   - API: " -NoNewline
if ($apiRunning) { 
    Write-Host "✅ Corriendo" -ForegroundColor Green 
} else { 
    Write-Host "❌ No corriendo" -ForegroundColor Red 
}

Write-Host ""
Write-Host "💡 Si algo no está corriendo, ejecuta:" -ForegroundColor Yellow
Write-Host "   .\clean-start-docker.ps1" -ForegroundColor White
Write-Host ""

# Mostrar logs recientes si hay errores
if (-not $apiRunning -or -not $postgresRunning) {
    Write-Host "📃 Últimos logs de error:" -ForegroundColor Red
    docker-compose logs --tail=20
}

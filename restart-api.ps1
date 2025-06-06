# Script para reiniciar solo el contenedor del API
Write-Host "🔄 Reiniciando solo el contenedor del API..." -ForegroundColor Green
Write-Host ""

# Detener el contenedor del API
Write-Host "🛑 Deteniendo el API..." -ForegroundColor Yellow
docker-compose stop api

Write-Host ""
Write-Host "📊 Aplicando configuración actualizada..." -ForegroundColor Yellow
docker-compose up -d api

Write-Host ""
Write-Host "⏳ Esperando a que el API esté listo..." -ForegroundColor Yellow
$retries = 30
$ready = $false

while ($retries -gt 0 -and -not $ready) {
    Start-Sleep -Seconds 2
    $status = docker-compose ps api --format json | ConvertFrom-Json
    if ($status.State -eq "running") {
        $ready = $true
        Write-Host "✅ API está corriendo!" -ForegroundColor Green
    } else {
        Write-Host "." -NoNewline
    }
    $retries--
}

if (-not $ready) {
    Write-Host ""
    Write-Host "❌ El API no pudo iniciar correctamente" -ForegroundColor Red
    Write-Host "Verificando logs..." -ForegroundColor Yellow
    docker-compose logs --tail=50 api
    exit 1
}

Write-Host ""
Write-Host ""

# Verificar si el API responde
Write-Host "🔍 Verificando endpoint del API..." -ForegroundColor Yellow
Start-Sleep -Seconds 5
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/playground" -Method HEAD -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ GraphQL Playground accesible!" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "📊 Ejecutando migraciones..." -ForegroundColor Yellow
    docker-compose exec -T api go run cmd/migrate/main.go -env=local -cmd=up
    
    Write-Host ""
    Write-Host "🌱 Ejecutando seed con datos de prueba..." -ForegroundColor Yellow
    docker-compose exec -T api go run cmd/seed/main.go -env=local -type=minimal
    
    Write-Host ""
    Write-Host "👤 Creando usuarios de prueba..." -ForegroundColor Yellow
    Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db
    
} catch {
    Write-Host "⚠️  No se puede acceder al API" -ForegroundColor Yellow
    Write-Host "Verificando logs..." -ForegroundColor Yellow
    docker-compose logs --tail=50 api
}

Write-Host ""
Write-Host "✅ Proceso completado!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Para ver logs en tiempo real:" -ForegroundColor Cyan
Write-Host "   docker-compose logs -f api" -ForegroundColor White

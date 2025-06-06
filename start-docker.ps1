# Script PowerShell para iniciar ASAM Backend con Docker
Write-Host "🐳 Iniciando ASAM Backend con Docker..." -ForegroundColor Green
Write-Host ""

# Detener contenedores existentes
Write-Host "🛑 Deteniendo contenedores existentes..." -ForegroundColor Yellow
docker-compose down

Write-Host ""
Write-Host "🏗️  Construyendo e iniciando servicios..." -ForegroundColor Yellow
docker-compose up -d --build

Write-Host ""
Write-Host "⏳ Esperando a que PostgreSQL esté listo..." -ForegroundColor Yellow
$retries = 30
$connected = $false

while ($retries -gt 0 -and -not $connected) {
    Start-Sleep -Seconds 2
    try {
        docker-compose exec -T postgres pg_isready -U postgres -d asam_db | Out-Null
        if ($LASTEXITCODE -eq 0) {
            $connected = $true
            Write-Host "✅ PostgreSQL está listo!" -ForegroundColor Green
        }
    }
    catch {
        # Ignorar errores
    }
    $retries--
}

if (-not $connected) {
    Write-Host "❌ PostgreSQL no pudo iniciar correctamente" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "📊 Ejecutando migraciones..." -ForegroundColor Yellow
docker-compose exec -T api go run cmd/migrate/main.go -env=local -cmd=up

Write-Host ""
Write-Host "🌱 Ejecutando seed con datos de prueba..." -ForegroundColor Yellow
docker-compose exec -T api go run cmd/seed/main.go -env=local -type=minimal

Write-Host ""
Write-Host "👤 Creando usuarios de prueba..." -ForegroundColor Yellow
Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db

Write-Host ""
Write-Host "✅ ¡ASAM Backend está corriendo!" -ForegroundColor Green
Write-Host ""
Write-Host "📍 URLs disponibles:" -ForegroundColor Cyan
Write-Host "   - API: http://localhost:8080" -ForegroundColor White
Write-Host "   - GraphQL Playground: http://localhost:8080/playground" -ForegroundColor White
Write-Host "   - PostgreSQL: localhost:5432" -ForegroundColor White
Write-Host ""
Write-Host "🔐 Credenciales de prueba:" -ForegroundColor Cyan
Write-Host "   - Email: admin@asam.org" -ForegroundColor White
Write-Host "   - Password: admin123" -ForegroundColor White
Write-Host ""
Write-Host "📋 Comandos útiles:" -ForegroundColor Cyan
Write-Host "   - Ver logs: docker-compose logs -f api" -ForegroundColor White
Write-Host "   - Detener: docker-compose down" -ForegroundColor White
Write-Host "   - Reiniciar: docker-compose restart" -ForegroundColor White
Write-Host ""

# Abrir el navegador automáticamente
$openBrowser = Read-Host "¿Deseas abrir el GraphQL Playground en el navegador? (s/n)"
if ($openBrowser -eq "s") {
    Start-Process "http://localhost:8080/playground"
}

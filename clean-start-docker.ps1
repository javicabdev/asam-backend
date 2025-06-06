# Script PowerShell para limpiar y reiniciar ASAM Backend con Docker
Write-Host "🧹 Limpiando y reiniciando ASAM Backend con Docker..." -ForegroundColor Green
Write-Host ""

# Detener contenedores existentes y eliminar volúmenes
Write-Host "🛑 Deteniendo todos los contenedores y eliminando volúmenes..." -ForegroundColor Yellow
docker-compose down -v

Write-Host ""
Write-Host "🗑️  Eliminando imágenes antiguas..." -ForegroundColor Yellow
docker rmi asam-backend-api 2>$null

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
    Write-Host "." -NoNewline
}

if (-not $connected) {
    Write-Host ""
    Write-Host "❌ PostgreSQL no pudo iniciar correctamente" -ForegroundColor Red
    Write-Host "Verificando logs..." -ForegroundColor Yellow
    docker-compose logs postgres
    exit 1
}

Write-Host ""
Write-Host ""
Write-Host "📊 Ejecutando migraciones..." -ForegroundColor Yellow
docker-compose exec -T api go run cmd/migrate/main.go -env=local -cmd=up
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  Las migraciones fallaron, pero continuando..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🌱 Ejecutando seed con datos de prueba..." -ForegroundColor Yellow
docker-compose exec -T api go run cmd/seed/main.go -env=local -type=minimal
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  El seed falló, pero continuando..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "👤 Creando usuarios de prueba..." -ForegroundColor Yellow
Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  La creación de usuarios falló, pero continuando..." -ForegroundColor Yellow
}

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
Write-Host "   - Ver logs: docker-compose logs -f" -ForegroundColor White
Write-Host "   - Ver logs del API: docker-compose logs -f api" -ForegroundColor White
Write-Host "   - Ver logs de PostgreSQL: docker-compose logs -f postgres" -ForegroundColor White
Write-Host "   - Detener: docker-compose down" -ForegroundColor White
Write-Host "   - Reiniciar: docker-compose restart" -ForegroundColor White
Write-Host ""

# Mostrar los logs del API
Write-Host "📃 Mostrando logs del API (Ctrl+C para salir)..." -ForegroundColor Yellow
docker-compose logs -f api

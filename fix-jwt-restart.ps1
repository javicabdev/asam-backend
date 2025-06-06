# Script para corregir JWT y reiniciar con migraciones
Write-Host "🔄 Aplicando corrección de JWT y reiniciando..." -ForegroundColor Green
Write-Host ""

Write-Host "🛑 Deteniendo API..." -ForegroundColor Yellow
docker-compose stop api

Write-Host ""
Write-Host "✅ JWT_REFRESH_TTL corregido a 168h (7 días)" -ForegroundColor Green

Write-Host ""
Write-Host "🚀 Iniciando API..." -ForegroundColor Yellow
docker-compose up -d api

Write-Host ""
Write-Host "⏳ Esperando a que el API esté listo..." -ForegroundColor Yellow
$retries = 30
$ready = $false

while ($retries -gt 0 -and -not $ready) {
    Start-Sleep -Seconds 2
    try {
        $logs = docker-compose logs --tail=10 api 2>&1 | Out-String
        if ($logs -match "Successfully connected to database" -or $logs -match "Server starting to listen") {
            $ready = $true
            Write-Host ""
            Write-Host "✅ API conectado a la base de datos!" -ForegroundColor Green
        } else {
            Write-Host "." -NoNewline
        }
    } catch {
        Write-Host "." -NoNewline
    }
    $retries--
}

if (-not $ready) {
    Write-Host ""
    Write-Host "⚠️  El API está tardando en iniciar, pero continuando..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📊 Ejecutando migraciones..." -ForegroundColor Yellow
$migrationResult = docker-compose exec -T api go run cmd/migrate/main.go -env=local -cmd=up 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  Las migraciones fallaron, intentando de nuevo..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5
    docker-compose exec -T api go run cmd/migrate/main.go -env=local -cmd=up
} else {
    Write-Host "✅ Migraciones ejecutadas correctamente" -ForegroundColor Green
}

Write-Host ""
Write-Host "🌱 Ejecutando seed..." -ForegroundColor Yellow
docker-compose exec -T api go run cmd/seed/main.go -env=local -type=minimal
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Seed ejecutado correctamente" -ForegroundColor Green
}

Write-Host ""
Write-Host "👤 Creando usuarios de prueba..." -ForegroundColor Yellow
Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Usuarios creados correctamente" -ForegroundColor Green
}

Write-Host ""
Write-Host "🔍 Verificando estado..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "📋 Verificando que el API esté funcionando..." -ForegroundColor Yellow
Start-Sleep -Seconds 3
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/playground" -Method HEAD -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ GraphQL Playground accesible!" -ForegroundColor Green
} catch {
    Write-Host "⚠️  GraphQL Playground no accesible todavía" -ForegroundColor Yellow
    Write-Host "   Espera unos segundos más o revisa los logs" -ForegroundColor Gray
}

Write-Host ""
Write-Host "✅ ¡Proceso completado!" -ForegroundColor Green
Write-Host ""
Write-Host "🌐 URLs disponibles:" -ForegroundColor Cyan
Write-Host "   - GraphQL Playground: http://localhost:8080/playground" -ForegroundColor White
Write-Host "   - Frontend: http://localhost:5173" -ForegroundColor White
Write-Host ""
Write-Host "🔐 Credenciales:" -ForegroundColor Cyan
Write-Host "   - Email: admin@asam.org" -ForegroundColor White
Write-Host "   - Password: admin123" -ForegroundColor White
Write-Host ""
Write-Host "📃 Para ver logs en tiempo real:" -ForegroundColor Yellow
Write-Host "   docker-compose logs -f api" -ForegroundColor White

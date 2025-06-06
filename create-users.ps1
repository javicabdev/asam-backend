# Script para crear usuarios de prueba en ASAM Backend
Write-Host "👤 Creando usuarios de prueba en ASAM Backend..." -ForegroundColor Green
Write-Host ""

# Verificar que Docker esté corriendo
try {
    docker ps | Out-Null
} catch {
    Write-Host "❌ Docker no está corriendo. Por favor inicia Docker Desktop." -ForegroundColor Red
    exit 1
}

# Verificar que los contenedores estén corriendo
$postgresRunning = docker ps --filter "name=asam-postgres" --format "{{.Names}}" | Select-String "asam-postgres"
if (-not $postgresRunning) {
    Write-Host "❌ El contenedor de PostgreSQL no está corriendo." -ForegroundColor Red
    Write-Host "   Ejecuta './start-docker.ps1' primero." -ForegroundColor Yellow
    exit 1
}

Write-Host "📝 Creando usuarios de prueba en la base de datos..." -ForegroundColor Yellow

# Ejecutar el script SQL
Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db

Write-Host ""
Write-Host "✅ Usuarios creados exitosamente!" -ForegroundColor Green
Write-Host ""
Write-Host "🔐 Credenciales de prueba:" -ForegroundColor Cyan
Write-Host "   Administrador:" -ForegroundColor White
Write-Host "   - Email: admin@asam.org" -ForegroundColor Gray
Write-Host "   - Password: admin123" -ForegroundColor Gray
Write-Host ""
Write-Host "   Usuario regular:" -ForegroundColor White
Write-Host "   - Email: user@asam.org" -ForegroundColor Gray
Write-Host "   - Password: admin123" -ForegroundColor Gray
Write-Host ""
Write-Host "🚀 Puedes probar el login en:" -ForegroundColor Cyan
Write-Host "   - Frontend: http://localhost:5173" -ForegroundColor White
Write-Host "   - GraphQL Playground: http://localhost:8080/playground" -ForegroundColor White

# Clean Sessions Script

Write-Host "🧹 Limpiando sesiones..." -ForegroundColor Cyan

$confirm = Read-Host "¿Estás seguro que quieres eliminar TODAS las sesiones? (yes/no)"
if ($confirm -ne "yes") {
    Write-Host "Operación cancelada." -ForegroundColor Yellow
    exit
}

Write-Host "`nEliminando todas las sesiones..." -ForegroundColor Yellow
$deleteQuery = "TRUNCATE TABLE refresh_tokens;"
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$deleteQuery"

Write-Host "✅ Todas las sesiones han sido eliminadas." -ForegroundColor Green

Write-Host "`nVerificando..." -ForegroundColor Yellow
$verifyQuery = "SELECT COUNT(*) as total_tokens FROM refresh_tokens;"
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$verifyQuery"

Write-Host "`n💡 Ahora puedes hacer login fresco para ver los nuevos valores." -ForegroundColor Cyan

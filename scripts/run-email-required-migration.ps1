# Script para ejecutar la migración 000007_make_email_required
# Este script hace el campo email obligatorio en la tabla users

Write-Host "Ejecutando migración para hacer el campo email obligatorio..." -ForegroundColor Yellow

# Ejecutar dentro del contenedor Docker
docker exec -i asam-backend-api sh -c "go run ./cmd/migrate -env=local -cmd=up"

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Migración ejecutada exitosamente!" -ForegroundColor Green
    Write-Host "El campo email ahora es obligatorio en la tabla users." -ForegroundColor Cyan
} else {
    Write-Host "`n❌ Error al ejecutar la migración" -ForegroundColor Red
    Write-Host "Puede que necesites revisar los logs o ejecutar la migración manualmente." -ForegroundColor Yellow
}

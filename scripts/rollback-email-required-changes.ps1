# Script para revertir los cambios de email obligatorio (ROLLBACK)

Write-Host "=== Revirtiendo cambios de email obligatorio ===" -ForegroundColor Yellow
Write-Host "⚠️  ADVERTENCIA: Esto hará el campo email opcional nuevamente" -ForegroundColor Red

$confirmation = Read-Host "¿Estás seguro de que quieres revertir estos cambios? (si/no)"
if ($confirmation -ne "si") {
    Write-Host "Operación cancelada." -ForegroundColor Yellow
    exit
}

# Ejecutar la migración de reversión
Write-Host "`nRevirtiendo migración..." -ForegroundColor Yellow

$rollbackContent = @"
-- Revert: Make email column optional (NULL) in users table

-- Remove the NOT NULL constraint
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Remove comment
COMMENT ON COLUMN users.email IS NULL;
"@

$rollbackContent | docker exec -i asam-postgres psql -U postgres -d asam_db

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Cambios revertidos exitosamente!" -ForegroundColor Green
    Write-Host "El campo email ahora es opcional nuevamente." -ForegroundColor Cyan
    
    # Mostrar el estado actual
    Write-Host "`nEstado actual de la columna email:" -ForegroundColor Yellow
    docker exec -i asam-postgres psql -U postgres -d asam_db -c "\d users" | Select-String "email"
} else {
    Write-Host "`n❌ Error al revertir los cambios" -ForegroundColor Red
}

Write-Host "`n⚠️  NOTA: Necesitarás revertir manualmente los cambios en el código Go si es necesario." -ForegroundColor Yellow

# Script para limpiar tokens de verificación usados o expirados

Write-Host "=== Limpieza de Tokens de Verificación ===" -ForegroundColor Cyan

# Primero, mostrar estadísticas actuales
Write-Host "`nEstadísticas actuales de tokens:" -ForegroundColor Yellow

$statsQuery = @"
SELECT 
    type,
    COUNT(*) as total,
    COUNT(CASE WHEN used_at IS NOT NULL THEN 1 END) as usados,
    COUNT(CASE WHEN expires_at < NOW() AND used_at IS NULL THEN 1 END) as expirados,
    COUNT(CASE WHEN expires_at >= NOW() AND used_at IS NULL THEN 1 END) as validos
FROM verification_tokens
GROUP BY type;
"@

$statsQuery | docker exec -i asam-postgres psql -U postgres -d asam_db

# Preguntar si desea proceder con la limpieza
Write-Host "`n¿Desea eliminar todos los tokens usados y expirados?" -ForegroundColor Yellow
Write-Host "Esto no afectará a los tokens válidos que aún no han expirado." -ForegroundColor Cyan
$confirm = Read-Host "Escriba 'SI' para confirmar"

if ($confirm -ne "SI") {
    Write-Host "Operación cancelada." -ForegroundColor Yellow
    exit
}

Write-Host "`nEliminando tokens..." -ForegroundColor Yellow

# Eliminar tokens usados o expirados
$cleanupQuery = @"
-- Eliminar tokens usados
DELETE FROM verification_tokens
WHERE used_at IS NOT NULL;

-- Eliminar tokens expirados
DELETE FROM verification_tokens
WHERE expires_at < NOW() AND used_at IS NULL;
"@

$cleanupQuery | docker exec -i asam-postgres psql -U postgres -d asam_db

# Mostrar estadísticas después de la limpieza
Write-Host "`nEstadísticas después de la limpieza:" -ForegroundColor Green

$statsQuery | docker exec -i asam-postgres psql -U postgres -d asam_db

Write-Host "`n✅ Limpieza completada!" -ForegroundColor Green

# Script para marcar manualmente un usuario como verificado

Write-Host "=== Verificación Manual de Email ===" -ForegroundColor Cyan
Write-Host "⚠️  Este script marca manualmente un usuario como verificado" -ForegroundColor Yellow

# Solicitar el username
$username = Read-Host "Ingrese el username del usuario a verificar (ej: admin@asam.org)"

if ([string]::IsNullOrWhiteSpace($username)) {
    Write-Host "Username no puede estar vacío" -ForegroundColor Red
    exit
}

# Confirmar acción
Write-Host "`n¿Está seguro de que desea marcar a '$username' como verificado?" -ForegroundColor Yellow
$confirm = Read-Host "Escriba 'SI' para confirmar"

if ($confirm -ne "SI") {
    Write-Host "Operación cancelada." -ForegroundColor Yellow
    exit
}

Write-Host "`nActualizando usuario..." -ForegroundColor Yellow

# Actualizar el estado de verificación
$updateQuery = @"
-- Primero, mostrar el estado actual
SELECT id, username, email, email_verified, email_verified_at
FROM users
WHERE username = '$username';

-- Actualizar el estado de verificación
UPDATE users
SET 
    email_verified = true,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE username = '$username';

-- Mostrar el estado actualizado
SELECT id, username, email, email_verified, email_verified_at
FROM users
WHERE username = '$username';

-- Marcar todos los tokens de verificación como usados
UPDATE verification_tokens
SET used_at = NOW()
WHERE user_id = (SELECT id FROM users WHERE username = '$username')
  AND type = 'email_verification'
  AND used_at IS NULL;
"@

$updateQuery | docker exec -i asam-postgres psql -U postgres -d asam_db

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Usuario verificado exitosamente!" -ForegroundColor Green
    Write-Host "El usuario '$username' ahora tiene el email verificado." -ForegroundColor Cyan
} else {
    Write-Host "`n❌ Error al verificar el usuario" -ForegroundColor Red
}

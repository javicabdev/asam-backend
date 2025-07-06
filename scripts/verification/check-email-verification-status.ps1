# Script para verificar el estado de verificación de email de un usuario

Write-Host "=== Verificación de Estado de Email ===" -ForegroundColor Cyan

# Solicitar el username
$username = Read-Host "Ingrese el username del usuario (ej: admin@asam.org)"

if ([string]::IsNullOrWhiteSpace($username)) {
    Write-Host "Username no puede estar vacío" -ForegroundColor Red
    exit
}

Write-Host "`nBuscando usuario '$username'..." -ForegroundColor Yellow

# Consultar el estado del usuario
$query = @"
SELECT 
    id,
    username,
    email,
    email_verified,
    email_verified_at,
    is_active,
    role,
    created_at,
    updated_at
FROM users
WHERE username = '$username';
"@

$query | docker exec -i asam-postgres psql -U postgres -d asam_db

Write-Host "`n=== Tokens de Verificación ===" -ForegroundColor Cyan
Write-Host "Buscando tokens de verificación para este usuario..." -ForegroundColor Yellow

$tokenQuery = @"
SELECT 
    vt.id,
    vt.token,
    vt.type,
    vt.used_at,
    vt.expires_at,
    vt.created_at,
    CASE 
        WHEN vt.used_at IS NOT NULL THEN 'USADO'
        WHEN vt.expires_at < NOW() THEN 'EXPIRADO'
        ELSE 'VÁLIDO'
    END as estado
FROM verification_tokens vt
JOIN users u ON vt.user_id = u.id
WHERE u.username = '$username'
ORDER BY vt.created_at DESC
LIMIT 5;
"@

$tokenQuery | docker exec -i asam-postgres psql -U postgres -d asam_db

Write-Host "`n=== Resumen ===" -ForegroundColor Green
Write-Host "Si email_verified = 't' (true), el email ya está verificado." -ForegroundColor Cyan
Write-Host "Si email_verified = 'f' (false), el email aún no está verificado." -ForegroundColor Yellow
Write-Host "Los tokens con estado 'USADO' ya han sido utilizados y no pueden volver a usarse." -ForegroundColor Yellow

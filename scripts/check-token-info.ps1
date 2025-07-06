# Script para obtener información detallada de un token de verificación

Write-Host "=== Información de Token de Verificación ===" -ForegroundColor Cyan

# Solicitar el token o parte del token
$tokenInput = Read-Host "Ingrese el token completo o los primeros caracteres"

if ([string]::IsNullOrWhiteSpace($tokenInput)) {
    Write-Host "Token no puede estar vacío" -ForegroundColor Red
    exit
}

Write-Host "`nBuscando información del token..." -ForegroundColor Yellow

# Consultar información del token
$query = @"
SELECT 
    vt.id as token_id,
    vt.token,
    vt.type,
    vt.user_id,
    u.username,
    u.email,
    u.email_verified,
    vt.created_at,
    vt.expires_at,
    vt.used_at,
    CASE 
        WHEN vt.used_at IS NOT NULL THEN 'USADO'
        WHEN vt.expires_at < NOW() THEN 'EXPIRADO'
        ELSE 'VÁLIDO'
    END as estado,
    CASE
        WHEN vt.expires_at > NOW() THEN 
            EXTRACT(EPOCH FROM (vt.expires_at - NOW()))/3600 || ' horas'
        ELSE 'N/A'
    END as tiempo_restante
FROM verification_tokens vt
JOIN users u ON vt.user_id = u.id
WHERE vt.token LIKE '$tokenInput%'
ORDER BY vt.created_at DESC
LIMIT 1;
"@

$result = $query | docker exec -i asam-postgres psql -U postgres -d asam_db -t

if ($LASTEXITCODE -eq 0 -and $result -match '\S') {
    # Mostrar la información formateada
    $query | docker exec -i asam-postgres psql -U postgres -d asam_db
    
    Write-Host "`n=== Interpretación ===" -ForegroundColor Green
    
    if ($result -match 'USADO') {
        Write-Host "❌ Este token ya fue usado y no puede volver a utilizarse." -ForegroundColor Red
        Write-Host "   El usuario asociado debería tener el email verificado." -ForegroundColor Yellow
    }
    elseif ($result -match 'EXPIRADO') {
        Write-Host "⏰ Este token ha expirado y ya no es válido." -ForegroundColor Yellow
        Write-Host "   El usuario necesita solicitar un nuevo token de verificación." -ForegroundColor Cyan
    }
    elseif ($result -match 'VÁLIDO') {
        Write-Host "✅ Este token es válido y puede ser usado para verificar el email." -ForegroundColor Green
    }
} else {
    Write-Host "`n❌ No se encontró ningún token que comience con '$tokenInput'" -ForegroundColor Red
    Write-Host "   Verifique que el token sea correcto." -ForegroundColor Yellow
}

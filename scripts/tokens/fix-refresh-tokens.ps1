# Fix existing refresh tokens with zero time value

Write-Host "🔧 Arreglando valores de last_used_at en tokens existentes..." -ForegroundColor Cyan

$updateQuery = @"
-- Update tokens with zero time value to use created_at
UPDATE refresh_tokens
SET last_used_at = created_at
WHERE last_used_at = '0001-01-01 00:00:00+00'
   OR last_used_at IS NULL;
"@

Write-Host "`nEjecutando actualización en la base de datos..." -ForegroundColor Yellow
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$updateQuery"

Write-Host "`nVerificando resultados..." -ForegroundColor Yellow
$verifyQuery = @"
SELECT 
    uuid,
    user_id,
    ip_address,
    device_name,
    created_at,
    last_used_at,
    CASE 
        WHEN last_used_at = '0001-01-01 00:00:00+00' THEN 'ZERO VALUE'
        WHEN last_used_at IS NULL THEN 'NULL'
        ELSE 'OK'
    END as status
FROM refresh_tokens
ORDER BY created_at DESC
LIMIT 10;
"@

docker-compose exec -T postgres psql -U postgres -d asam_db -c "$verifyQuery"

Write-Host "`n✅ Proceso completado!" -ForegroundColor Green

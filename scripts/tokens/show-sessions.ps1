# Show Active Sessions Script

Write-Host "📊 Mostrando sesiones activas..." -ForegroundColor Cyan

$query = @"
-- Sesiones activas por usuario
WITH user_sessions AS (
    SELECT 
        u.username,
        COUNT(*) as total_sessions,
        COUNT(CASE WHEN rt.expires_at > EXTRACT(EPOCH FROM NOW()) THEN 1 END) as active_sessions,
        COUNT(CASE WHEN rt.expires_at <= EXTRACT(EPOCH FROM NOW()) THEN 1 END) as expired_sessions
    FROM users u
    LEFT JOIN refresh_tokens rt ON u.id = rt.user_id
    GROUP BY u.username
)
SELECT * FROM user_sessions ORDER BY total_sessions DESC;
"@

Write-Host "`nResumen de sesiones por usuario:" -ForegroundColor Yellow
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$query"

Write-Host "`n`nDetalle de sesiones activas:" -ForegroundColor Yellow
$detailQuery = @"
SELECT 
    u.username,
    rt.device_name,
    rt.ip_address,
    TO_TIMESTAMP(rt.expires_at) as expires_at,
    rt.created_at,
    rt.last_used_at,
    CASE 
        WHEN rt.last_used_at > rt.created_at THEN 'YES'
        ELSE 'NO'
    END as was_refreshed,
    CASE 
        WHEN rt.expires_at > EXTRACT(EPOCH FROM NOW()) THEN 'ACTIVE'
        ELSE 'EXPIRED'
    END as status
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE rt.expires_at > EXTRACT(EPOCH FROM NOW())
ORDER BY rt.created_at DESC;
"@

docker-compose exec -T postgres psql -U postgres -d asam_db -c "$detailQuery"

Write-Host "`n`n🧹 Para limpiar tokens expirados, ejecuta:" -ForegroundColor Gray
Write-Host "docker-compose exec postgres psql -U postgres -d asam_db -c ""DELETE FROM refresh_tokens WHERE expires_at < EXTRACT(EPOCH FROM NOW());""" -ForegroundColor DarkGray

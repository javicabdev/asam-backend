# Token Status Dashboard

Write-Host @"
╔═══════════════════════════════════════╗
║    Refresh Tokens Status Dashboard    ║
╚═══════════════════════════════════════╝
"@ -ForegroundColor Cyan

# Summary
Write-Host "`n📊 RESUMEN GENERAL" -ForegroundColor Yellow
$summaryQuery = @"
SELECT 
    COUNT(*) as total_tokens,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT device_name) as device_types,
    COUNT(CASE WHEN expires_at > EXTRACT(EPOCH FROM NOW()) THEN 1 END) as active_tokens,
    COUNT(CASE WHEN expires_at <= EXTRACT(EPOCH FROM NOW()) THEN 1 END) as expired_tokens,
    COUNT(CASE WHEN last_used_at = '0001-01-01 00:00:00+00' THEN 1 END) as tokens_with_zero_date
FROM refresh_tokens;
"@
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$summaryQuery"

# Device Distribution
Write-Host "`n📱 DISTRIBUCIÓN POR DISPOSITIVO" -ForegroundColor Yellow
$deviceQuery = @"
SELECT 
    COALESCE(device_name, 'Unknown') as device,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM refresh_tokens
GROUP BY device_name
ORDER BY count DESC;
"@
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$deviceQuery"

# Recent Activity
Write-Host "`n🕐 ACTIVIDAD RECIENTE (Últimas 24 horas)" -ForegroundColor Yellow
$recentQuery = @"
SELECT 
    u.username,
    rt.device_name,
    rt.ip_address,
    rt.created_at,
    CASE 
        WHEN rt.last_used_at = rt.created_at THEN 'Never refreshed'
        ELSE TO_CHAR(rt.last_used_at, 'YYYY-MM-DD HH24:MI:SS')
    END as last_activity
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE rt.created_at > NOW() - INTERVAL '24 hours'
ORDER BY rt.created_at DESC
LIMIT 10;
"@
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$recentQuery"

# Data Quality Check
Write-Host "`n✅ VERIFICACIÓN DE CALIDAD DE DATOS" -ForegroundColor Yellow
$qualityQuery = @"
SELECT 
    'IP Address' as field,
    COUNT(CASE WHEN ip_address IS NOT NULL AND ip_address != '' THEN 1 END) as filled,
    COUNT(*) as total,
    ROUND(COUNT(CASE WHEN ip_address IS NOT NULL AND ip_address != '' THEN 1 END) * 100.0 / COUNT(*), 2) as percentage
FROM refresh_tokens
UNION ALL
SELECT 
    'Device Name' as field,
    COUNT(CASE WHEN device_name IS NOT NULL AND device_name != '' THEN 1 END) as filled,
    COUNT(*) as total,
    ROUND(COUNT(CASE WHEN device_name IS NOT NULL AND device_name != '' THEN 1 END) * 100.0 / COUNT(*), 2) as percentage
FROM refresh_tokens
UNION ALL
SELECT 
    'User Agent' as field,
    COUNT(CASE WHEN user_agent IS NOT NULL AND user_agent != '' THEN 1 END) as filled,
    COUNT(*) as total,
    ROUND(COUNT(CASE WHEN user_agent IS NOT NULL AND user_agent != '' THEN 1 END) * 100.0 / COUNT(*), 2) as percentage
FROM refresh_tokens
UNION ALL
SELECT 
    'Last Used At' as field,
    COUNT(CASE WHEN last_used_at != '0001-01-01 00:00:00+00' THEN 1 END) as filled,
    COUNT(*) as total,
    ROUND(COUNT(CASE WHEN last_used_at != '0001-01-01 00:00:00+00' THEN 1 END) * 100.0 / COUNT(*), 2) as percentage
FROM refresh_tokens;
"@
docker-compose exec -T postgres psql -U postgres -d asam_db -c "$qualityQuery"

Write-Host "`n" -ForegroundColor Gray
Write-Host "💡 Comandos útiles:" -ForegroundColor Gray
Write-Host "  - Ver sesiones activas: .\scripts\tokens\show-sessions.ps1" -ForegroundColor DarkGray
Write-Host "  - Limpiar todas las sesiones: .\scripts\tokens\clean-sessions.ps1" -ForegroundColor DarkGray
Write-Host "  - Hacer test de login: .\scripts\tokens\test-refresh-tokens.ps1" -ForegroundColor DarkGray
Write-Host "  - Arreglar fechas zero: .\scripts\tokens\fix-refresh-tokens.ps1" -ForegroundColor DarkGray

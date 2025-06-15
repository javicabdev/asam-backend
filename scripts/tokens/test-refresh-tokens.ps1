# Test Script para verificar refresh_tokens con información del cliente

Write-Host "Ejecutando test de login para verificar captura de información del cliente..." -ForegroundColor Cyan

# Login request
$loginBody = @{
    operationName = "Login"
    query = @"
mutation Login(`$username: String!, `$password: String!) {
  login(input: {username: `$username, password: `$password}) {
    user {
      id
      username
      role
    }
    accessToken
    refreshToken
    expiresAt
  }
}
"@
    variables = @{
        username = "admin@asam.org"
        password = "admin123"
    }
} | ConvertTo-Json -Depth 10

$headers = @{
    "Content-Type" = "application/json"
    "User-Agent" = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    "X-Forwarded-For" = "192.168.1.100"
    "X-Real-IP" = "192.168.1.100"
}

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/graphql" `
        -Method POST `
        -Headers $headers `
        -Body $loginBody `
        -UseBasicParsing
    
    $data = $response.Content | ConvertFrom-Json
    
    if ($data.data.login) {
        Write-Host "✅ Login exitoso!" -ForegroundColor Green
        Write-Host "Usuario: $($data.data.login.user.username)" -ForegroundColor Gray
        Write-Host "Role: $($data.data.login.user.role)" -ForegroundColor Gray
        
        # Ahora verificar en la BD
        Write-Host "`nVerificando en la base de datos..." -ForegroundColor Yellow
        
        $query = @"
SELECT 
    rt.uuid,
    rt.user_id,
    rt.ip_address,
    rt.device_name,
    LENGTH(rt.user_agent) as ua_length,
    rt.created_at,
    rt.last_used_at,
    CASE 
        WHEN rt.last_used_at = '0001-01-01 00:00:00+00' THEN 'ZERO VALUE - ERROR'
        WHEN rt.last_used_at IS NULL THEN 'NULL'
        WHEN rt.last_used_at = rt.created_at THEN 'SAME AS CREATED'
        ELSE 'UPDATED'
    END as last_used_status
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE u.username = 'admin@asam.org'
ORDER BY rt.created_at DESC
LIMIT 5;
"@
        
        docker-compose exec -T postgres psql -U postgres -d asam_db -c $query
        
        # Mostrar también el user agent completo del último token
        Write-Host "`nUser Agent del último token:" -ForegroundColor Yellow
        $uaQuery = @"
SELECT user_agent
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
WHERE u.username = 'admin@asam.org'
ORDER BY rt.created_at DESC
LIMIT 1;
"@
        docker-compose exec -T postgres psql -U postgres -d asam_db -t -c $uaQuery
        
    } else {
        Write-Host "❌ Error en login" -ForegroundColor Red
        $data | ConvertTo-Json -Depth 10
    }
} catch {
    Write-Host "❌ Error al hacer la petición: $_" -ForegroundColor Red
}

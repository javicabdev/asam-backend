# Script para verificar el estado del dominio ASAM
# Uso: .\verify-domain.ps1

$domain = "mutuaasam.org"

Write-Host "🔍 Verificando dominio: $domain" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# 1. Verificar si el dominio responde
Write-Host "`n📡 1. Verificando DNS..." -ForegroundColor Yellow
try {
    $dns = Resolve-DnsName -Name $domain -ErrorAction Stop
    Write-Host "✅ El dominio está registrado y los DNS responden" -ForegroundColor Green
    $dns | Format-Table -AutoSize
} catch {
    Write-Host "❌ El dominio no responde a consultas DNS" -ForegroundColor Red
    Write-Host "   Esto es normal si acabas de registrarlo (puede tardar hasta 48h)" -ForegroundColor Yellow
}

# 2. Verificar registros DNS específicos
Write-Host "`n📋 2. Registros DNS actuales:" -ForegroundColor Yellow

# Registros A
Write-Host "`nRegistros A:" -ForegroundColor Cyan
try {
    $aRecords = Resolve-DnsName -Name $domain -Type A -ErrorAction SilentlyContinue
    if ($aRecords) {
        $aRecords | Select-Object Name, IPAddress | Format-Table -AutoSize
    } else {
        Write-Host "No hay registros A configurados aún" -ForegroundColor Gray
    }
} catch {
    Write-Host "No se encontraron registros A" -ForegroundColor Gray
}

# Registros CNAME para www
Write-Host "`nRegistros CNAME (www):" -ForegroundColor Cyan
try {
    $cnameRecords = Resolve-DnsName -Name "www.$domain" -Type CNAME -ErrorAction SilentlyContinue
    if ($cnameRecords) {
        $cnameRecords | Select-Object Name, NameHost | Format-Table -AutoSize
    } else {
        Write-Host "No hay registros CNAME configurados aún" -ForegroundColor Gray
    }
} catch {
    Write-Host "No se encontraron registros CNAME" -ForegroundColor Gray
}

# Registros MX (email)
Write-Host "`nRegistros MX (email):" -ForegroundColor Cyan
try {
    $mxRecords = Resolve-DnsName -Name $domain -Type MX -ErrorAction SilentlyContinue
    if ($mxRecords) {
        $mxRecords | Select-Object Name, NameExchange, Preference | Format-Table -AutoSize
    } else {
        Write-Host "No hay registros MX configurados aún" -ForegroundColor Gray
    }
} catch {
    Write-Host "No se encontraron registros MX" -ForegroundColor Gray
}

# 3. Verificar nameservers
Write-Host "`n🌐 3. Nameservers:" -ForegroundColor Yellow
try {
    $nsRecords = Resolve-DnsName -Name $domain -Type NS -ErrorAction SilentlyContinue
    if ($nsRecords) {
        $nsRecords | Select-Object NameHost | Format-Table -AutoSize
    }
} catch {
    Write-Host "No se pudieron obtener los nameservers" -ForegroundColor Gray
}

# 4. Verificar respuesta HTTP
Write-Host "`n🌐 4. Verificando respuesta HTTP/HTTPS:" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://$domain" -Method Head -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Hay respuesta del servidor web (Status: $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "⚠️  No hay servidor web configurado aún (normal si acabas de registrar)" -ForegroundColor Yellow
}

# 5. Información adicional
Write-Host "`n📌 5. Próximos pasos:" -ForegroundColor Cyan
Write-Host "   1. Configurar los DNS en tu registrador" -ForegroundColor White
Write-Host "   2. Esperar propagación DNS (hasta 48 horas)" -ForegroundColor White
Write-Host "   3. Configurar email (MX records)" -ForegroundColor White
Write-Host "   4. Conectar con Netlify (frontend)" -ForegroundColor White
Write-Host "   5. Configurar subdominio api.* para Cloud Run" -ForegroundColor White

Write-Host "`n🔗 Enlaces útiles:" -ForegroundColor Cyan
Write-Host "   - Verificar propagación: https://www.whatsmydns.net/#A/$domain" -ForegroundColor White
Write-Host "   - DNS Checker: https://dnschecker.org/all-dns-records-of-domain.php?query=$domain" -ForegroundColor White

Write-Host "`n✅ Verificación completada" -ForegroundColor Green

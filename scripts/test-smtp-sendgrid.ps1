# Test SMTP SendGrid Configuration
Write-Host "Testing SMTP configuration for SendGrid..." -ForegroundColor Cyan

# Read configuration from .env
$envPath = Join-Path $PSScriptRoot ".." ".env"
if (Test-Path $envPath) {
    Get-Content $envPath | ForEach-Object {
        if ($_ -match '^([^#=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, [EnvironmentVariableTarget]::Process)
        }
    }
}

$smtpServer = $env:SMTP_SERVER
$smtpPort = $env:SMTP_PORT
$smtpUser = $env:SMTP_USER
$smtpPass = $env:SMTP_PASSWORD
$from = $env:SMTP_FROM_EMAIL
$to = "javierfernandezc@gmail.com"  # Tu email personal para pruebas

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Server: $smtpServer`:$smtpPort" -ForegroundColor Gray
Write-Host "  User: $smtpUser" -ForegroundColor Gray
Write-Host "  From: $from" -ForegroundColor Gray
Write-Host "  To: $to" -ForegroundColor Gray

try {
    $password = ConvertTo-SecureString $smtpPass -AsPlainText -Force
    $credential = New-Object System.Management.Automation.PSCredential($smtpUser, $password)
    
    $mailParams = @{
        SmtpServer = $smtpServer
        Port = $smtpPort
        UseSsl = $true
        Credential = $credential
        From = $from
        To = $to
        Subject = "Test ASAM - SendGrid Configuration"
        Body = @"
<html>
<body>
    <h2>¡Éxito! 🎉</h2>
    <p>La configuración SMTP de SendGrid está funcionando correctamente.</p>
    <p>Este es un email de prueba enviado desde el backend de ASAM.</p>
    <hr>
    <p><small>Enviado: $(Get-Date -Format "dd/MM/yyyy HH:mm:ss")</small></p>
    <p><small>Servidor: SendGrid</small></p>
</body>
</html>
"@
        BodyAsHtml = $true
    }
    
    Send-MailMessage @mailParams
    Write-Host "✅ Email enviado correctamente a $to" -ForegroundColor Green
    Write-Host "Por favor, verifica tu bandeja de entrada." -ForegroundColor Yellow
}
catch {
    Write-Host "❌ Error al enviar email:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host "`nPosibles causas:" -ForegroundColor Yellow
    Write-Host "- La API key no es correcta" -ForegroundColor Yellow
    Write-Host "- La API key no tiene permisos de 'Mail Send'" -ForegroundColor Yellow
    Write-Host "- Tu cuenta de SendGrid no está activa" -ForegroundColor Yellow
    Write-Host "- Los DNS del dominio no están verificados" -ForegroundColor Yellow
}

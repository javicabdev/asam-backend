# PowerShell script to set required environment variables for ASAM Backend in Cloud Run

param(
    [string]$ServiceName = "asam-backend",
    [string]$Region = "europe-west1"
)

Write-Host "Configurando variables de entorno para $ServiceName en Cloud Run..." -ForegroundColor Yellow

# Check if service exists
try {
    $service = gcloud run services describe $ServiceName --region=$Region 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: El servicio $ServiceName no existe en la región $Region" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error verificando el servicio: $_" -ForegroundColor Red
    exit 1
}

# Function to generate secure random secrets
function Generate-Secret {
    $bytes = New-Object byte[] 32
    [Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
    return [Convert]::ToBase64String($bytes).Substring(0, 32) -replace '[+/=]', ''
}

Write-Host "Generando secretos seguros..." -ForegroundColor Green
$JWT_ACCESS_SECRET = Generate-Secret
$JWT_REFRESH_SECRET = Generate-Secret

# Prompt for required values
Write-Host "`nPor favor, proporciona los siguientes valores:" -ForegroundColor Yellow
$ADMIN_USER = Read-Host "ADMIN_USER (usuario administrador)"
$ADMIN_PASSWORD = Read-Host "ADMIN_PASSWORD (contraseña administrador)" -AsSecureString
$ADMIN_PASSWORD_PLAIN = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($ADMIN_PASSWORD))

# SMTP is optional
Write-Host "`nConfiguración SMTP (opcional - presiona Enter para omitir):" -ForegroundColor Yellow
$SMTP_USER = Read-Host "SMTP_USER (usuario SMTP)"
$SMTP_PASSWORD = Read-Host "SMTP_PASSWORD (contraseña SMTP)" -AsSecureString

# Convert secure string to plain text if provided
if ($SMTP_PASSWORD.Length -gt 0) {
    $SMTP_PASSWORD_PLAIN = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($SMTP_PASSWORD))
} else {
    $SMTP_PASSWORD_PLAIN = ""
}

# Set defaults if not provided
if ([string]::IsNullOrWhiteSpace($SMTP_USER)) {
    $SMTP_USER = "noreply@asam.org"
    $SMTP_PASSWORD_PLAIN = "temp-smtp-pass"
}

# Build environment variables string
# Note: PORT is automatically set by Cloud Run and should NOT be overridden
$envVars = @(
    "ENVIRONMENT=production",
    "ADMIN_USER=$ADMIN_USER",
    "ADMIN_PASSWORD=$ADMIN_PASSWORD_PLAIN",
    "JWT_ACCESS_SECRET=$JWT_ACCESS_SECRET",
    "JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET",
    "SMTP_USER=$SMTP_USER",
    "SMTP_PASSWORD=$SMTP_PASSWORD_PLAIN"
) -join ','

Write-Host "`nActualizando variables de entorno..." -ForegroundColor Yellow

# Update environment variables
$updateCommand = "gcloud run services update $ServiceName --region=$Region --update-env-vars `"$envVars`""
Invoke-Expression $updateCommand

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nVariables de entorno actualizadas exitosamente!" -ForegroundColor Green
    Write-Host "Nota: Guarda estos valores en un lugar seguro:" -ForegroundColor Yellow
    Write-Host "JWT_ACCESS_SECRET=$JWT_ACCESS_SECRET"
    Write-Host "JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET"
} else {
    Write-Host "`nError al actualizar las variables de entorno" -ForegroundColor Red
    exit 1
}

Write-Host "`nConfiguración completada!" -ForegroundColor Green

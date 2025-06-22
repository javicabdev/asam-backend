# Script para crear secrets en Google Secret Manager
# Uso: .\create-secrets.ps1

param(
    [string]$ProjectId = $env:PROJECT_ID
)

Write-Host "🔐 Creando secrets en Google Secret Manager..." -ForegroundColor Cyan

# Función para crear o actualizar un secret
function CreateOrUpdateSecret {
    param(
        [string]$SecretName,
        [string]$SecretValue
    )
    
    # Verificar si el secret existe
    $exists = gcloud secrets describe $SecretName --project=$ProjectId 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Actualizando secret: $SecretName" -ForegroundColor Yellow
        $SecretValue | gcloud secrets versions add $SecretName --data-file=- --project=$ProjectId
    } else {
        Write-Host "Creando secret: $SecretName" -ForegroundColor Green
        $SecretValue | gcloud secrets create $SecretName --data-file=- --project=$ProjectId
    }
}

# Verificar que PROJECT_ID esté configurado
if (-not $ProjectId) {
    Write-Host "❌ Error: PROJECT_ID no está configurado" -ForegroundColor Red
    Write-Host "Ejecuta: `$env:PROJECT_ID='tu-proyecto-id'" -ForegroundColor Yellow
    Write-Host "O pasa el parámetro: .\create-secrets.ps1 -ProjectId 'tu-proyecto-id'" -ForegroundColor Yellow
    exit 1
}

Write-Host "📋 Proyecto: $ProjectId" -ForegroundColor Green

# Habilitar Secret Manager API si no está habilitada
Write-Host "Habilitando Secret Manager API..." -ForegroundColor Yellow
gcloud services enable secretmanager.googleapis.com --project=$ProjectId

# Database secrets
Write-Host "`n🗄️  Configurando secrets de base de datos..." -ForegroundColor Cyan
CreateOrUpdateSecret -SecretName "db-host" -SecretValue "pg-asam-asam-backend-db.l.aivencloud.com"
CreateOrUpdateSecret -SecretName "db-port" -SecretValue "14276"
CreateOrUpdateSecret -SecretName "db-user" -SecretValue "avnadmin"

$dbPassword = Read-Host "Ingresa DB_PASSWORD" -AsSecureString
$dbPasswordPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($dbPassword))
CreateOrUpdateSecret -SecretName "db-password" -SecretValue $dbPasswordPlain

CreateOrUpdateSecret -SecretName "db-name" -SecretValue "asam-backend-db"

# JWT secrets
Write-Host "`n🔑 Configurando secrets JWT..." -ForegroundColor Cyan
Write-Host "Generando JWT secrets aleatorios..." -ForegroundColor Yellow

# Generar secrets aleatorios
$bytes = New-Object byte[] 32
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
$jwtAccessSecret = [Convert]::ToBase64String($bytes)

[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
$jwtRefreshSecret = [Convert]::ToBase64String($bytes)

CreateOrUpdateSecret -SecretName "jwt-access-secret" -SecretValue $jwtAccessSecret
CreateOrUpdateSecret -SecretName "jwt-refresh-secret" -SecretValue $jwtRefreshSecret

# Admin credentials
Write-Host "`n👤 Configurando credenciales de admin..." -ForegroundColor Cyan
$adminUser = Read-Host "Ingresa ADMIN_USER [admin]"
if ([string]::IsNullOrEmpty($adminUser)) { $adminUser = "admin" }

$adminPassword = Read-Host "Ingresa ADMIN_PASSWORD" -AsSecureString
$adminPasswordPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($adminPassword))

CreateOrUpdateSecret -SecretName "admin-user" -SecretValue $adminUser
CreateOrUpdateSecret -SecretName "admin-password" -SecretValue $adminPasswordPlain

# SMTP (opcional)
$configureSMTP = Read-Host "`n📧 ¿Configurar SMTP? (s/n)"
if ($configureSMTP -eq "s") {
    $smtpUser = Read-Host "SMTP_USER"
    $smtpPassword = Read-Host "SMTP_PASSWORD" -AsSecureString
    $smtpPasswordPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($smtpPassword))
    
    CreateOrUpdateSecret -SecretName "smtp-user" -SecretValue $smtpUser
    CreateOrUpdateSecret -SecretName "smtp-password" -SecretValue $smtpPasswordPlain
}

Write-Host "`n✅ Secrets creados exitosamente!" -ForegroundColor Green

# Dar permisos a la cuenta de servicio de Cloud Run
Write-Host "`n🔧 Configurando permisos para Cloud Run..." -ForegroundColor Cyan

# Obtener la cuenta de servicio predeterminada
$serviceAccount = gcloud iam service-accounts list --filter="displayName:'Compute Engine default service account'" --format="value(email)" --project=$ProjectId

if ($serviceAccount) {
    Write-Host "Otorgando acceso a secrets para: $serviceAccount" -ForegroundColor Yellow
    
    # Lista de secrets
    $secrets = @(
        "db-host", "db-port", "db-user", "db-password", "db-name",
        "jwt-access-secret", "jwt-refresh-secret",
        "admin-user", "admin-password"
    )
    
    # Agregar SMTP si se configuró
    if ($configureSMTP -eq "s") {
        $secrets += @("smtp-user", "smtp-password")
    }
    
    # Otorgar permisos
    foreach ($secret in $secrets) {
        gcloud secrets add-iam-policy-binding $secret `
            --member="serviceAccount:$serviceAccount" `
            --role="roles/secretmanager.secretAccessor" `
            --project=$ProjectId 2>$null
    }
    
    Write-Host "✅ Permisos configurados" -ForegroundColor Green
} else {
    Write-Host "⚠️  No se pudo encontrar la cuenta de servicio. Configura los permisos manualmente." -ForegroundColor Yellow
}

Write-Host "`n🎉 Configuración completada!" -ForegroundColor Green
Write-Host "Los secrets están listos para usar en Cloud Run" -ForegroundColor Cyan

# Mostrar resumen
Write-Host "`n📝 Resumen de secrets creados:" -ForegroundColor Cyan
Write-Host "  - Base de datos: db-host, db-port, db-user, db-password, db-name" -ForegroundColor White
Write-Host "  - JWT: jwt-access-secret, jwt-refresh-secret" -ForegroundColor White
Write-Host "  - Admin: admin-user, admin-password" -ForegroundColor White
if ($configureSMTP -eq "s") {
    Write-Host "  - SMTP: smtp-user, smtp-password" -ForegroundColor White
}

Write-Host "`n💡 Para desplegar a Cloud Run, ejecuta el workflow de GitHub Actions" -ForegroundColor Yellow

# Script para verificar el estado de los secretos en Google Secret Manager

$PROJECT_ID = "babacar-asam"
$SERVICE_ACCOUNT = "67685084900-compute@developer.gserviceaccount.com"

Write-Host "🔍 Verificando estado de los secretos en Google Secret Manager..." -ForegroundColor Cyan
Write-Host "📌 Proyecto: $PROJECT_ID" -ForegroundColor Yellow
Write-Host ""

# Verificar si los secretos existen
Write-Host "📋 Lista de secretos en el proyecto:" -ForegroundColor Green
gcloud secrets list --project=$PROJECT_ID --format="table(name,createTime)"

Write-Host ""
Write-Host "🔐 Verificando permisos de la cuenta de servicio..." -ForegroundColor Green
Write-Host "📌 Cuenta: $SERVICE_ACCOUNT" -ForegroundColor Yellow
Write-Host ""

# Verificar roles actuales de la cuenta de servicio
Write-Host "📋 Roles actuales de la cuenta de servicio:" -ForegroundColor Green
gcloud projects get-iam-policy $PROJECT_ID `
    --flatten="bindings[].members" `
    --filter="bindings.members:serviceAccount:$SERVICE_ACCOUNT" `
    --format="table(bindings.role)"

Write-Host ""
Write-Host "💡 Consejo: Si no ves 'roles/secretmanager.secretAccessor' en la lista," -ForegroundColor Yellow
Write-Host "   ejecuta el script Grant-SecretAccess.ps1 para otorgar los permisos." -ForegroundColor Yellow

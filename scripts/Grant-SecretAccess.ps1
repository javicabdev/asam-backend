# Script para otorgar permisos de Secret Manager a la cuenta de servicio de Cloud Run

$PROJECT_ID = "babacar-asam"
$PROJECT_NUMBER = "67685084900"
$SERVICE_ACCOUNT = "${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# Lista de secretos
$SECRETS = @(
    "db-host",
    "db-port",
    "db-user",
    "db-password",
    "db-name",
    "jwt-access-secret",
    "jwt-refresh-secret",
    "admin-user",
    "admin-password",
    "smtp-user",
    "smtp-password"
)

Write-Host "🔧 Otorgando permisos de Secret Manager a la cuenta de servicio de Cloud Run..." -ForegroundColor Cyan
Write-Host "📌 Cuenta de servicio: $SERVICE_ACCOUNT" -ForegroundColor Yellow
Write-Host ""

# Opción 1: Otorgar permisos a nivel de proyecto (más simple pero más amplio)
Write-Host "🚀 Otorgando rol de Secret Manager Secret Accessor a nivel de proyecto..." -ForegroundColor Green
gcloud projects add-iam-policy-binding $PROJECT_ID `
    --member="serviceAccount:${SERVICE_ACCOUNT}" `
    --role="roles/secretmanager.secretAccessor"

Write-Host ""
Write-Host "✅ Permisos otorgados a nivel de proyecto" -ForegroundColor Green
Write-Host ""

# Opción 2: Si prefieres otorgar permisos a nivel de secreto individual (más granular)
# Descomenta las siguientes líneas si prefieres este enfoque:

# foreach ($SECRET in $SECRETS) {
#     Write-Host "🔐 Otorgando acceso al secreto: $SECRET" -ForegroundColor Cyan
#     gcloud secrets add-iam-policy-binding $SECRET `
#         --project=$PROJECT_ID `
#         --member="serviceAccount:${SERVICE_ACCOUNT}" `
#         --role="roles/secretmanager.secretAccessor"
# }

Write-Host ""
Write-Host "✅ ¡Listo! Ahora intenta desplegar de nuevo en Cloud Run." -ForegroundColor Green
Write-Host ""
Write-Host "📝 Si el problema persiste, verifica que:" -ForegroundColor Yellow
Write-Host "   1. Los secretos existen con: gcloud secrets list --project=$PROJECT_ID"
Write-Host "   2. Los secretos tienen valores con: gcloud secrets versions list SECRET_NAME --project=$PROJECT_ID"

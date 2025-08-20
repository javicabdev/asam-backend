# Script para configurar Google Artifact Registry para ASAM Backend

param(
    [string]$ProjectId = $(gcloud config get-value project)
)

$Region = "europe-west1"
$RepositoryName = "asam-backend"

Write-Host "=== Configurando Google Artifact Registry ===" -ForegroundColor Cyan
Write-Host "Project: $ProjectId" -ForegroundColor Gray
Write-Host "Region: $Region" -ForegroundColor Gray
Write-Host "Repository: $RepositoryName" -ForegroundColor Gray
Write-Host ""

# Habilitar la API de Artifact Registry
Write-Host "1. Habilitando Artifact Registry API..." -ForegroundColor Yellow
gcloud services enable artifactregistry.googleapis.com --project=$ProjectId

# Crear el repositorio
Write-Host ""
Write-Host "2. Creando repositorio Docker..." -ForegroundColor Yellow
gcloud artifacts repositories create $RepositoryName `
    --repository-format=docker `
    --location=$Region `
    --description="Docker images for ASAM Backend" `
    --project=$ProjectId

# Verificar que se creó
Write-Host ""
Write-Host "3. Verificando repositorio..." -ForegroundColor Yellow
gcloud artifacts repositories describe $RepositoryName `
    --location=$Region `
    --project=$ProjectId

Write-Host ""
Write-Host "✅ Configuración completada!" -ForegroundColor Green
Write-Host ""
Write-Host "URL del repositorio:" -ForegroundColor Cyan
Write-Host "${Region}-docker.pkg.dev/${ProjectId}/${RepositoryName}" -ForegroundColor Gray
Write-Host ""
Write-Host "Para autenticarte localmente:" -ForegroundColor Cyan
Write-Host "gcloud auth configure-docker ${Region}-docker.pkg.dev" -ForegroundColor Gray

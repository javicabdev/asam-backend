# Script para configurar permisos de Google Container Registry
# Este script debe ejecutarse por alguien con permisos de administrador en el proyecto

param(
    [Parameter(Position=0)]
    [string]$ProjectId = $env:GCP_PROJECT_ID,
    
    [Parameter(Position=1)]
    [string]$ServiceAccountEmail
)

if (-not $ProjectId) {
    Write-Error "Error: Debes proporcionar el PROJECT_ID como primer argumento o establecer GCP_PROJECT_ID"
    Write-Host "Uso: .\fix-gcr-permissions.ps1 <PROJECT_ID> [SERVICE_ACCOUNT_EMAIL]"
    exit 1
}

if (-not $ServiceAccountEmail) {
    $ServiceAccountEmail = "github-actions-deploy@${ProjectId}.iam.gserviceaccount.com"
}

Write-Host "=== Configurando permisos de GCR para el proyecto $ProjectId ===" -ForegroundColor Green
Write-Host "Cuenta de servicio: $ServiceAccountEmail"

# Configurar el proyecto
Write-Host ""
Write-Host "Configurando proyecto..." -ForegroundColor Yellow
gcloud config set project $ProjectId

# Opción 1: Habilitar la API de Container Registry (si no está habilitada)
Write-Host ""
Write-Host "1. Habilitando API de Container Registry..." -ForegroundColor Yellow
gcloud services enable containerregistry.googleapis.com

# Opción 2: Agregar permisos necesarios a la cuenta de servicio
Write-Host ""
Write-Host "2. Agregando roles necesarios a la cuenta de servicio..." -ForegroundColor Yellow

# Roles necesarios
$roles = @(
    "roles/storage.admin",  # Para crear repositorios en GCR
    "roles/cloudbuild.builds.builder"  # Para construir imágenes
)

foreach ($role in $roles) {
    Write-Host "   Agregando rol: $role"
    gcloud projects add-iam-policy-binding $ProjectId `
        --member="serviceAccount:$ServiceAccountEmail" `
        --role="$role" `
        --quiet
}

# Opción 3: Crear el repositorio manualmente (opcional)
Write-Host ""
Write-Host "3. Creando el repositorio inicial en GCR..." -ForegroundColor Yellow

# Verificar si Docker está disponible
try {
    $null = Get-Command docker -ErrorAction Stop
    $dockerInfo = docker info 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   Docker está disponible, creando repositorio..."
        
        # Configurar Docker para GCR
        gcloud auth configure-docker --quiet
        
        # Crear un repositorio vacío subiendo una imagen dummy
        docker pull busybox:latest
        docker tag busybox:latest "gcr.io/${ProjectId}/asam-backend:init"
        docker push "gcr.io/${ProjectId}/asam-backend:init"
        
        # Eliminar la imagen inicial
        gcloud container images delete "gcr.io/${ProjectId}/asam-backend:init" --quiet
        
        Write-Host "✅ Repositorio creado exitosamente" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Docker no está ejecutándose" -ForegroundColor Yellow
        Write-Host "   El repositorio se creará automáticamente en el primer push desde GitHub Actions"
        Write-Host "   Esto no es un problema, el Release Pipeline funcionará correctamente"
    }
} catch {
    Write-Host "⚠️  Docker no está instalado localmente" -ForegroundColor Yellow
    Write-Host "   El repositorio se creará automáticamente en el primer push desde GitHub Actions"
    Write-Host "   Esto no es un problema, el Release Pipeline funcionará correctamente"
}

Write-Host ""
Write-Host "✅ Configuración completada!" -ForegroundColor Green
Write-Host ""
Write-Host "La cuenta de servicio ahora tiene permisos para:" -ForegroundColor Cyan
Write-Host "- Crear repositorios en GCR automáticamente"
Write-Host "- Subir y descargar imágenes"
Write-Host ""
Write-Host "Vuelve a ejecutar el workflow de Release y debería funcionar." -ForegroundColor Green

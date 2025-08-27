# Script para validar deployments antes de ejecutarlos
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("staging", "production")]
    [string]$Environment,
    
    [Parameter(Mandatory=$true)]
    [string]$ImageTag
)

Write-Host "=== Validación de Deployment ===" -ForegroundColor Cyan
Write-Host "Environment: $Environment" -ForegroundColor Yellow
Write-Host "Image Tag: $ImageTag" -ForegroundColor Yellow
Write-Host ""

# Función para validar tag semántico
function Test-SemanticVersion {
    param([string]$Tag)
    return $Tag -match '^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$'
}

$isValid = $true
$warnings = @()
$errors = @()

# Validaciones según ambiente
if ($Environment -eq "production") {
    if (-not (Test-SemanticVersion $ImageTag)) {
        $errors += "Production requiere tags semánticos (ej: v1.0.0, v2.1.3)"
        $isValid = $false
    }
    
    if ($ImageTag -eq "latest") {
        $errors += "Production NO puede usar 'latest'"
        $isValid = $false
    }
    
    if ($ImageTag -match '-dev$' -or $ImageTag -match '-test$') {
        $errors += "Production no debe usar tags de desarrollo (-dev, -test)"
        $isValid = $false
    }
} else {
    # Staging - solo warnings
    if ($ImageTag -eq "latest") {
        $warnings += "Usando 'latest' - el contenido puede cambiar"
    }
    
    if (-not (Test-SemanticVersion $ImageTag)) {
        $warnings += "Considera usar tags semánticos para mejor trazabilidad"
    }
}

# Verificar que la imagen existe
Write-Host "Verificando que la imagen existe..." -ForegroundColor Yellow
$imageExists = gcloud container images describe "gcr.io/babacar-asam/asam-backend:$ImageTag" 2>$null
if ($LASTEXITCODE -ne 0) {
    $errors += "La imagen gcr.io/babacar-asam/asam-backend:$ImageTag NO existe"
    $isValid = $false
} else {
    Write-Host "✅ Imagen encontrada" -ForegroundColor Green
}

# Mostrar resultados
Write-Host ""
if ($warnings.Count -gt 0) {
    Write-Host "⚠️  Advertencias:" -ForegroundColor Yellow
    $warnings | ForEach-Object { Write-Host "   - $_" -ForegroundColor Yellow }
}

if ($errors.Count -gt 0) {
    Write-Host "❌ Errores:" -ForegroundColor Red
    $errors | ForEach-Object { Write-Host "   - $_" -ForegroundColor Red }
}

if ($isValid) {
    Write-Host ""
    Write-Host "✅ Deployment válido" -ForegroundColor Green
    
    # Mostrar comando para deploy
    Write-Host ""
    Write-Host "Comando para deployar:" -ForegroundColor Cyan
    if ($Environment -eq "staging") {
        $serviceName = "asam-backend-staging"
    } else {
        $serviceName = "asam-backend"
    }
    
    Write-Host "gcloud run deploy $serviceName ``" -ForegroundColor White
    Write-Host "  --image=gcr.io/babacar-asam/asam-backend:$ImageTag ``" -ForegroundColor White
    Write-Host "  --region=europe-west1 ``" -ForegroundColor White
    Write-Host "  --platform=managed" -ForegroundColor White
    
    exit 0
} else {
    Write-Host ""
    Write-Host "❌ Deployment NO válido" -ForegroundColor Red
    exit 1
}

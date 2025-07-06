# Script para verificar la configuración de GCP antes de ejecutar un release
# Ayuda a detectar problemas comunes antes de que fallen los workflows

Write-Host "=== Verificando configuración de GCP para ASAM Backend ===" -ForegroundColor Cyan
Write-Host ""

# Verificar que gcloud está instalado
try {
    $null = Get-Command gcloud -ErrorAction Stop
} catch {
    Write-Host "❌ Error: gcloud CLI no está instalado" -ForegroundColor Red
    Write-Host "   Instálalo desde: https://cloud.google.com/sdk/docs/install"
    exit 1
}

# Obtener el proyecto actual
$currentProject = gcloud config get-value project 2>$null
if (-not $currentProject) {
    Write-Host "❌ Error: No hay un proyecto configurado en gcloud" -ForegroundColor Red
    Write-Host "   Ejecuta: gcloud config set project <PROJECT_ID>"
    exit 1
}

Write-Host "📋 Proyecto actual: $currentProject" -ForegroundColor Green
Write-Host ""

# Verificar autenticación
Write-Host "🔐 Verificando autenticación..." -ForegroundColor Yellow
$activeAccount = gcloud auth list --filter="status:ACTIVE" --format="value(account)" 2>$null
if (-not $activeAccount) {
    Write-Host "❌ Error: No estás autenticado en gcloud" -ForegroundColor Red
    Write-Host "   Ejecuta: gcloud auth login"
    exit 1
}
Write-Host "✅ Autenticado correctamente como: $activeAccount" -ForegroundColor Green

# Verificar APIs habilitadas
Write-Host ""
Write-Host "🔧 Verificando APIs necesarias..." -ForegroundColor Yellow
$requiredApis = @(
    "containerregistry.googleapis.com",
    "run.googleapis.com",
    "cloudbuild.googleapis.com"
)

foreach ($api in $requiredApis) {
    $enabledApi = gcloud services list --enabled --filter="name:$api" --format="value(name)" 2>$null
    if ($enabledApi -match $api) {
        Write-Host "✅ $api está habilitada" -ForegroundColor Green
    } else {
        Write-Host "❌ $api NO está habilitada" -ForegroundColor Red
        Write-Host "   Ejecuta: gcloud services enable $api"
    }
}

# Verificar cuenta de servicio
Write-Host ""
Write-Host "👤 Verificando cuenta de servicio GitHub Actions..." -ForegroundColor Yellow
$saEmail = "github-actions-deploy@${currentProject}.iam.gserviceaccount.com"

try {
    $null = gcloud iam service-accounts describe $saEmail 2>$null
    Write-Host "✅ Cuenta de servicio existe: $saEmail" -ForegroundColor Green
    
    # Verificar roles
    Write-Host ""
    Write-Host "🔑 Verificando roles de la cuenta de servicio..." -ForegroundColor Yellow
    
    $iamPolicy = gcloud projects get-iam-policy $currentProject --format=json | ConvertFrom-Json
    $saRoles = @()
    
    foreach ($binding in $iamPolicy.bindings) {
        if ($binding.members -contains "serviceAccount:$saEmail") {
            $saRoles += $binding.role
        }
    }
    
    $requiredRoles = @{
        "roles/run.admin" = "Desplegar en Cloud Run"
        "roles/cloudbuild.builds.builder" = "Construir imágenes"
        "roles/iam.serviceAccountUser" = "Actuar como cuenta de servicio"
        "roles/storage.admin" = "Crear repositorios en GCR"
    }
    
    foreach ($role in $requiredRoles.Keys) {
        if ($saRoles -contains $role) {
            Write-Host "✅ Rol asignado: $role" -ForegroundColor Green
        } else {
            Write-Host "❌ Rol NO asignado: $role" -ForegroundColor Red
            Write-Host "   Necesario para: $($requiredRoles[$role])"
        }
    }
} catch {
    Write-Host "❌ La cuenta de servicio NO existe" -ForegroundColor Red
    Write-Host "   Créala siguiendo docs/gcp-project-setup.md"
}

# Verificar permisos de Docker/GCR
Write-Host ""
Write-Host "🐳 Verificando acceso a Google Container Registry..." -ForegroundColor Yellow

# Primero verificar si Docker está instalado y ejecutándose
try {
    $null = Get-Command docker -ErrorAction Stop
    $dockerVersion = docker --version 2>$null
    if ($dockerVersion) {
        Write-Host "✅ Docker instalado: $dockerVersion" -ForegroundColor Green
        
        # Verificar si Docker está ejecutándose
        $dockerInfo = docker info 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Docker está ejecutándose" -ForegroundColor Green
            
            # Verificar configuración de gcloud para Docker
            Write-Host "   Verificando configuración de Docker para GCR..."
            $gcrConfig = Get-Content "$env:USERPROFILE\.docker\config.json" -ErrorAction SilentlyContinue | ConvertFrom-Json
            if ($gcrConfig.credHelpers."gcr.io" -eq "gcloud") {
                Write-Host "✅ Docker configurado para usar GCR" -ForegroundColor Green
            } else {
                Write-Host "⚠️  Docker no está configurado para GCR" -ForegroundColor Yellow
                Write-Host "   Ejecuta: gcloud auth configure-docker"
            }
        } else {
            Write-Host "⚠️  Docker no está ejecutándose" -ForegroundColor Yellow
            Write-Host "   Inicia Docker Desktop o el servicio de Docker"
        }
    }
} catch {
    Write-Host "⚠️  Docker no está instalado" -ForegroundColor Yellow
    Write-Host "   No es crítico para el Release Pipeline (se ejecuta en GitHub Actions)"
}

# Verificar si el repositorio existe
Write-Host ""
Write-Host "📦 Verificando repositorio en GCR..." -ForegroundColor Yellow
$repoName = "asam-backend"
$repoList = gcloud container images list --repository="gcr.io/$currentProject" --filter="name:$repoName" 2>$null

if ($repoList -match $repoName) {
    Write-Host "✅ El repositorio existe en GCR" -ForegroundColor Green
    
    # Mostrar las últimas imágenes
    Write-Host ""
    Write-Host "📋 Últimas imágenes disponibles:" -ForegroundColor Cyan
    try {
        gcloud container images list-tags "gcr.io/$currentProject/$repoName" --limit=5 --format="table(tags,timestamp)"
    } catch {
        Write-Host "   No hay imágenes aún"
    }
} else {
    Write-Host "⚠️  El repositorio NO existe en GCR" -ForegroundColor Yellow
    Write-Host "   Se creará automáticamente en el primer release"
}

# Resumen
Write-Host ""
Write-Host "=== RESUMEN ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Si todos los checks están en ✅, estás listo para ejecutar el Release Pipeline." -ForegroundColor Green
Write-Host ""
Write-Host "Si hay algún ❌, sigue las instrucciones para corregirlo." -ForegroundColor Yellow
Write-Host ""
Write-Host "Para problemas de permisos de GCR, ejecuta:" -ForegroundColor Cyan
Write-Host "  .\scripts\gcp\fix-gcr-permissions.ps1 $currentProject"
Write-Host ""

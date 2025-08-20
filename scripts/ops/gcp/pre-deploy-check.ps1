# Script de diagnóstico para verificar la configuración antes de ejecutar el workflow
# Este script ayuda a identificar problemas comunes antes de ejecutar el despliegue

param(
    [Parameter(Mandatory=$false)]
    [string]$ProjectId = $env:GCP_PROJECT_ID
)

if (-not $ProjectId) {
    Write-Error "Error: Debes proporcionar el PROJECT_ID como parámetro o establecer GCP_PROJECT_ID"
    Write-Host "Uso: .\pre-deploy-check.ps1 -ProjectId <PROJECT_ID>"
    exit 1
}

Write-Host "=== Pre-Deploy Check para ASAM Backend ===" -ForegroundColor Cyan
Write-Host "Proyecto: $ProjectId" -ForegroundColor Green
Write-Host ""

$hasErrors = $false

# Función para marcar errores
function Mark-Error {
    $script:hasErrors = $true
}

# 1. Verificar Google Cloud SDK
Write-Host "1. Verificando Google Cloud SDK..." -ForegroundColor Yellow
$gcloudVersion = gcloud version --format="value(Google Cloud SDK)" 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Google Cloud SDK instalado: $gcloudVersion" -ForegroundColor Green
} else {
    Write-Host "   ❌ Google Cloud SDK no está instalado o no está en el PATH" -ForegroundColor Red
    Write-Host "   Instálalo desde: https://cloud.google.com/sdk/docs/install" -ForegroundColor Cyan
    Mark-Error
}

# 2. Verificar autenticación
Write-Host ""
Write-Host "2. Verificando autenticación de Google Cloud..." -ForegroundColor Yellow
$account = gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>$null
if ($account) {
    Write-Host "   ✅ Autenticado como: $account" -ForegroundColor Green
} else {
    Write-Host "   ❌ No estás autenticado en Google Cloud" -ForegroundColor Red
    Write-Host "   Ejecuta: gcloud auth login" -ForegroundColor Cyan
    Mark-Error
}

# 3. Verificar proyecto configurado
Write-Host ""
Write-Host "3. Verificando proyecto configurado..." -ForegroundColor Yellow
gcloud config set project $ProjectId 2>$null
$currentProject = gcloud config get-value project 2>$null
if ($currentProject -eq $ProjectId) {
    Write-Host "   ✅ Proyecto configurado correctamente: $currentProject" -ForegroundColor Green
} else {
    Write-Host "   ❌ El proyecto no está configurado correctamente" -ForegroundColor Red
    Write-Host "   Ejecuta: gcloud config set project $ProjectId" -ForegroundColor Cyan
    Mark-Error
}

# 4. Verificar APIs habilitadas
Write-Host ""
Write-Host "4. Verificando APIs habilitadas..." -ForegroundColor Yellow
$requiredApis = @(
    @{Name="Cloud Run API"; Service="run.googleapis.com"},
    @{Name="Cloud Build API"; Service="cloudbuild.googleapis.com"},
    @{Name="Container Registry API"; Service="containerregistry.googleapis.com"},
    @{Name="Secret Manager API"; Service="secretmanager.googleapis.com"}
)

foreach ($api in $requiredApis) {
    $enabled = gcloud services list --enabled --filter="name:$($api.Service)" --format="value(name)" 2>$null
    if ($enabled) {
        Write-Host "   ✅ $($api.Name) habilitada" -ForegroundColor Green
    } else {
        Write-Host "   ❌ $($api.Name) NO está habilitada" -ForegroundColor Red
        Write-Host "      Ejecuta: gcloud services enable $($api.Service)" -ForegroundColor Cyan
        Mark-Error
    }
}

# 5. Verificar cuenta de servicio
Write-Host ""
Write-Host "5. Verificando cuenta de servicio para GitHub Actions..." -ForegroundColor Yellow
$saEmail = "github-actions-deploy@${ProjectId}.iam.gserviceaccount.com"
$saExists = gcloud iam service-accounts describe $saEmail 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Cuenta de servicio existe: $saEmail" -ForegroundColor Green
    
    # Verificar roles
    $requiredRoles = @(
        "roles/run.admin",
        "roles/cloudbuild.builds.builder",
        "roles/iam.serviceAccountUser",
        "roles/storage.admin",
        "roles/secretmanager.secretAccessor"
    )
    
    $roles = gcloud projects get-iam-policy $ProjectId `
        --flatten="bindings[].members" `
        --filter="bindings.members:serviceAccount:$saEmail" `
        --format="value(bindings.role)" 2>$null
    
    Write-Host "   Verificando roles:" -ForegroundColor Yellow
    foreach ($role in $requiredRoles) {
        if ($roles -contains $role) {
            Write-Host "      ✅ $role" -ForegroundColor Green
        } else {
            Write-Host "      ❌ $role faltante" -ForegroundColor Red
            Mark-Error
        }
    }
} else {
    Write-Host "   ❌ Cuenta de servicio NO existe" -ForegroundColor Red
    Write-Host "   Crea la cuenta de servicio y configura los permisos según la documentación" -ForegroundColor Cyan
    Mark-Error
}

# 6. Verificar secretos de base de datos
Write-Host ""
Write-Host "6. Verificando secretos de base de datos..." -ForegroundColor Yellow
$dbSecrets = @("db-host", "db-port", "db-user", "db-password", "db-name")
$allSecretsOk = $true

foreach ($secret in $dbSecrets) {
    $exists = gcloud secrets describe $secret 2>$null
    if ($LASTEXITCODE -eq 0) {
        $versions = gcloud secrets versions list $secret --limit=1 --format="value(name)" 2>$null
        if ($versions) {
            Write-Host "   ✅ $secret configurado" -ForegroundColor Green
        } else {
            Write-Host "   ❌ $secret existe pero no tiene versiones" -ForegroundColor Red
            $allSecretsOk = $false
            Mark-Error
        }
    } else {
        Write-Host "   ❌ $secret NO existe" -ForegroundColor Red
        $allSecretsOk = $false
        Mark-Error
    }
}

if (-not $allSecretsOk) {
    Write-Host ""
    Write-Host "   Para configurar los secretos, ejecuta:" -ForegroundColor Cyan
    Write-Host "   .\verify-db-secrets.ps1 -ProjectId $ProjectId -CreateSecrets" -ForegroundColor White
}

# 7. Verificar otros secretos necesarios
Write-Host ""
Write-Host "7. Verificando otros secretos necesarios..." -ForegroundColor Yellow
$otherSecrets = @(
    "jwt-access-secret",
    "jwt-refresh-secret",
    "admin-user",
    "admin-password"
)

foreach ($secret in $otherSecrets) {
    $exists = gcloud secrets describe $secret 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ $secret configurado" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️ $secret NO existe (opcional pero recomendado)" -ForegroundColor Yellow
    }
}

# 8. Verificar imagen Docker más reciente
Write-Host ""
Write-Host "8. Verificando imágenes Docker disponibles..." -ForegroundColor Yellow
$images = gcloud container images list-tags "gcr.io/$ProjectId/asam-backend" --limit=5 --format="table(tags,timestamp)" 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Repositorio de imágenes encontrado" -ForegroundColor Green
    if ($images) {
        Write-Host ""
        Write-Host "   Últimas imágenes disponibles:" -ForegroundColor Cyan
        Write-Host $images
    } else {
        Write-Host "   ⚠️ No hay imágenes en el repositorio aún" -ForegroundColor Yellow
        Write-Host "   Las imágenes se crean automáticamente cuando se hace push a main" -ForegroundColor Cyan
    }
} else {
    Write-Host "   ⚠️ Repositorio de imágenes no encontrado o sin acceso" -ForegroundColor Yellow
    Write-Host "   Se creará automáticamente en el primer despliegue" -ForegroundColor Cyan
}

# Resumen final
Write-Host ""
Write-Host "=== Resumen ===" -ForegroundColor Cyan
Write-Host ""

if ($hasErrors) {
    Write-Host "❌ Se encontraron errores que deben solucionarse antes del despliegue" -ForegroundColor Red
    Write-Host ""
    Write-Host "Revisa los errores anteriores y sigue las instrucciones para solucionarlos." -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "✅ Todo está listo para el despliegue!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Próximos pasos:" -ForegroundColor Cyan
    Write-Host "1. Asegúrate de que los secretos en GitHub estén configurados" -ForegroundColor White
    Write-Host "2. Ve a GitHub Actions en tu repositorio" -ForegroundColor White
    Write-Host "3. Ejecuta el workflow 'Deploy to Google Cloud Run'" -ForegroundColor White
    Write-Host "4. Selecciona 'Run database migrations' si necesitas ejecutar migraciones" -ForegroundColor White
}

Write-Host ""

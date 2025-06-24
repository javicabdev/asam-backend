# Script para configurar Secret Manager y permisos para ASAM Backend

param(
    [string]$ProjectId = $(gcloud config get-value project),
    [switch]$CreateSecrets,
    [switch]$GrantPermissions
)

$ServiceAccount = "$ProjectId-compute@developer.gserviceaccount.com"

Write-Host "=== Configuración de Secret Manager para ASAM Backend ===" -ForegroundColor Cyan
Write-Host "Project ID: $ProjectId" -ForegroundColor Gray
Write-Host "Service Account: $ServiceAccount" -ForegroundColor Gray
Write-Host ""

# Lista de secretos requeridos
$secrets = @(
    @{name="db-host"; description="Database host"},
    @{name="db-port"; description="Database port"; value="5432"},
    @{name="db-user"; description="Database user"},
    @{name="db-password"; description="Database password"},
    @{name="db-name"; description="Database name"},
    @{name="jwt-access-secret"; description="JWT access token secret"},
    @{name="jwt-refresh-secret"; description="JWT refresh token secret"},
    @{name="admin-user"; description="Admin username"},
    @{name="admin-password"; description="Admin password"},
    @{name="smtp-user"; description="SMTP username (optional)"},
    @{name="smtp-password"; description="SMTP password (optional)"}
)

# Habilitar Secret Manager API
Write-Host "1. Habilitando Secret Manager API..." -ForegroundColor Yellow
gcloud services enable secretmanager.googleapis.com --project=$ProjectId

if ($CreateSecrets) {
    Write-Host ""
    Write-Host "2. Creando secretos..." -ForegroundColor Yellow
    
    foreach ($secret in $secrets) {
        Write-Host "   Creando secreto: $($secret.name)" -ForegroundColor Gray
        
        # Verificar si el secreto ya existe
        $exists = gcloud secrets describe $secret.name --project=$ProjectId 2>$null
        
        if ($exists) {
            Write-Host "   ✓ El secreto $($secret.name) ya existe" -ForegroundColor Green
        } else {
            # Crear el secreto
            gcloud secrets create $secret.name `
                --replication-policy="automatic" `
                --project=$ProjectId
            
            # Si tiene un valor por defecto, añadirlo
            if ($secret.value) {
                echo $secret.value | gcloud secrets versions add $secret.name `
                    --data-file=- `
                    --project=$ProjectId
                Write-Host "   ✓ Secreto $($secret.name) creado con valor por defecto" -ForegroundColor Green
            } else {
                Write-Host "   ✓ Secreto $($secret.name) creado (sin valor)" -ForegroundColor Yellow
                Write-Host "     Necesitas añadir el valor con:" -ForegroundColor DarkGray
                Write-Host "     echo -n 'TU_VALOR' | gcloud secrets versions add $($secret.name) --data-file=-" -ForegroundColor DarkGray
            }
        }
    }
}

if ($GrantPermissions) {
    Write-Host ""
    Write-Host "3. Otorgando permisos a la cuenta de servicio..." -ForegroundColor Yellow
    
    foreach ($secret in $secrets) {
        Write-Host "   Otorgando acceso a: $($secret.name)" -ForegroundColor Gray
        
        gcloud secrets add-iam-policy-binding $secret.name `
            --member="serviceAccount:$ServiceAccount" `
            --role="roles/secretmanager.secretAccessor" `
            --project=$ProjectId `
            --quiet
    }
    
    Write-Host ""
    Write-Host "✓ Permisos otorgados exitosamente" -ForegroundColor Green
}

# Mostrar instrucciones
Write-Host ""
Write-Host "=== Instrucciones ===" -ForegroundColor Cyan

if (-not $CreateSecrets -and -not $GrantPermissions) {
    Write-Host "No se especificó ninguna acción. Usa:" -ForegroundColor Yellow
    Write-Host "  -CreateSecrets     Para crear los secretos" -ForegroundColor Gray
    Write-Host "  -GrantPermissions  Para otorgar permisos" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Ejemplo:" -ForegroundColor Yellow
    Write-Host "  .\setup-secret-manager.ps1 -CreateSecrets -GrantPermissions" -ForegroundColor Gray
}

# Listar secretos que necesitan valores
Write-Host ""
Write-Host "=== Secretos que necesitan valores ===" -ForegroundColor Yellow
Write-Host ""
Write-Host "Ejecuta estos comandos con los valores reales:" -ForegroundColor Cyan
Write-Host ""

$secretsNeedingValues = @(
    @{name="db-host"; example="tu-host.aivencloud.com"},
    @{name="db-user"; example="avnadmin"},
    @{name="db-password"; example="tu-password-seguro"},
    @{name="db-name"; example="defaultdb"},
    @{name="jwt-access-secret"; example="tu-jwt-access-secret-aleatorio"},
    @{name="jwt-refresh-secret"; example="tu-jwt-refresh-secret-aleatorio"},
    @{name="admin-user"; example="admin"},
    @{name="admin-password"; example="password-admin-seguro"}
)

foreach ($secret in $secretsNeedingValues) {
    Write-Host "echo -n '$($secret.example)' | gcloud secrets versions add $($secret.name) --data-file=-" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Para SMTP (opcional):" -ForegroundColor Yellow
Write-Host "echo -n 'tu-usuario-smtp' | gcloud secrets versions add smtp-user --data-file=-" -ForegroundColor Gray
Write-Host "echo -n 'tu-password-smtp' | gcloud secrets versions add smtp-password --data-file=-" -ForegroundColor Gray

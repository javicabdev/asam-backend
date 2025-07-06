# Script para verificar y crear secretos de base de datos en Google Secret Manager
# Este script ayuda a configurar los secretos necesarios para las migraciones

param(
    [Parameter(Mandatory=$false)]
    [string]$ProjectId = $env:GCP_PROJECT_ID,
    
    [Parameter(Mandatory=$false)]
    [switch]$CreateSecrets,
    
    [Parameter(Mandatory=$false)]
    [switch]$TestConnection
)

if (-not $ProjectId) {
    Write-Error "Error: Debes proporcionar el PROJECT_ID como parámetro o establecer GCP_PROJECT_ID"
    Write-Host "Uso: .\verify-db-secrets.ps1 -ProjectId <PROJECT_ID> [-CreateSecrets] [-TestConnection]"
    exit 1
}

Write-Host "=== Verificando secretos de base de datos en Google Secret Manager ===" -ForegroundColor Cyan
Write-Host "Proyecto: $ProjectId" -ForegroundColor Green
Write-Host ""

# Configurar el proyecto
gcloud config set project $ProjectId

# Lista de secretos requeridos
$requiredSecrets = @(
    "db-host",
    "db-port", 
    "db-user",
    "db-password",
    "db-name"
)

$missingSecrets = @()
$secretValues = @{}

Write-Host "Verificando secretos existentes..." -ForegroundColor Yellow
Write-Host ""

foreach ($secret in $requiredSecrets) {
    try {
        $exists = gcloud secrets describe $secret 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ $secret existe" -ForegroundColor Green
            
            # Verificar si tiene versiones
            $versions = gcloud secrets versions list $secret --limit=1 --format="value(name)" 2>$null
            if ($versions) {
                Write-Host "   └─ Tiene versiones activas" -ForegroundColor DarkGray
                
                # Si vamos a hacer test de conexión, obtener el valor
                if ($TestConnection) {
                    $value = gcloud secrets versions access latest --secret=$secret 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        $secretValues[$secret] = $value
                    }
                }
            } else {
                Write-Host "   └─ ⚠️ No tiene versiones" -ForegroundColor Yellow
                $missingSecrets += $secret
            }
        } else {
            Write-Host "❌ $secret NO existe" -ForegroundColor Red
            $missingSecrets += $secret
        }
    } catch {
        Write-Host "❌ $secret NO existe" -ForegroundColor Red
        $missingSecrets += $secret
    }
}

Write-Host ""

# Verificar permisos de la cuenta de servicio
Write-Host "Verificando permisos de la cuenta de servicio..." -ForegroundColor Yellow
$saEmail = "github-actions-deploy@${ProjectId}.iam.gserviceaccount.com"

# Verificar si la cuenta de servicio existe
$saExists = gcloud iam service-accounts describe $saEmail 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️ La cuenta de servicio $saEmail no existe" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Para crear la cuenta de servicio, ejecuta:" -ForegroundColor Cyan
    Write-Host "gcloud iam service-accounts create github-actions-deploy --display-name='GitHub Actions Deploy Service Account'" -ForegroundColor White
} else {
    # Verificar si tiene el rol secretmanager.secretAccessor
    $roles = gcloud projects get-iam-policy $ProjectId `
        --flatten="bindings[].members" `
        --filter="bindings.members:serviceAccount:$saEmail" `
        --format="value(bindings.role)" 2>$null
    
    if ($roles -match "roles/secretmanager.secretAccessor") {
        Write-Host "✅ La cuenta de servicio tiene acceso a los secretos" -ForegroundColor Green
    } else {
        Write-Host "⚠️ La cuenta de servicio NO tiene el rol secretmanager.secretAccessor" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Para agregar el rol, ejecuta:" -ForegroundColor Cyan
        Write-Host "gcloud projects add-iam-policy-binding $ProjectId --member='serviceAccount:$saEmail' --role='roles/secretmanager.secretAccessor'" -ForegroundColor White
    }
}

Write-Host ""

if ($missingSecrets.Count -eq 0) {
    Write-Host "✅ Todos los secretos están configurados!" -ForegroundColor Green
    
    if ($TestConnection -and $secretValues.Count -eq 5) {
        Write-Host ""
        Write-Host "Probando conexión a la base de datos..." -ForegroundColor Yellow
        
        # Construir la cadena de conexión
        $host = $secretValues["db-host"]
        $port = $secretValues["db-port"]
        $user = $secretValues["db-user"]
        $password = $secretValues["db-password"]
        $dbname = $secretValues["db-name"]
        
        Write-Host ("Host: {0}:{1}" -f $host, $port) -ForegroundColor DarkGray
        Write-Host "Database: $dbname" -ForegroundColor DarkGray
        Write-Host "User: $user" -ForegroundColor DarkGray
        
        # Usar psql si está disponible
        $psqlPath = Get-Command psql -ErrorAction SilentlyContinue
        if ($psqlPath) {
            $env:PGPASSWORD = $password
            $result = & psql -h $host -p $port -U $user -d $dbname -c 'SELECT version();' 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Conexión exitosa!" -ForegroundColor Green
                Write-Host $result -ForegroundColor DarkGray
            } else {
                Write-Host "❌ Error al conectar:" -ForegroundColor Red
                Write-Host $result -ForegroundColor Red
            }
            Remove-Item Env:\PGPASSWORD
        } else {
            Write-Host "⚠️ psql no está instalado. No se puede probar la conexión." -ForegroundColor Yellow
            Write-Host "Para instalar psql:" -ForegroundColor Cyan
            Write-Host "  - Windows: choco install postgresql" -ForegroundColor White
            Write-Host "  - Linux: sudo apt-get install postgresql-client" -ForegroundColor White
            Write-Host "  - Mac: brew install postgresql" -ForegroundColor White
        }
    }
    
} else {
    Write-Host "❌ Faltan los siguientes secretos:" -ForegroundColor Red
    $missingSecrets | ForEach-Object { Write-Host "   - $_" -ForegroundColor Red }
    
    if ($CreateSecrets) {
        Write-Host ""
        Write-Host "Creando secretos faltantes..." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Por favor, ingresa los valores para cada secreto:" -ForegroundColor Cyan
        
        foreach ($secret in $missingSecrets) {
            Write-Host ""
            switch ($secret) {
                "db-host" {
                    $value = Read-Host "DB Host (ej: pg-xxx.aivencloud.com)"
                }
                "db-port" {
                    $value = Read-Host "DB Port (ej: 14276)"
                }
                "db-user" {
                    $value = Read-Host "DB User (ej: avnadmin)"
                }
                "db-password" {
                    $securePassword = Read-Host "DB Password" -AsSecureString
                    $value = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
                        [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
                    )
                }
                "db-name" {
                    $value = Read-Host "DB Name (ej: defaultdb)"
                }
            }
            
            if ($value) {
                Write-Host "Creando secreto $secret..." -ForegroundColor Yellow
                
                # Verificar si el secreto existe pero no tiene versiones
                $exists = gcloud secrets describe $secret 2>$null
                if ($LASTEXITCODE -eq 0) {
                    # El secreto existe, solo agregar una versión
                    echo $value | gcloud secrets versions add $secret --data-file=-
                } else {
                    # Crear el secreto y agregar la versión
                    echo $value | gcloud secrets create $secret --data-file=-
                }
                
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "✅ $secret creado exitosamente" -ForegroundColor Green
                } else {
                    Write-Host "❌ Error al crear $secret" -ForegroundColor Red
                }
            }
        }
        
        Write-Host ""
        Write-Host "✅ Proceso completado!" -ForegroundColor Green
        
        # Otorgar permisos a la cuenta de servicio
        Write-Host ""
        Write-Host "Otorgando permisos a la cuenta de servicio..." -ForegroundColor Yellow
        $saEmail = "github-actions-deploy@${ProjectId}.iam.gserviceaccount.com"
        
        gcloud projects add-iam-policy-binding $ProjectId `
            --member="serviceAccount:$saEmail" `
            --role="roles/secretmanager.secretAccessor" `
            --quiet
            
        Write-Host "✅ Permisos otorgados" -ForegroundColor Green
        
    } else {
        Write-Host ""
        Write-Host "Para crear los secretos, ejecuta:" -ForegroundColor Cyan
        Write-Host ".\verify-db-secrets.ps1 -ProjectId $ProjectId -CreateSecrets" -ForegroundColor White
    }
}

Write-Host ""
Write-Host "=== Resumen ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Si necesitas actualizar un secreto existente:" -ForegroundColor Yellow
Write-Host 'echo "nuevo-valor" | gcloud secrets versions add <secret-name> --data-file=-' -ForegroundColor White
Write-Host ""
Write-Host "Para ver el valor actual de un secreto:" -ForegroundColor Yellow
Write-Host "gcloud secrets versions access latest --secret=<secret-name>" -ForegroundColor White
Write-Host ""
Write-Host "Para probar la conexión a la base de datos:" -ForegroundColor Yellow
Write-Host (".\verify-db-secrets.ps1 -ProjectId {0} -TestConnection" -f $ProjectId) -ForegroundColor White
Write-Host ""

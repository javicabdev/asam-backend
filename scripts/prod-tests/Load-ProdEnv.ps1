# Script para cargar variables de entorno de producción y ejecutar pruebas

param(
    [switch]$CheckOnly,
    [switch]$QuickTest,
    [switch]$FullTest
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

# Change to project root
$projectRoot = (Get-Item $PSScriptRoot).Parent.Parent.FullName
Set-Location $projectRoot

Write-Info "=========================================="
Write-Info "Cargando entorno de producción"
Write-Info "=========================================="
Write-Host ""

# Read .env.production file
$envFile = ".env.production"
if (!(Test-Path $envFile)) {
    Write-Error "ERROR: No se encontró el archivo $envFile"
    exit 1
}

Write-Info "Leyendo archivo: $envFile"

# Parse and set environment variables
$envContent = Get-Content $envFile
foreach ($line in $envContent) {
    # Skip comments and empty lines
    if ($line -match "^\s*#" -or $line -match "^\s*$") {
        continue
    }
    
    # Parse KEY=VALUE
    if ($line -match "^([^=]+)=(.*)$") {
        $key = $Matches[1].Trim()
        $value = $Matches[2].Trim()
        
        # Remove quotes if present
        if ($value.StartsWith('"') -and $value.EndsWith('"')) {
            $value = $value.Substring(1, $value.Length - 2)
        } elseif ($value.StartsWith("'") -and $value.EndsWith("'")) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        
        # Skip if value contains ${} placeholders
        if (-not $value.Contains('${')) {
            [Environment]::SetEnvironmentVariable($key, $value, [EnvironmentVariableTarget]::Process)
            
            # Display loaded variables (hide sensitive data)
            if ($key -like "*PASSWORD*" -or $key -like "*SECRET*") {
                Write-Host "  ✓ $key = ******" -ForegroundColor DarkGray
            } else {
                Write-Host "  ✓ $key = $value" -ForegroundColor DarkGray
            }
        }
    }
}

# Force production environment
[Environment]::SetEnvironmentVariable("APP_ENV", "production", [EnvironmentVariableTarget]::Process)
[Environment]::SetEnvironmentVariable("ENVIRONMENT", "production", [EnvironmentVariableTarget]::Process)

Write-Host ""
Write-Success "Variables de entorno cargadas exitosamente"

# Verify database connection info
$dbHost = [Environment]::GetEnvironmentVariable("DB_HOST")
$dbPort = [Environment]::GetEnvironmentVariable("DB_PORT")
$dbName = [Environment]::GetEnvironmentVariable("DB_NAME")

Write-Host ""
Write-Info "Configuración de base de datos:"
Write-Host "  Host: $dbHost" -ForegroundColor DarkGray
Write-Host "  Puerto: $dbPort" -ForegroundColor DarkGray
Write-Host "  Base de datos: $dbName" -ForegroundColor DarkGray

if ($dbHost -eq "pg-asam-asam-backend-db.l.aivencloud.com") {
    Write-Success "  ✓ Usando base de datos de PRODUCCIÓN en Aiven"
} else {
    Write-Warning "  ⚠ No estás usando la base de datos de producción"
}

# Execute requested action
if ($CheckOnly) {
    Write-Host ""
    Write-Info "Ejecutando verificación de entorno..."
    & go run scripts/prod-tests/check_env.go
} elseif ($QuickTest) {
    Write-Host ""
    Write-Info "Ejecutando prueba rápida de conexión..."
    & go run scripts/prod-tests/quick_connection_check.go
} elseif ($FullTest) {
    Write-Host ""
    Write-Info "Ejecutando pruebas completas de CRUD..."
    & go run scripts/prod-tests/test_database_operations.go
} else {
    Write-Host ""
    Write-Info "Opciones disponibles:"
    Write-Host "  -CheckOnly  : Solo verificar variables de entorno" -ForegroundColor DarkGray
    Write-Host "  -QuickTest  : Prueba rápida de conexión" -ForegroundColor DarkGray
    Write-Host "  -FullTest   : Pruebas completas de CRUD" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Ejemplo: .\Load-ProdEnv.ps1 -QuickTest" -ForegroundColor Yellow
}

Write-Host ""

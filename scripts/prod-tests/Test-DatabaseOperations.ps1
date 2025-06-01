# Script de pruebas de base de datos para producción
# Este script verifica las operaciones CRUD en la base de datos

param(
    [switch]$SkipConnectionTest,
    [switch]$Verbose,
    [switch]$UseLocalEnv
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

# Header
Write-Host ""
Write-Info "=========================================="
Write-Info "Pruebas de Base de Datos - Producción"
Write-Info "=========================================="
Write-Host ""

# Change to project root
$projectRoot = (Get-Item $PSScriptRoot).Parent.Parent.FullName
Set-Location $projectRoot
Write-Info "Directorio del proyecto: $projectRoot"

# Check if .env.production exists
if ($UseLocalEnv) {
    $envFile = ".env.development"
    Write-Warning "Usando archivo de entorno local: $envFile"
} else {
    $envFile = ".env.production"
    Write-Info "Usando archivo de entorno: $envFile"
}

if (!(Test-Path $envFile)) {
    Write-Error "ERROR: No se encontró el archivo $envFile"
    exit 1
}

# Load environment variables
Write-Info "Cargando variables de entorno..."
$envContent = Get-Content $envFile | Where-Object { $_ -match "^[^#].*=" }
foreach ($line in $envContent) {
    $parts = $line -split "=", 2
    if ($parts.Count -eq 2) {
        $name = $parts[0].Trim()
        $value = $parts[1].Trim()
        if ($Verbose) {
            if ($name -like "*PASSWORD*" -or $name -like "*SECRET*") {
                Write-Host "  $name=***" -ForegroundColor DarkGray
            } else {
                Write-Host "  $name=$value" -ForegroundColor DarkGray
            }
        }
        [Environment]::SetEnvironmentVariable($name, $value, [EnvironmentVariableTarget]::Process)
    }
}

# Test database connection first
if (!$SkipConnectionTest) {
    Write-Host ""
    Write-Info "Probando conexión a la base de datos..."
    
    $dbHost = [Environment]::GetEnvironmentVariable("DB_HOST")
    $dbPort = [Environment]::GetEnvironmentVariable("DB_PORT")
    $dbName = [Environment]::GetEnvironmentVariable("DB_NAME")
    
    Write-Host "  Host: $dbHost" -ForegroundColor DarkGray
    Write-Host "  Puerto: $dbPort" -ForegroundColor DarkGray
    Write-Host "  Base de datos: $dbName" -ForegroundColor DarkGray
    
    # Test network connectivity
    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        $tcpClient.Connect($dbHost, $dbPort)
        $tcpClient.Close()
        Write-Success "  ✓ Conexión de red exitosa"
    } catch {
        Write-Error "  ✗ No se pudo conectar al servidor de base de datos"
        Write-Error "    Error: $_"
        exit 1
    }
}

# Download dependencies
Write-Host ""
Write-Info "Descargando dependencias..."
$goModResult = & go mod download 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Error "Error descargando dependencias:"
    Write-Error $goModResult
    exit 1
}
Write-Success "✓ Dependencias descargadas"

# Run tests
Write-Host ""
Write-Info "Ejecutando pruebas de operaciones CRUD..."
Write-Host ""

# Force production environment variables
[Environment]::SetEnvironmentVariable("APP_ENV", "production", [EnvironmentVariableTarget]::Process)
[Environment]::SetEnvironmentVariable("ENVIRONMENT", "production", [EnvironmentVariableTarget]::Process)

# Run the test script
$testResult = & go run scripts/prod-tests/test_database_operations.go 2>&1 | Out-String

# Display test output
Write-Host $testResult

# Check if tests passed
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Success "=========================================="
    Write-Success "✓ PRUEBAS COMPLETADAS EXITOSAMENTE"
    Write-Success "=========================================="
    
    # Additional info
    Write-Host ""
    Write-Info "Información adicional:"
    Write-Host "  - Las pruebas crearon y eliminaron datos temporales" -ForegroundColor DarkGray
    Write-Host "  - No quedaron datos de prueba en la base de datos" -ForegroundColor DarkGray
    Write-Host "  - La base de datos está lista para operaciones normales" -ForegroundColor DarkGray
} else {
    Write-Host ""
    Write-Error "=========================================="
    Write-Error "✗ ALGUNAS PRUEBAS FALLARON"
    Write-Error "=========================================="
    Write-Host ""
    Write-Warning "Por favor revisa los errores arriba y:"
    Write-Host "  1. Verifica las credenciales de la base de datos" -ForegroundColor DarkGray
    Write-Host "  2. Asegúrate de que el servidor está accesible" -ForegroundColor DarkGray
    Write-Host "  3. Confirma que las migraciones están actualizadas" -ForegroundColor DarkGray
    exit 1
}

Write-Host ""

# scripts/migrate.ps1

param (
    [Parameter(Mandatory = $false, Position = 0, HelpMessage = "Entorno para migraciones, por defecto 'development'")]
    [string]$Env = "development"
)

Write-Host "migrate.ps1 - Received Env: '$Env'" -ForegroundColor Cyan  # Depuración

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
Write-Host "migrate.ps1 - Project Root: '$projectRoot'" -ForegroundColor Cyan  # Depuración

# Establecer la variable de entorno APP_ENV
$env:APP_ENV = $Env
Write-Host "migrate.ps1 - Set APP_ENV to '$Env'" -ForegroundColor Cyan  # Depuración

# Determinar el nombre de la base de datos según el entorno
if ($Env -eq "test") {
    $dbName = "asam_db_test"
} else {
    $dbName = "asam_db"
}
Write-Host "migrate.ps1 - Database Name: '$dbName'" -ForegroundColor Cyan  # Depuración

# Verificar que psql está instalado
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "Error: psql no está instalado o no está en el PATH." -ForegroundColor Red
    exit 1
}

# [Debug] Verificar Variables de Entorno
Write-Host "migrate.ps1 - DB_HOST: '$env:DB_HOST'" -ForegroundColor Yellow
Write-Host "migrate.ps1 - DB_PORT: '$env:DB_PORT'" -ForegroundColor Yellow
Write-Host "migrate.ps1 - DB_USER: '$env:DB_USER'" -ForegroundColor Yellow
Write-Host "migrate.ps1 - DB_PASSWORD: '$env:DB_PASSWORD'" -ForegroundColor Yellow

# Establecer la variable PGPASSWORD para evitar la solicitud de contraseña interactiva
$env:PGPASSWORD = $env:DB_PASSWORD
Write-Host "migrate.ps1 - Set PGPASSWORD." -ForegroundColor Cyan  # Depuración

# Ejecutar migraciones localmente usando psql
Write-Host "Aplicando migraciones a la base de datos '$dbName'..." -ForegroundColor Green
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $dbName -f "$projectRoot/init.sql"

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error al aplicar las migraciones." -ForegroundColor Red
    exit 1
} else {
    Write-Host "Migraciones aplicadas correctamente a la base de datos '$dbName'." -ForegroundColor Green
}

# Limpiar la variable de entorno
Remove-Item Env:\APP_ENV -ErrorAction SilentlyContinue
Write-Host "migrate.ps1 - Limpieza de APP_ENV completada." -ForegroundColor Cyan  # Depuración

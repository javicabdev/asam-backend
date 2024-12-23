# scripts/run.ps1

param (
    [string]$Env = "development"
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot

# Ruta al directorio 'cmd/api'
$cmdApiPath = Join-Path $projectRoot "cmd\api"

# Verificar si el directorio existe
if (-Not (Test-Path $cmdApiPath)) {
    Write-Host "Error: El directorio '$cmdApiPath' no existe." -ForegroundColor Red
    exit 1
}

# Establecer la variable de entorno APP_ENV
$env:APP_ENV = $Env

# Ejecutar la aplicación Go
Write-Host "Ejecutando la aplicación en el entorno '$Env' desde '$cmdApiPath'..." -ForegroundColor Green
go run "$cmdApiPath"

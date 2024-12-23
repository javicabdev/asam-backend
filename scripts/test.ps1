# scripts/test.ps1

param (
    [string]$Env = "test"
)

Write-Host "test.ps1 - Received Env: '$Env'" -ForegroundColor Cyan  # Depuración

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
Write-Host "test.ps1 - Project Root: '$projectRoot'" -ForegroundColor Cyan  # Depuración

# Establecer la variable de entorno APP_ENV
$env:APP_ENV = $Env
Write-Host "test.ps1 - Set APP_ENV to '$Env'" -ForegroundColor Cyan  # Depuración

# Cargar las variables de entorno desde .env.test
Write-Host "Cargando configuración desde .env.$Env..." -ForegroundColor Green
if (Test-Path "$projectRoot\.env.$Env") {
    Get-Content "$projectRoot\.env.$Env" | ForEach-Object {
        $pair = $_ -split "="
        if ($pair.Length -eq 2) {
            [Environment]::SetEnvironmentVariable($pair[0], $pair[1], "Process")
            Write-Host "test.ps1 - Set variable '$($pair[0])' to '$($pair[1])'" -ForegroundColor Cyan  # Depuración
        }
    }
} else {
    Write-Host "Advertencia: El archivo .env.$Env no existe." -ForegroundColor Yellow
}

# Ejecutar migraciones para la base de datos de pruebas
Write-Host "Ejecutando migraciones para el entorno '$Env'..." -ForegroundColor Green
& "$PSScriptRoot\run-all.ps1" migrate $Env  # Ruta corregida

# Ejecutar los tests
Write-Host "Ejecutando tests en el entorno '$Env'..." -ForegroundColor Green
go test ./test/integration/... -v

# Limpiar la variable de entorno
Remove-Item Env:\APP_ENV -ErrorAction SilentlyContinue
Write-Host "test.ps1 - Limpieza de APP_ENV completada." -ForegroundColor Cyan  # Depuración

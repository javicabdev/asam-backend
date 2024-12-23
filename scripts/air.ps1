# scripts/air.ps1

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot

# Establecer la variable de entorno APP_ENV a development
$env:APP_ENV = "development"

# Ejecutar Air desde el directorio raíz del proyecto
Push-Location $projectRoot
air
Pop-Location

# Limpiar la variable de entorno
Remove-Item Env:\APP_ENV

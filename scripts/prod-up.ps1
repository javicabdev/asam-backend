# scripts/prod-up.ps1
param (
    [switch]$Build
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot

# Establecer el entorno a producción
$env:APP_ENV = "production"

# Construir y levantar los contenedores
if ($Build) {
    Write-Host "Construyendo y levantando contenedores de producción..." -ForegroundColor Green
    docker-compose -f "$projectRoot\docker-compose.prod.yml" up -d --build
} else {
    Write-Host "Levantando contenedores de producción..." -ForegroundColor Green
    docker-compose -f "$projectRoot\docker-compose.prod.yml" up -d
}

Write-Host "Entorno de producción iniciado" -ForegroundColor Green
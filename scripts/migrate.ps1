# scripts/migrate.ps1
param (
    [Parameter(Mandatory = $false, Position = 0)]
    [string]$Env = "development",
    [Parameter(Mandatory = $false, Position = 1)]
    [string]$Command = "up"
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot

# Cargar variables de entorno según el entorno
if (Test-Path "$projectRoot\.env.$Env") {
    Get-Content "$projectRoot\.env.$Env" | ForEach-Object {
        $pair = $_ -split "="
        if ($pair.Length -eq 2) {
            [Environment]::SetEnvironmentVariable($pair[0], $pair[1], "Process")
        }
    }
}

# Construir la URL de conexión
$connectionString = "postgresql://{0}:{1}@{2}:{3}/{4}?sslmode={5}" -f `
       $env:DB_USER, `
       $env:DB_PASSWORD, `
       $env:DB_HOST, `
       $env:DB_PORT, `
       $env:DB_NAME, `
       $env:DB_SSL_MODE

# Ejecutar migrate
Write-Host "Ejecutando migraciones para el entorno '$Env'..." -ForegroundColor Green
migrate -database $connectionString -path "migrations" $Command

if ($LASTEXITCODE -eq 0) {
    Write-Host "Migraciones completadas exitosamente." -ForegroundColor Green
} else {
    Write-Host "Error al ejecutar las migraciones." -ForegroundColor Red
    exit 1
}
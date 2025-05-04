# Script PowerShell para instalar dependencias y ejecutar migraciones
param (
    [Parameter(Position=0)]
    [string]$Environment = "local",
    
    [Parameter(Position=1)]
    [string]$Command = "up",
    
    [Parameter(Position=2, ValueFromRemainingArguments=$true)]
    [string[]]$ExtraArgs
)

# Instalar las dependencias necesarias
Write-Host "Instalando dependencias necesarias..." -ForegroundColor Cyan
$installProcess = Start-Process -FilePath "go" -ArgumentList @(
    "get", 
    "github.com/golang-migrate/migrate/v4",
    "github.com/golang-migrate/migrate/v4/database/postgres",
    "github.com/golang-migrate/migrate/v4/source/file"
) -Wait -NoNewWindow -PassThru

if ($installProcess.ExitCode -ne 0) {
    Write-Host "Error al instalar dependencias (código: $($installProcess.ExitCode))" -ForegroundColor Red
    exit $installProcess.ExitCode
}

# Ejecutar el comando tidy para resolver las dependencias
Write-Host "Ejecutando go mod tidy..." -ForegroundColor Cyan
$tidyProcess = Start-Process -FilePath "go" -ArgumentList @("mod", "tidy") -Wait -NoNewWindow -PassThru

if ($tidyProcess.ExitCode -ne 0) {
    Write-Host "Error al ejecutar go mod tidy (código: $($tidyProcess.ExitCode))" -ForegroundColor Red
    exit $tidyProcess.ExitCode
}

Write-Host "Dependencias instaladas correctamente" -ForegroundColor Green

# Ahora ejecutar el script de migración original
Write-Host "Ejecutando migraciones..." -ForegroundColor Cyan
& .\migrate.ps1 $Environment $Command $ExtraArgs

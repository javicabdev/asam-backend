# scripts/run-all.ps1

param (
    [string]$Command,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Args
)

# Agregar depuración para verificar los parámetros recibidos
Write-Host "run-all.ps1 - Command: '$Command'" -ForegroundColor Yellow
Write-Host "run-all.ps1 - Args: '$($Args -join ', ')" -ForegroundColor Yellow

# Verificar si se proporcionó un comando
if (-not $Command) {
    Write-Host "Uso: .\scripts\run-all.ps1 <comando> [args...]" -ForegroundColor Cyan
    Write-Host "Comandos disponibles:" -ForegroundColor Yellow
    Get-ChildItem -Path $PSScriptRoot -Filter *.ps1 | ForEach-Object { $_.BaseName } | Write-Host
    exit 0
}

# Obtener la ruta completa al script que se va a ejecutar
$scriptPath = Join-Path -Path $PSScriptRoot -ChildPath "$Command.ps1"

# Verificar si el script existe
if (Test-Path $scriptPath) {
    Write-Host "Ejecutando '$Command'..." -ForegroundColor Green
    if ($Args.Count -gt 0) {
        & $scriptPath @Args
    } else {
        & $scriptPath
    }
} else {
    Write-Host "Comando no encontrado: $Command" -ForegroundColor Red
    Write-Host "Comandos disponibles:" -ForegroundColor Yellow
    Get-ChildItem -Path $PSScriptRoot -Filter *.ps1 | ForEach-Object { $_.BaseName } | Write-Host
    exit 1
}

# scripts/clean.ps1
param (
    [switch]$All,                    # Limpieza completa incluyendo docker
    [switch]$RemoveVolumes,          # Eliminar volúmenes de Docker
    [int]$LogRetentionDays = 7,      # Días a mantener logs
    [int]$BackupRetentionDays = 7    # Días a mantener backups
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
Write-Host "Iniciando limpieza de recursos..." -ForegroundColor Green

# Limpiar directorio tmp
Write-Host "Limpiando directorio tmp..." -ForegroundColor Cyan
$tmpPath = Join-Path $projectRoot "tmp"
if (Test-Path $tmpPath) {
    Remove-Item -Path "$tmpPath\*" -Recurse -Force
    Write-Host "✅ Directorio tmp limpiado" -ForegroundColor Green
}

# Limpiar logs antiguos
Write-Host "Limpiando logs antiguos..." -ForegroundColor Cyan
$logsPath = Join-Path $projectRoot "logs"
if (Test-Path $logsPath) {
    Get-ChildItem -Path $logsPath -Filter "*.log" |
            Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$LogRetentionDays) } |
            ForEach-Object {
                Remove-Item $_.FullName -Force
                Write-Host "Eliminado log antiguo: $($_.Name)" -ForegroundColor Yellow
            }
    Write-Host "✅ Logs antiguos eliminados" -ForegroundColor Green
}

# Limpiar backups antiguos
Write-Host "Limpiando backups antiguos..." -ForegroundColor Cyan
$backupsPath = Join-Path $projectRoot "backups"
if (Test-Path $backupsPath) {
    Get-ChildItem -Path $backupsPath -Filter "asam_backup_*.sql" |
            Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$BackupRetentionDays) } |
            ForEach-Object {
                Remove-Item $_.FullName -Force
                Write-Host "Eliminado backup antiguo: $($_.Name)" -ForegroundColor Yellow
            }
    Write-Host "✅ Backups antiguos eliminados" -ForegroundColor Green
}

# Si se especifica -All, limpiar recursos de Docker
if ($All) {
    Write-Host "Limpiando recursos de Docker..." -ForegroundColor Cyan

    # Detener y eliminar contenedores
    & "$PSScriptRoot\docker-down.ps1"

    # Eliminar imágenes no utilizadas
    docker image prune -f
    Write-Host "✅ Imágenes no utilizadas eliminadas" -ForegroundColor Green

    # Eliminar volúmenes si se especifica
    if ($RemoveVolumes) {
        docker volume rm $(docker volume ls -q -f dangling=true)
        Write-Host "✅ Volúmenes eliminados" -ForegroundColor Green
    }
}

Write-Host "`nLimpieza completada! 🧹" -ForegroundColor Green

# Mostrar resumen del espacio liberado
if ($IsWindows) {
    $beforeSize = (Get-ChildItem $projectRoot -Recurse | Measure-Object -Property Length -Sum).Sum
    Write-Host "Espacio total del proyecto: $([math]::Round($beforeSize/1MB, 2)) MB" -ForegroundColor Cyan
}
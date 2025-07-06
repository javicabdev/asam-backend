# scripts/logs.ps1
param (
    [Parameter(Mandatory = $false)]
    [ValidateSet('docker', 'app', 'all')]
    [string]$Type = 'all',

    [Parameter(Mandatory = $false)]
    [ValidateSet('error', 'warn', 'info', 'debug', 'all')]
    [string]$Level = 'all',

    [Parameter(Mandatory = $false)]
    [switch]$Follow,

    [Parameter(Mandatory = $false)]
    [int]$Tail = 100,

    [Parameter(Mandatory = $false)]
    [string]$Since = "1h"
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot

# Función para formatear la salida de logs
function Format-LogOutput {
    param (
        [Parameter(ValueFromPipeline = $true)]
        [string]$LogLine
    )
    process {
        switch -Regex ($LogLine) {
            'ERROR|error|Error' { Write-Host $LogLine -ForegroundColor Red }
            'WARN|warn|Warning' { Write-Host $LogLine -ForegroundColor Yellow }
            'INFO|info' { Write-Host $LogLine -ForegroundColor Green }
            'DEBUG|debug' { Write-Host $LogLine -ForegroundColor Gray }
            default { Write-Host $LogLine }
        }
    }
}

# Función para mostrar logs de Docker
function Show-DockerLogs {
    if ($Follow) {
        docker-compose logs --follow --tail=$Tail | Format-LogOutput
    }
    else {
        docker-compose logs --tail=$Tail | Format-LogOutput
    }
}

# Función para mostrar logs de la aplicación
function Show-AppLogs {
    $logsPath = Join-Path $projectRoot "logs"
    if (-not (Test-Path $logsPath)) {
        Write-Host "No se encuentra el directorio de logs" -ForegroundColor Red
        return
    }

    # Filtrar por nivel de log si se especifica
    $filter = "*.log"
    if ($Level -ne 'all') {
        $filter = "*.$Level.log"
    }

    # Obtener los archivos de log
    $logFiles = Get-ChildItem -Path $logsPath -Filter $filter |
            Sort-Object LastWriteTime -Descending

    if ($Follow) {
        # Mostrar los últimos logs y seguir los cambios del archivo más reciente
        if ($logFiles.Count -gt 0) {
            $latestLog = $logFiles[0]
            Write-Host "Siguiendo $($latestLog.Name)..." -ForegroundColor Cyan
            Get-Content -Path $latestLog.FullName -Tail $Tail -Wait | Format-LogOutput
        }
    }
    else {
        # Mostrar los últimos logs de todos los archivos
        foreach ($logFile in $logFiles) {
            Write-Host "`nLogs de $($logFile.Name):" -ForegroundColor Cyan
            Get-Content -Path $logFile.FullName -Tail $Tail | Format-LogOutput
        }
    }
}

# Función principal
switch ($Type) {
    'docker' {
        Write-Host "Mostrando logs de Docker..." -ForegroundColor Cyan
        Show-DockerLogs
    }
    'app' {
        Write-Host "Mostrando logs de la aplicación..." -ForegroundColor Cyan
        Show-AppLogs
    }
    'all' {
        Write-Host "Mostrando todos los logs..." -ForegroundColor Cyan
        Write-Host "`n=== Logs de Docker ===" -ForegroundColor Cyan
        Show-DockerLogs
        Write-Host "`n=== Logs de la Aplicación ===" -ForegroundColor Cyan
        Show-AppLogs
    }
}
# Script para crear tarea programada de mantenimiento de tokens en Windows
# Ejecutar como administrador

param(
    [Parameter(Mandatory=$false)]
    [string]$TaskName = "ASAM-TokenMaintenance",
    
    [Parameter(Mandatory=$false)]
    [string]$Description = "Limpieza automática de tokens expirados para ASAM Backend",
    
    [Parameter(Mandatory=$false)]
    [string]$WorkingDirectory = (Get-Location).Path,
    
    [Parameter(Mandatory=$false)]
    [string]$Time = "03:00",
    
    [Parameter(Mandatory=$false)]
    [switch]$Remove = $false
)

# Verificar permisos de administrador
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Este script requiere permisos de administrador." -ForegroundColor Red
    Write-Host "Por favor, ejecuta PowerShell como administrador e intenta de nuevo." -ForegroundColor Yellow
    exit 1
}

# Remover tarea si se solicita
if ($Remove) {
    try {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
        Write-Host "Tarea '$TaskName' eliminada exitosamente." -ForegroundColor Green
    } catch {
        Write-Host "Error al eliminar la tarea: $_" -ForegroundColor Red
    }
    exit
}

# Crear la acción de la tarea
$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$WorkingDirectory\scripts\maintenance.ps1`" -All -Report" `
    -WorkingDirectory $WorkingDirectory

# Crear el trigger (diario a las 3:00 AM por defecto)
$trigger = New-ScheduledTaskTrigger -Daily -At $Time

# Crear las configuraciones adicionales
$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -RunOnlyIfNetworkAvailable `
    -MultipleInstances IgnoreNew

# Crear la tarea
try {
    # Verificar si la tarea ya existe
    $existingTask = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
    
    if ($existingTask) {
        Write-Host "La tarea '$TaskName' ya existe. Actualizando..." -ForegroundColor Yellow
        Set-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Settings $settings
    } else {
        Write-Host "Creando nueva tarea '$TaskName'..." -ForegroundColor Cyan
        Register-ScheduledTask `
            -TaskName $TaskName `
            -Description $Description `
            -Action $action `
            -Trigger $trigger `
            -Settings $settings `
            -RunLevel Highest `
            -User "SYSTEM"
    }
    
    Write-Host @"

Tarea programada creada exitosamente!

Detalles:
- Nombre: $TaskName
- Ejecuta: Diariamente a las $Time
- Comando: maintenance.ps1 -All -Report
- Directorio: $WorkingDirectory

Para gestionar la tarea:
- Ver estado: Get-ScheduledTask -TaskName "$TaskName"
- Ejecutar ahora: Start-ScheduledTask -TaskName "$TaskName"
- Deshabilitar: Disable-ScheduledTask -TaskName "$TaskName"
- Habilitar: Enable-ScheduledTask -TaskName "$TaskName"
- Eliminar: .\create-scheduled-task.ps1 -Remove

Los logs se guardarán en: $WorkingDirectory\logs\maintenance.log
"@ -ForegroundColor Green

} catch {
    Write-Host "Error al crear la tarea: $_" -ForegroundColor Red
    exit 1
}

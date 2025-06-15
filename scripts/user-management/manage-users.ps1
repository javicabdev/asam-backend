# Script para gestión manual de usuarios
# Ejecuta el programa de gestión de usuarios con el entorno correcto

$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent (Split-Path -Parent $scriptPath)

# Cambiar al directorio del proyecto
Push-Location $projectRoot

try {
    Write-Host "Iniciando gestión de usuarios..." -ForegroundColor Green
    
    # Ejecutar el script de Go
    go run scripts/user-management/manage-users.go
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error al ejecutar el gestor de usuarios" -ForegroundColor Red
    }
} finally {
    Pop-Location
}

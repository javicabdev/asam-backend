# Script para ejecutar tareas de mantenimiento de tokens

param(
    [switch]$Cleanup = $false,
    [switch]$Limit = $false,
    [switch]$All = $false,
    [switch]$DryRun = $false,
    [switch]$Report = $false,
    [int]$TokenLimit = 0,
    [switch]$Help = $false
)

function Show-Help {
    Write-Host @"
Uso: .\maintenance.ps1 [opciones]

Opciones:
  -Cleanup         Limpiar tokens expirados
  -Limit           Aplicar límite de tokens por usuario
  -All             Ejecutar todas las tareas de mantenimiento
  -DryRun          Mostrar qué se haría sin ejecutar cambios
  -Report          Generar reporte de mantenimiento
  -TokenLimit N    Establecer límite personalizado de tokens (por defecto: config)
  -Help            Mostrar esta ayuda

Ejemplos:
  .\maintenance.ps1 -All                    # Ejecutar todas las tareas
  .\maintenance.ps1 -Cleanup -DryRun        # Ver qué tokens se limpiarían
  .\maintenance.ps1 -Limit -TokenLimit 3    # Limitar a 3 tokens por usuario
"@
}

# Mostrar ayuda si se solicita
if ($Help) {
    Show-Help
    exit 0
}

# Construir comando
$cmd = "go run cmd/maintenance/main.go"
$args = @()

if ($Cleanup) {
    $args += "-cleanup-tokens"
}

if ($Limit) {
    $args += "-enforce-token-limit"
}

if ($All) {
    $args += "-all"
}

if ($DryRun) {
    $args += "-dry-run"
}

if ($Report) {
    $args += "-report"
}

if ($TokenLimit -gt 0) {
    $args += "-token-limit=$TokenLimit"
}

# Ejecutar comando
$fullCommand = "$cmd $($args -join ' ')"
Write-Host "Ejecutando: $fullCommand" -ForegroundColor Cyan

# Ejecutar
$process = Start-Process -FilePath "go" -ArgumentList "run", "cmd/maintenance/main.go", $args -NoNewWindow -PassThru -Wait

# Verificar resultado
if ($process.ExitCode -eq 0) {
    Write-Host "`nTareas de mantenimiento completadas exitosamente" -ForegroundColor Green
} else {
    Write-Host "`nError durante las tareas de mantenimiento (código: $($process.ExitCode))" -ForegroundColor Red
    exit $process.ExitCode
}

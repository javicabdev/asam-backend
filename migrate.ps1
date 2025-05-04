# Script PowerShell para ejecutar migraciones
param (
    [Parameter(Position=0)]
    [string]$Environment = "local",
    
    [Parameter(Position=1)]
    [string]$Command = "up",
    
    [Parameter(Position=2, ValueFromRemainingArguments=$true)]
    [string[]]$ExtraArgs
)

# Validar el entorno
if ($Environment -notin @("local", "aiven", "all")) {
    Write-Host "Entorno inválido. Debe ser 'local', 'aiven' o 'all'" -ForegroundColor Red
    exit 1
}

# Mostrar información
Write-Host "Ejecutando migraciones para entorno: $Environment, comando: $Command $ExtraArgs" -ForegroundColor Green

# Construir los argumentos
$argList = @("-env=$Environment", "-cmd=$Command")
if ($ExtraArgs) {
    $argList += $ExtraArgs
}

# Ejecutar el comando de migración a través de Go
Write-Host "Ejecutando via Go: go run cmd/migrate/main.go $($argList -join ' ')" -ForegroundColor Cyan
$process = Start-Process -FilePath "go" -ArgumentList (@("run", "cmd/migrate/main.go") + $argList) -Wait -NoNewWindow -PassThru

# Verificar el resultado
if ($process.ExitCode -eq 0) {
    Write-Host "Migración completada exitosamente" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Error al ejecutar la migración (código: $($process.ExitCode))" -ForegroundColor Red
    exit $process.ExitCode
}

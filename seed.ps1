# Script PowerShell para ejecutar seed
param (
    [Parameter(Position=0)]
    [string]$Environment = "local",
    
    [Parameter(Position=1)]
    [string]$Type = "minimal",
    
    [Parameter()]
    [switch]$Clean,
    
    [Parameter()]
    [int]$Members = 0,
    
    [Parameter()]
    [int]$Families = 0,
    
    [Parameter()]
    [int]$Familiares = 0,
    
    [Parameter()]
    [int]$Payments = 0,
    
    [Parameter()]
    [int]$Cashflows = 0,
    
    [Parameter()]
    [string]$Scenario = "payment_overdue"
)

# Validar el entorno
if ($Environment -notin @("local", "aiven", "all")) {
    Write-Host "Entorno inválido. Debe ser 'local', 'aiven' o 'all'" -ForegroundColor Red
    exit 1
}

# Mostrar información
Write-Host "Ejecutando seed para entorno: $Environment, tipo: $Type" -ForegroundColor Green

# Construir argumentos
$argList = @("-env=$Environment", "-type=$Type")

if ($Clean) {
    $argList += "-clean"
    Write-Host "Modo limpieza activado" -ForegroundColor Yellow
}

if ($Type -eq "scenario") {
    $argList += "-scenario=$Scenario"
    Write-Host "Escenario: $Scenario" -ForegroundColor Yellow
}

if ($Type -eq "custom") {
    if ($Members -gt 0) { $argList += "-members=$Members" }
    if ($Families -gt 0) { $argList += "-families=$Families" }
    if ($Familiares -gt 0) { $argList += "-familiares=$Familiares" }
    if ($Payments -gt 0) { $argList += "-payments=$Payments" }
    if ($Cashflows -gt 0) { $argList += "-cashflows=$Cashflows" }
    
    Write-Host "Configuración personalizada: Members=$Members, Families=$Families, Familiares=$Familiares, Payments=$Payments, Cashflows=$Cashflows" -ForegroundColor Yellow
}

# Ejecutar el comando de seed directamente con Go
Write-Host "Ejecutando via Go: go run cmd/seed/main.go $($argList -join ' ')" -ForegroundColor Cyan
$process = Start-Process -FilePath "go" -ArgumentList (@("run", "cmd/seed/main.go") + $argList) -Wait -NoNewWindow -PassThru

# Verificar el resultado
if ($process.ExitCode -eq 0) {
    Write-Host "Seed completado exitosamente" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Error al ejecutar el seed (código: $($process.ExitCode))" -ForegroundColor Red
    exit $process.ExitCode
}

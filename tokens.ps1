# Token Management Helper Script

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("test", "dashboard", "sessions", "clean", "fix", "rebuild")]
    [string]$Command = "dashboard"
)

Write-Host "🔐 Token Management Helper" -ForegroundColor Cyan

switch ($Command) {
    "test" {
        Write-Host "Ejecutando test de tokens..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\test-refresh-tokens.ps1"
    }
    "dashboard" {
        Write-Host "Mostrando dashboard de tokens..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\token-dashboard.ps1"
    }
    "sessions" {
        Write-Host "Mostrando sesiones activas..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\show-sessions.ps1"
    }
    "clean" {
        Write-Host "Limpiando todas las sesiones..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\clean-sessions.ps1"
    }
    "fix" {
        Write-Host "Arreglando tokens con fechas incorrectas..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\fix-refresh-tokens.ps1"
    }
    "rebuild" {
        Write-Host "Reconstruyendo y testeando..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\tokens\quick-rebuild.ps1"
    }
}

Write-Host "`n✅ Comando completado!" -ForegroundColor Green

if ($Command -eq "dashboard") {
    Write-Host "`n💡 Otros comandos disponibles:" -ForegroundColor Gray
    Write-Host "  .\tokens.ps1 test      - Ejecutar test de login" -ForegroundColor DarkGray
    Write-Host "  .\tokens.ps1 sessions  - Ver sesiones activas" -ForegroundColor DarkGray
    Write-Host "  .\tokens.ps1 clean     - Limpiar todas las sesiones" -ForegroundColor DarkGray
    Write-Host "  .\tokens.ps1 fix       - Arreglar fechas incorrectas" -ForegroundColor DarkGray
    Write-Host "  .\tokens.ps1 rebuild   - Reconstruir API y testear" -ForegroundColor DarkGray
}

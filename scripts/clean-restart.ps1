# Script de inicio rápido para problema de migraciones
# Limpia todo y reinicia desde cero

Write-Host @"
╔═══════════════════════════════════════╗
║   ASAM Backend - Reinicio Completo    ║
╚═══════════════════════════════════════╝
"@ -ForegroundColor Cyan

Write-Host "`n🧹 Limpiando todo para un inicio limpio..." -ForegroundColor Yellow

# Detener todos los contenedores
Write-Host "`n🛑 Deteniendo contenedores..." -ForegroundColor Yellow
docker-compose down -v

# Eliminar el archivo .env para empezar limpio
if (Test-Path ".env") {
    Write-Host "🗑️  Eliminando archivo .env existente..." -ForegroundColor Yellow
    Remove-Item ".env" -Force
}

# Esperar un momento
Start-Sleep -Seconds 2

Write-Host "`n🚀 Iniciando con configuración limpia..." -ForegroundColor Green

# Ejecutar el script principal con flag --clean
& "$PSScriptRoot\..\start-docker.ps1" --clean

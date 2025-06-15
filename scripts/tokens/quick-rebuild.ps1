# Quick rebuild and test

Write-Host "🔄 Rebuild rápido..." -ForegroundColor Cyan

# Solo detener el contenedor de la API
Write-Host "`n🛑 Deteniendo API..." -ForegroundColor Yellow
docker-compose stop api

# Rebuild solo la API
Write-Host "`n🔨 Reconstruyendo API..." -ForegroundColor Yellow
docker-compose build api

# Reiniciar la API
Write-Host "`n🚀 Reiniciando API..." -ForegroundColor Yellow
docker-compose up -d api

# Esperar un poco
Write-Host "`n⏳ Esperando a que la API esté lista..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Arreglar tokens existentes
Write-Host "`n🔧 Arreglando tokens existentes..." -ForegroundColor Yellow
& "$PSScriptRoot\fix-refresh-tokens.ps1"

# Ejecutar test
Write-Host "`n🧪 Ejecutando test..." -ForegroundColor Yellow
& "$PSScriptRoot\test-refresh-tokens.ps1"

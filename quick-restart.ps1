# Reinicio rápido del API con nueva configuración
Write-Host "🔄 Aplicando nueva configuración y reiniciando API..." -ForegroundColor Green
Write-Host ""

Write-Host "🛑 Deteniendo API..." -ForegroundColor Yellow
docker-compose stop api

Write-Host ""
Write-Host "🚀 Iniciando API con nueva configuración..." -ForegroundColor Yellow
docker-compose up -d api

Write-Host ""
Write-Host "⏳ Esperando a que el API inicie..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host ""
Write-Host "📋 Últimos logs del API:" -ForegroundColor Cyan
docker-compose logs --tail=50 api

Write-Host ""
Write-Host "🔍 Para seguir viendo logs en tiempo real:" -ForegroundColor Yellow
Write-Host "   docker-compose logs -f api" -ForegroundColor White

# Script para verificar las variables de entorno en el contenedor
Write-Host "🔍 Verificando variables de entorno en el contenedor del API..." -ForegroundColor Green
Write-Host ""

Write-Host "Variables JWT:" -ForegroundColor Yellow
docker-compose exec api sh -c 'env | grep JWT | sort'

Write-Host ""
Write-Host "Variables de Base de Datos:" -ForegroundColor Yellow
docker-compose exec api sh -c 'env | grep DB_ | sort'

Write-Host ""
Write-Host "Todas las variables:" -ForegroundColor Yellow
docker-compose exec api sh -c 'env | sort'

Write-Host ""
Write-Host "📋 Si faltan variables JWT, verifica que estén en el archivo .env" -ForegroundColor Cyan

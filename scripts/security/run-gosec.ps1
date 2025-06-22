# Script para ejecutar gosec localmente con la misma configuración que en CI/CD
# Esto ayuda a verificar que los problemas de seguridad están resueltos antes de hacer push

Write-Host "🔍 Ejecutando análisis de seguridad con gosec..." -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan

# Verificar si gosec está instalado
$gosecPath = Get-Command gosec -ErrorAction SilentlyContinue
if (-not $gosecPath) {
    Write-Host "❌ gosec no está instalado." -ForegroundColor Red
    Write-Host "Instalando gosec..." -ForegroundColor Yellow
    go install github.com/securego/gosec/v2/cmd/gosec@latest
}

# Generar código GraphQL si es necesario
if (-not (Test-Path "internal/adapters/gql/generated")) {
    Write-Host "📝 Generando código GraphQL..." -ForegroundColor Yellow
    go run ./cmd/generate/main.go
}

# Ejecutar gosec con la configuración del proyecto
Write-Host "🔒 Analizando código con gosec..." -ForegroundColor Yellow
$result = gosec -conf .gosec.json ./... 2>&1

# Mostrar resultado
Write-Host $result

# Verificar si hubo errores
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Análisis de seguridad completado sin problemas" -ForegroundColor Green
} else {
    Write-Host "❌ Se encontraron problemas de seguridad" -ForegroundColor Red
    exit 1
}

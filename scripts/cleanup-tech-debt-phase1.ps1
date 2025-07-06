# Script para limpiar deuda técnica - Fase 1

Write-Host "=== Limpieza de Deuda Técnica - Fase 1 ===" -ForegroundColor Cyan
Write-Host ""

# 1. Eliminar la carpeta infrastructure
Write-Host "1. Eliminando carpeta internal/infrastructure..." -ForegroundColor Yellow
if (Test-Path "internal\infrastructure") {
    Remove-Item -Path "internal\infrastructure" -Recurse -Force
    Write-Host "   ✓ Carpeta internal/infrastructure eliminada" -ForegroundColor Green
} else {
    Write-Host "   ⚠ La carpeta internal/infrastructure no existe" -ForegroundColor DarkYellow
}

# 2. Ejecutar go mod tidy
Write-Host ""
Write-Host "2. Limpiando dependencias no utilizadas..." -ForegroundColor Yellow
& go mod tidy
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ go mod tidy ejecutado" -ForegroundColor Green
} else {
    Write-Host "   ✗ Error al ejecutar go mod tidy" -ForegroundColor Red
    exit 1
}

# 3. Verificar que el proyecto compila
Write-Host ""
Write-Host "3. Verificando que el proyecto compila..." -ForegroundColor Yellow
& go build ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ El proyecto compila correctamente" -ForegroundColor Green
} else {
    Write-Host "   ✗ Error al compilar el proyecto" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Limpieza completada exitosamente ===" -ForegroundColor Cyan

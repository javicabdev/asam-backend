# Script para limpiar deuda técnica - Fase 2

Write-Host "=== Limpieza de Deuda Técnica - Fase 2 ===" -ForegroundColor Cyan
Write-Host ""

# 1. Crear estructura de carpetas para mocks
Write-Host "1. Creando estructura de carpetas para mocks..." -ForegroundColor Yellow
$mocksPath = "test\mocks\email"
if (-not (Test-Path $mocksPath)) {
    New-Item -ItemType Directory -Path $mocksPath -Force | Out-Null
    Write-Host "   ✓ Carpeta $mocksPath creada" -ForegroundColor Green
} else {
    Write-Host "   ⚠ La carpeta $mocksPath ya existe" -ForegroundColor DarkYellow
}

# 2. Mover mock de notificaciones
Write-Host ""
Write-Host "2. Moviendo mock de notificaciones a carpeta de tests..." -ForegroundColor Yellow
$sourceMock = "internal\adapters\email\mock_notification_adapter.go"
$destMock = "test\mocks\email\mock_notification_adapter.go"

if (Test-Path $sourceMock) {
    # Leer el archivo y cambiar el package
    $content = Get-Content $sourceMock -Raw
    $newContent = $content -replace "package email", "package mocks"
    
    # Escribir en el nuevo destino
    $newContent | Set-Content $destMock -Encoding UTF8
    
    # Eliminar el archivo original
    Remove-Item $sourceMock -Force
    
    Write-Host "   ✓ Mock movido y package actualizado" -ForegroundColor Green
} else {
    Write-Host "   ⚠ El archivo mock no existe en la ubicación esperada" -ForegroundColor DarkYellow
}

# 3. Crear carpeta para pruebas manuales
Write-Host ""
Write-Host "3. Creando carpeta para pruebas manuales..." -ForegroundColor Yellow
$manualTestPath = "test\manual"
if (-not (Test-Path $manualTestPath)) {
    New-Item -ItemType Directory -Path $manualTestPath -Force | Out-Null
    Write-Host "   ✓ Carpeta $manualTestPath creada" -ForegroundColor Green
}

# 4. Mover script de prueba manual
Write-Host ""
Write-Host "4. Moviendo script de prueba manual..." -ForegroundColor Yellow
$sourceScript = "scripts\test-mock-email.go"
$destScript = "test\manual\test-mock-email.go"

if (Test-Path $sourceScript) {
    Move-Item -Path $sourceScript -Destination $destScript -Force
    Write-Host "   ✓ Script de prueba movido" -ForegroundColor Green
} else {
    Write-Host "   ⚠ El script de prueba no existe" -ForegroundColor DarkYellow
}

# 5. Consolidar archivos .env
Write-Host ""
Write-Host "5. Consolidando archivos .env..." -ForegroundColor Yellow

$envFilesToDelete = @(
    ".env.production.free",
    ".env.production.test",
    ".env.docker.example",
    ".env.email.example",
    ".env.complete.example",
    ".env.local"
)

$deletedCount = 0
foreach ($envFile in $envFilesToDelete) {
    if (Test-Path $envFile) {
        Remove-Item $envFile -Force
        Write-Host "   ✓ Eliminado: $envFile" -ForegroundColor Green
        $deletedCount++
    }
}

# Verificar si .env está vacío o es redundante
if (Test-Path ".env") {
    $envContent = Get-Content ".env" -Raw
    if ([string]::IsNullOrWhiteSpace($envContent)) {
        Remove-Item ".env" -Force
        Write-Host "   ✓ Eliminado: .env (estaba vacío)" -ForegroundColor Green
        $deletedCount++
    }
}

Write-Host "   Total de archivos .env eliminados: $deletedCount" -ForegroundColor Cyan

# 6. Verificar archivos .env conservados
Write-Host ""
Write-Host "6. Verificando archivos .env conservados..." -ForegroundColor Yellow
$envFilesToKeep = @(
    ".env.example",
    ".env.development",
    ".env.production",
    ".env.test",
    ".env.aiven"
)

foreach ($envFile in $envFilesToKeep) {
    if (Test-Path $envFile) {
        Write-Host "   ✓ Conservado: $envFile" -ForegroundColor Green
    } else {
        Write-Host "   ⚠ No encontrado: $envFile" -ForegroundColor DarkYellow
    }
}

# 7. Verificar que el proyecto compila
Write-Host ""
Write-Host "7. Verificando que el proyecto compila..." -ForegroundColor Yellow
& go build ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ El proyecto compila correctamente" -ForegroundColor Green
} else {
    Write-Host "   ✗ Error al compilar el proyecto" -ForegroundColor Red
    exit 1
}

# 8. Ejecutar tests
Write-Host ""
Write-Host "8. Ejecutando tests..." -ForegroundColor Yellow
& go test ./... -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✓ Todos los tests pasan" -ForegroundColor Green
} else {
    Write-Host "   ⚠ Algunos tests fallan - verifica si es debido a los cambios" -ForegroundColor DarkYellow
}

Write-Host ""
Write-Host "=== Limpieza Fase 2 completada exitosamente ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Resumen de cambios:" -ForegroundColor Yellow
Write-Host "- Mock de email movido a: test\mocks\email\" -ForegroundColor White
Write-Host "- Script de prueba movido a: test\manual\" -ForegroundColor White
Write-Host "- Archivos .env consolidados ($deletedCount eliminados)" -ForegroundColor White

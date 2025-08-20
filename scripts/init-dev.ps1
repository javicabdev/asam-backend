# scripts/init-dev.ps1

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
Write-Host "Inicializando entorno de desarrollo local..." -ForegroundColor Green

# 1. Crear .env.development si no existe
$envFile = "$projectRoot\.env.development"
if (-not (Test-Path $envFile)) {
    Write-Host "Creando archivo .env.development..." -ForegroundColor Yellow
    Copy-Item "$projectRoot\.env.development.example" $envFile
    Write-Host "⚠️  Por favor, revisa y configura las variables en .env.development" -ForegroundColor Yellow
}

# 2. Descargar dependencias de Go
Write-Host "Descargando dependencias de Go..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error descargando dependencias de Go" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Dependencias descargadas" -ForegroundColor Green

# 3. Verificar Docker
Write-Host "Verificando Docker..." -ForegroundColor Cyan
docker --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker no encontrado. Por favor, instálalo antes de continuar." -ForegroundColor Red
    Write-Host "   Descarga Docker Desktop desde: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ Docker verificado" -ForegroundColor Green

# 4. Verificar que air está instalado
Write-Host "Verificando air..." -ForegroundColor Cyan
$airCommand = Get-Command air -ErrorAction SilentlyContinue
if (-not $airCommand) {
    Write-Host "❌ Air no encontrado. Instalando..." -ForegroundColor Yellow
    go install github.com/air-verse/air@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error instalando air" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Air instalado" -ForegroundColor Green
} else {
    Write-Host "✅ Air verificado" -ForegroundColor Green
}

Write-Host "`nEntorno de desarrollo inicializado! 🚀" -ForegroundColor Green
Write-Host "`nPasos siguientes:"
Write-Host "1. Revisa y configura las variables en .env.development" -ForegroundColor Yellow
Write-Host "2. Ejecuta: .\scripts\run-all.ps1 docker-up" -ForegroundColor Yellow
Write-Host "3. Ejecuta: .\scripts\run-all.ps1 air" -ForegroundColor Yellow

Write-Host "`nPara más información, consulta el README.md del proyecto" -ForegroundColor Cyan
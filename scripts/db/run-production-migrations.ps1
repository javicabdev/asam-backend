# Script para ejecutar migraciones en producción
# Requiere gcloud CLI configurado y acceso a los secretos

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("up", "down", "version")]
    [string]$Command = "up"
)

Write-Host "=== Ejecutando migraciones en producción ===" -ForegroundColor Cyan
Write-Host ""

# Verificar que gcloud está instalado
try {
    $null = gcloud --version
} catch {
    Write-Host "Error: gcloud CLI no está instalado o no está en el PATH" -ForegroundColor Red
    Write-Host "Instálalo desde: https://cloud.google.com/sdk/docs/install" -ForegroundColor Yellow
    exit 1
}

# Obtener secretos de Google Secret Manager
Write-Host "Obteniendo credenciales de base de datos..." -ForegroundColor Yellow

try {
    $env:DB_HOST = gcloud secrets versions access latest --secret=db-host 2>$null
    $env:DB_PORT = gcloud secrets versions access latest --secret=db-port 2>$null
    $env:DB_USER = gcloud secrets versions access latest --secret=db-user 2>$null
    $env:DB_PASSWORD = gcloud secrets versions access latest --secret=db-password 2>$null
    $env:DB_NAME = gcloud secrets versions access latest --secret=db-name 2>$null
    $env:DB_SSL_MODE = "require"
} catch {
    Write-Host "Error al obtener secretos. Asegúrate de:" -ForegroundColor Red
    Write-Host "1. Estar autenticado con: gcloud auth login" -ForegroundColor Yellow
    Write-Host "2. Tener el proyecto correcto: gcloud config set project YOUR_PROJECT_ID" -ForegroundColor Yellow
    Write-Host "3. Tener permisos para acceder a Secret Manager" -ForegroundColor Yellow
    exit 1
}

# Verificar que tenemos todas las variables
if (-not $env:DB_HOST -or -not $env:DB_PORT -or -not $env:DB_USER -or -not $env:DB_PASSWORD -or -not $env:DB_NAME) {
    Write-Host "Error: No se pudieron obtener todas las credenciales de la base de datos" -ForegroundColor Red
    exit 1
}

Write-Host "Credenciales obtenidas correctamente" -ForegroundColor Green
Write-Host ""

# Mostrar información de conexión (sin password)
Write-Host "Conectando a:" -ForegroundColor Cyan
Write-Host "  Host: $env:DB_HOST" -ForegroundColor Gray
Write-Host "  Port: $env:DB_PORT" -ForegroundColor Gray
Write-Host "  Database: $env:DB_NAME" -ForegroundColor Gray
Write-Host "  User: $env:DB_USER" -ForegroundColor Gray
Write-Host "  SSL: $env:DB_SSL_MODE" -ForegroundColor Gray
Write-Host ""

# Ejecutar migraciones
Write-Host "Ejecutando comando: $Command" -ForegroundColor Yellow
Write-Host ""

try {
    go run cmd/migrate/main.go -cmd $Command
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✓ Migraciones ejecutadas exitosamente" -ForegroundColor Green
    } else {
        Write-Host ""
        Write-Host "✗ Error al ejecutar migraciones" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error al ejecutar migraciones: $_" -ForegroundColor Red
    exit 1
}

# Limpiar variables de entorno
Remove-Item Env:DB_HOST -ErrorAction SilentlyContinue
Remove-Item Env:DB_PORT -ErrorAction SilentlyContinue
Remove-Item Env:DB_USER -ErrorAction SilentlyContinue
Remove-Item Env:DB_PASSWORD -ErrorAction SilentlyContinue
Remove-Item Env:DB_NAME -ErrorAction SilentlyContinue
Remove-Item Env:DB_SSL_MODE -ErrorAction SilentlyContinue

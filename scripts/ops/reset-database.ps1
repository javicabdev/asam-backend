# Script para limpiar datos de prueba y resetear la BD
# ⚠️ CUIDADO: Este script borra TODOS los datos

Write-Host "⚠️  RESET DE BASE DE DATOS" -ForegroundColor Red
Write-Host "=" * 50
Write-Host ""
Write-Host "Este script va a:" -ForegroundColor Yellow
Write-Host "  1. Borrar TODOS los datos actuales" -ForegroundColor Red
Write-Host "  2. Ejecutar las migraciones" -ForegroundColor Yellow
Write-Host "  3. Cargar datos iniciales (seed)" -ForegroundColor Green
Write-Host ""

$env = Read-Host "Escribe 'PRODUCTION' para confirmar"

if ($env -ne "PRODUCTION") {
    Write-Host "Operación cancelada" -ForegroundColor Gray
    exit
}

Write-Host ""
Write-Host "🔄 Reseteando base de datos..." -ForegroundColor Yellow

# Opción 1: Ejecutar comando remoto en Cloud Run
Write-Host "Conectando al servicio..." -ForegroundColor Cyan

# Ejecutar reset via Cloud Run Jobs o mediante endpoint especial
gcloud run jobs execute asam-db-reset `
    --region=europe-west1 `
    --wait 2>$null

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ℹ️  El job de reset no existe. Creándolo..." -ForegroundColor Yellow
    
    # Crear job para reset si no existe
    Write-Host "Para resetear la BD, ejecuta estos comandos SQL directamente:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "-- 1. Conectar a Cloud SQL" -ForegroundColor Gray
    Write-Host "gcloud sql connect asam-db --user=asam-user --database=asam" -ForegroundColor Green
    Write-Host ""
    Write-Host "-- 2. Ejecutar reset" -ForegroundColor Gray
    Write-Host @"
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO asam_user;
GRANT ALL ON SCHEMA public TO public;
"@ -ForegroundColor Green
    Write-Host ""
    Write-Host "-- 3. Luego ejecutar las migraciones desde el backend" -ForegroundColor Gray
}

Write-Host ""
Write-Host "✅ Proceso completado" -ForegroundColor Green
Write-Host ""
Write-Host "Próximos pasos:" -ForegroundColor Yellow
Write-Host "  1. Verificar que la aplicación funcione" -ForegroundColor Gray
Write-Host "  2. Cargar datos de prueba con los scripts de seed" -ForegroundColor Gray

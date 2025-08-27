# Script para hacer backup antes de cambios importantes
Write-Host "💾 Backup de Base de Datos ASAM" -ForegroundColor Cyan
Write-Host "=" * 50

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$backupName = "asam-backup-$timestamp"

Write-Host ""
Write-Host "Creando backup: $backupName" -ForegroundColor Yellow

# Exportar usando Cloud SQL
gcloud sql export sql asam-db `
    gs://asam-backups/$backupName.sql `
    --database=asam

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Backup creado exitosamente" -ForegroundColor Green
    Write-Host "📍 Ubicación: gs://asam-backups/$backupName.sql" -ForegroundColor Gray
    
    # Guardar referencia local
    $backupInfo = @{
        Timestamp = $timestamp
        Name = $backupName
        Date = Get-Date
    }
    
    $backupInfo | ConvertTo-Json | Out-File -FilePath ".\backups\last-backup.json"
    
    Write-Host ""
    Write-Host "Para restaurar este backup:" -ForegroundColor Yellow
    Write-Host "  .\scripts\ops\restore-database.ps1 -BackupName $backupName" -ForegroundColor Cyan
} else {
    Write-Host "❌ Error al crear backup" -ForegroundColor Red
    Write-Host ""
    Write-Host "Posibles soluciones:" -ForegroundColor Yellow
    Write-Host "  1. Verificar que el bucket gs://asam-backups existe" -ForegroundColor Gray
    Write-Host "  2. Crear el bucket: gsutil mb gs://asam-backups" -ForegroundColor Gray
    Write-Host "  3. Verificar permisos de Cloud SQL" -ForegroundColor Gray
}

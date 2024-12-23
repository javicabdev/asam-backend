# scripts/backup-db.ps1
param (
    [string]$BackupPath = ".\backups"
)

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$backupFile = Join-Path $BackupPath "asam_backup_${timestamp}.sql"

# Crear directorio si no existe
if (-not (Test-Path $BackupPath)) {
    New-Item -ItemType Directory -Path $BackupPath
}

# Cargar variables de entorno de producción
Get-Content "$projectRoot\.env.production" | ForEach-Object {
    $line = $_.Trim()
    if ($line -and !$line.StartsWith("#")) {
        $key, $value = $line.Split('=', 2)
        Set-Item -Path "env:$key" -Value $value
    }
}

# Backup usando pg_dump
$env:PGPASSWORD = $env:DB_PASSWORD
Write-Host "Iniciando backup de la base de datos..."
pg_dump -h $env:DB_HOST -U $env:DB_USER -d $env:DB_NAME -F c -f $backupFile

if ($LASTEXITCODE -eq 0) {
    Write-Host "Backup creado exitosamente: $backupFile" -ForegroundColor Green
} else {
    Write-Host "Error al crear el backup" -ForegroundColor Red
}

# Limpiar backups antiguos (mantener últimos 7 días)
Get-ChildItem -Path $BackupPath -Filter "asam_backup_*.sql" |
        Where-Object { $_.CreationTime -lt (Get-Date).AddDays(-7) } |
        ForEach-Object {
            Remove-Item $_.FullName
            Write-Host "Eliminado backup antiguo: $($_.Name)" -ForegroundColor Yellow
        }
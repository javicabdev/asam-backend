# scripts/restore-db.ps1
param (
    [string]$BackupFile,         # Ruta completa al archivo de backup
    [string]$BackupPath = ".\backups",  # Ruta al directorio de backups (opcional)
    [string]$Env = "production"  # Entorno (por defecto "production")
)

if (-not $BackupFile) {
    Write-Host "Error: Debes proporcionar la ruta al archivo de backup usando el parámetro -BackupFile." -ForegroundColor Red
    exit 1
}

function Get-ProjectRoot {
    return Split-Path -Parent $PSScriptRoot
}

$projectRoot = Get-ProjectRoot
Write-Host "restore-db.ps1 - Project Root: '$projectRoot'" -ForegroundColor Cyan  # Depuración

# Establecer la variable de entorno APP_ENV
$env:APP_ENV = $Env
Write-Host "restore-db.ps1 - Set APP_ENV to '$Env'" -ForegroundColor Cyan  # Depuración

# Cargar variables de entorno desde el archivo .env correspondiente
Write-Host "Cargando configuración desde .env.$Env..." -ForegroundColor Green
$envFile = "$projectRoot\.env.$Env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -and !$line.StartsWith("#")) {
            $key, $value = $line.Split('=', 2)
            Set-Item -Path "env:$key" -Value $value
            Write-Host "restore-db.ps1 - Set variable '$($key)' to '$($value)'" -ForegroundColor Cyan  # Depuración
        }
    }
} else {
    Write-Host "Advertencia: El archivo $envFile no existe." -ForegroundColor Yellow
}

# Parámetros de restauración
$dbHost = $env:DB_HOST
$dbPort = $env:DB_PORT
$dbUser = $env:DB_USER
$dbPassword = $env:DB_PASSWORD
$dbName = $env:DB_NAME

# Validar que el archivo de backup existe
if (!(Test-Path -Path $BackupFile)) {
    Write-Host "Error: El archivo de backup '$BackupFile' no existe." -ForegroundColor Red
    exit 1
}

# Exportar la variable PGPASSWORD para autenticación no interactiva
$env:PGPASSWORD = $dbPassword

Write-Host "restore-db.ps1 - Iniciando restauración de la base de datos '$dbName' desde '$BackupFile'..." -ForegroundColor Green

# Determinar el formato del backup basado en la extensión del archivo
$backupExtension = [System.IO.Path]::GetExtension($BackupFile).ToLower()

# Ruta a pg_restore.exe (ajusta según tu instalación)
$pgRestorePath = "C:\Program Files\PostgreSQL\13\bin\pg_restore.exe"  # Actualiza según tu versión y ruta

# Ruta a psql.exe (si usas psql para restaurar)
$psqlPath = "C:\Program Files\PostgreSQL\13\bin\psql.exe"  # Actualiza según tu versión y ruta

if ($backupExtension -eq ".sql") {
    # Usar psql para restaurar un backup en formato SQL
    Write-Host "restore-db.ps1 - Restaurando usando psql (Formato SQL)..." -ForegroundColor Green
    $psqlArgs = @(
        "--host=$dbHost",
        "--port=$dbPort",
        "--username=$dbUser",
        "--dbname=$dbName",
        "--file=`"$BackupFile`"",
        "--no-password"
    )

    try {
        & $psqlPath @psqlArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Host "restore-db.ps1 - Restauración completada exitosamente." -ForegroundColor Green
        } else {
            Write-Host "restore-db.ps1 - Error al restaurar la base de datos." -ForegroundColor Red
        }
    } catch {
        Write-Host "restore-db.ps1 - Error al ejecutar psql: $_" -ForegroundColor Red
    }
} elseif ($backupExtension -eq ".dump" -or $backupExtension -eq ".backup" -or $backupExtension -eq ".tar") {
    # Usar pg_restore para restaurar un backup en formato custom
    Write-Host "restore-db.ps1 - Restaurando usando pg_restore (Formato Custom)..." -ForegroundColor Green
    $pgRestoreArgs = @(
        "--host=$dbHost",
        "--port=$dbPort",
        "--username=$dbUser",
        "--dbname=$dbName",
        "--verbose",
        "--clean",    # Elimina objetos antes de recrearlos
        "--no-password",
        "`"$BackupFile`""
    )

    try {
        & $pgRestorePath @pgRestoreArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Host "restore-db.ps1 - Restauración completada exitosamente." -ForegroundColor Green
        } else {
            Write-Host "restore-db.ps1 - Error al restaurar la base de datos." -ForegroundColor Red
        }
    } catch {
        Write-Host "restore-db.ps1 - Error al ejecutar pg_restore: $_" -ForegroundColor Red
    }
} else {
    Write-Host "restore-db.ps1 - Formato de archivo de backup no soportado: '$backupExtension'" -ForegroundColor Red
}

# Limpieza de variables de entorno
Remove-Item Env:\APP_ENV -ErrorAction SilentlyContinue
Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
Write-Host "restore-db.ps1 - Limpieza de variables de entorno completada." -ForegroundColor Cyan  # Depuración

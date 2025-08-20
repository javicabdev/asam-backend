# Script para limpiar deuda técnica - Fase 3

Write-Host "=== Limpieza de Deuda Técnica - Fase 3 ===" -ForegroundColor Cyan
Write-Host "    Reorganización de scripts" -ForegroundColor Cyan
Write-Host ""

# Función para mover archivo preservando su estado
function Move-ScriptFile {
    param(
        [string]$Source,
        [string]$Destination
    )
    
    if (Test-Path $Source) {
        # Crear directorio destino si no existe
        $destDir = Split-Path -Parent $Destination
        if (-not (Test-Path $destDir)) {
            New-Item -ItemType Directory -Path $destDir -Force | Out-Null
        }
        
        Move-Item -Path $Source -Destination $Destination -Force
        Write-Host "   ✓ Movido: $Source → $Destination" -ForegroundColor Green
        return $true
    } else {
        Write-Host "   ⚠ No encontrado: $Source" -ForegroundColor DarkYellow
        return $false
    }
}

$movedCount = 0
$baseDir = "scripts"

# 1. Crear estructura de carpetas
Write-Host "1. Creando estructura de carpetas..." -ForegroundColor Yellow
$folders = @("db", "dev", "docker", "ops", "ops/gcp", "verification")
foreach ($folder in $folders) {
    $path = Join-Path $baseDir $folder
    if (-not (Test-Path $path)) {
        New-Item -ItemType Directory -Path $path -Force | Out-Null
        Write-Host "   ✓ Carpeta creada: $path" -ForegroundColor Green
    }
}

# 2. Mover scripts de base de datos
Write-Host ""
Write-Host "2. Moviendo scripts de base de datos..." -ForegroundColor Yellow
$dbScripts = @(
    @{src="migrate.ps1"; dst="db/migrate.ps1"},
    @{src="backup-db.ps1"; dst="db/backup-db.ps1"},
    @{src="restore-db.ps1"; dst="db/restore-db.ps1"},
    @{src="apply-email-required-changes.ps1"; dst="db/apply-email-required-changes.ps1"},
    @{src="rollback-email-required-changes.ps1"; dst="db/rollback-email-required-changes.ps1"},
    @{src="run-email-required-migration.ps1"; dst="db/run-email-required-migration.ps1"},
    @{src="run-production-migrations.ps1"; dst="db/run-production-migrations.ps1"},
    @{src="run-production-migrations.sh"; dst="db/run-production-migrations.sh"}
)

foreach ($script in $dbScripts) {
    if (Move-ScriptFile -Source (Join-Path $baseDir $script.src) -Destination (Join-Path $baseDir $script.dst)) {
        $movedCount++
    }
}

# 3. Mover scripts de desarrollo
Write-Host ""
Write-Host "3. Moviendo scripts de desarrollo..." -ForegroundColor Yellow
$devScripts = @(
    @{src="cleanup-tech-debt-phase1.ps1"; dst="dev/cleanup-tech-debt-phase1.ps1"},
    @{src="cleanup-tech-debt-phase1.sh"; dst="dev/cleanup-tech-debt-phase1.sh"},
    @{src="cleanup-tech-debt-phase2.ps1"; dst="dev/cleanup-tech-debt-phase2.ps1"},
    @{src="cleanup-tech-debt-phase2.sh"; dst="dev/cleanup-tech-debt-phase2.sh"},
    @{src="install-hooks.bat"; dst="dev/install-hooks.bat"},
    @{src="install-hooks.sh"; dst="dev/install-hooks.sh"},
    @{src="regenerate-graphql.ps1"; dst="dev/regenerate-graphql.ps1"},
    @{src="regenerate-graphql.sh"; dst="dev/regenerate-graphql.sh"},
    @{src="format.bat"; dst="dev/format.bat"},
    @{src="lint.bat"; dst="dev/lint.bat"},
    @{src="lint.sh"; dst="dev/lint.sh"},
    @{src="test.ps1"; dst="dev/test.ps1"},
    @{src="pre-commit"; dst="dev/pre-commit"},
    @{src="check-config.go"; dst="dev/check-config.go"}
)

foreach ($script in $devScripts) {
    if (Move-ScriptFile -Source (Join-Path $baseDir $script.src) -Destination (Join-Path $baseDir $script.dst)) {
        $movedCount++
    }
}

# 4. Mover scripts de Docker
Write-Host ""
Write-Host "4. Moviendo scripts de Docker..." -ForegroundColor Yellow
$dockerScripts = @(
    @{src="docker-down.ps1"; dst="docker/docker-down.ps1"},
    @{src="docker-restart.ps1"; dst="docker/docker-restart.ps1"},
    @{src="docker-up.ps1"; dst="docker/docker-up.ps1"},
    @{src="logs.ps1"; dst="docker/logs.ps1"}
)

foreach ($script in $dockerScripts) {
    if (Move-ScriptFile -Source (Join-Path $baseDir $script.src) -Destination (Join-Path $baseDir $script.dst)) {
        $movedCount++
    }
}

# 5. Mover scripts operacionales
Write-Host ""
Write-Host "5. Moviendo scripts operacionales..." -ForegroundColor Yellow
$opsScripts = @(
    @{src="build.ps1"; dst="ops/build.ps1"},
    @{src="clean.ps1"; dst="ops/clean.ps1"},
    @{src="run.ps1"; dst="ops/run.ps1"}
)

foreach ($script in $opsScripts) {
    if (Move-ScriptFile -Source (Join-Path $baseDir $script.src) -Destination (Join-Path $baseDir $script.dst)) {
        $movedCount++
    }
}

# 6. Mover scripts de GCP
Write-Host ""
Write-Host "6. Moviendo scripts de GCP..." -ForegroundColor Yellow
$gcpDir = Join-Path $baseDir "gcp"
if (Test-Path $gcpDir) {
    Get-ChildItem -Path $gcpDir -File | ForEach-Object {
        $dest = Join-Path $baseDir "ops/gcp" $_.Name
        Move-Item -Path $_.FullName -Destination $dest -Force
        Write-Host "   ✓ Movido: gcp/$($_.Name) → ops/gcp/$($_.Name)" -ForegroundColor Green
        $movedCount++
    }
    # Eliminar carpeta gcp vacía
    Remove-Item -Path $gcpDir -Force -ErrorAction SilentlyContinue
}

# 7. Mover scripts de verificación
Write-Host ""
Write-Host "7. Moviendo scripts de verificación de email..." -ForegroundColor Yellow
$verificationScripts = @(
    @{src="check-email-verification-status.ps1"; dst="verification/check-email-verification-status.ps1"},
    @{src="check-token-info.ps1"; dst="verification/check-token-info.ps1"},
    @{src="cleanup-verification-tokens.ps1"; dst="verification/cleanup-verification-tokens.ps1"},
    @{src="manually-verify-user.ps1"; dst="verification/manually-verify-user.ps1"},
    @{src="verify-email-manual.sql"; dst="verification/verify-email-manual.sql"}
)

foreach ($script in $verificationScripts) {
    if (Move-ScriptFile -Source (Join-Path $baseDir $script.src) -Destination (Join-Path $baseDir $script.dst)) {
        $movedCount++
    }
}

# 8. Mover scripts de test a test/manual
Write-Host ""
Write-Host "8. Moviendo scripts de test de email..." -ForegroundColor Yellow
$testScripts = @("test-email.go", "test-sendgrid-simple.go", "verify-sendgrid.go", "verify-sendgrid.sh")
$testManualDir = "test/manual"

foreach ($script in $testScripts) {
    $source = Join-Path $baseDir $script
    $dest = Join-Path $testManualDir $script
    if (Move-ScriptFile -Source $source -Destination $dest) {
        $movedCount++
    }
}

# 9. Eliminar archivos innecesarios
Write-Host ""
Write-Host "9. Eliminando archivos innecesarios..." -ForegroundColor Yellow
$toDelete = @("go.mod", "go.sum")
$deletedCount = 0

foreach ($file in $toDelete) {
    $path = Join-Path $baseDir $file
    if (Test-Path $path) {
        Remove-Item -Path $path -Force
        Write-Host "   ✓ Eliminado: $file" -ForegroundColor Green
        $deletedCount++
    }
}

# 10. Mover este script a su ubicación final
Write-Host ""
Write-Host "10. Moviendo este script a su ubicación final..." -ForegroundColor Yellow
$thisScript = "cleanup-tech-debt-phase3.ps1"
$source = Join-Path $baseDir $thisScript
$dest = Join-Path $baseDir "dev" $thisScript
if (-not (Test-Path $dest)) {
    Copy-Item -Path $MyInvocation.MyCommand.Path -Destination $dest -Force
    Write-Host "   ✓ Script copiado a: $dest" -ForegroundColor Green
}

# 11. Crear README.md para scripts
Write-Host ""
Write-Host "11. Creando documentación de scripts..." -ForegroundColor Yellow
$readmePath = Join-Path $baseDir "README.md"
$readmeContent = @'
# Scripts Directory

This directory contains all utility scripts for the ASAM Backend project, organized by functionality.

## Directory Structure

### 📁 db/ - Database Scripts
Scripts for database operations including migrations, backups, and maintenance.

- **migrate.ps1** - Run database migrations locally
- **backup-db.ps1** - Create database backups
- **restore-db.ps1** - Restore database from backup
- **apply-email-required-changes.ps1** - Apply email requirement migration
- **rollback-email-required-changes.ps1** - Rollback email requirement changes
- **run-email-required-migration.ps1** - Run specific email migration
- **run-production-migrations.ps1/.sh** - Run migrations in production

### 📁 dev/ - Development Scripts
Scripts for development workflow and code maintenance.

- **cleanup-tech-debt-phase*.ps1/.sh** - Technical debt cleanup scripts
- **install-hooks.bat/.sh** - Install git hooks
- **regenerate-graphql.ps1/.sh** - Regenerate GraphQL code
- **format.bat** - Format Go code
- **lint.bat/.sh** - Run linters
- **test.ps1** - Run tests
- **pre-commit** - Git pre-commit hook
- **check-config.go** - Verify configuration

### 📁 docker/ - Docker Scripts
Scripts for managing Docker containers.

- **docker-up.ps1** - Start Docker containers
- **docker-down.ps1** - Stop Docker containers
- **docker-restart.ps1** - Restart Docker containers
- **logs.ps1** - View Docker logs

### 📁 ops/ - Operational Scripts
Scripts for building, deployment, and operations.

- **build.ps1** - Build the application
- **clean.ps1** - Clean build artifacts
- **run.ps1** - Run the application
- **gcp/** - Google Cloud Platform specific scripts

### 📁 verification/ - Email Verification Scripts
Scripts for managing email verification functionality.

- **check-email-verification-status.ps1** - Check user verification status
- **check-token-info.ps1** - Check verification token information
- **cleanup-verification-tokens.ps1** - Clean expired tokens
- **manually-verify-user.ps1** - Manually verify a user
- **verify-email-manual.sql** - SQL for manual verification

### 📁 user-management/ - User Management Scripts
Scripts for user administration and management.

## Usage Examples

### Running Database Migrations
```powershell
# Local development
.\scripts\db\migrate.ps1

# Production
.\scripts\db\run-production-migrations.ps1
```

### Development Workflow
```powershell
# Install git hooks
.\scripts\dev\install-hooks.bat

# Regenerate GraphQL code
.\scripts\dev\regenerate-graphql.ps1

# Run tests
.\scripts\dev\test.ps1
```

### Docker Operations
```powershell
# Start services
.\scripts\docker\docker-up.ps1

# View logs
.\scripts\docker\logs.ps1

# Restart services
.\scripts\docker\docker-restart.ps1
```

## Script Dependencies

Some scripts have dependencies on others:

1. **Database scripts** require Docker to be running
2. **Development scripts** require Go to be installed
3. **GCP scripts** require Google Cloud SDK

## Contributing

When adding new scripts:
1. Place them in the appropriate category folder
2. Add documentation to this README
3. Include usage examples if the script is complex
4. Follow the existing naming conventions
'@

$readmeContent | Set-Content -Path $readmePath -Encoding UTF8
Write-Host "   ✓ README.md creado" -ForegroundColor Green

# 12. Verificar que el proyecto sigue compilando
Write-Host ""
Write-Host "12. Verificando que el proyecto compila..." -ForegroundColor Yellow
Push-Location -Path ".."
try {
    & go build ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ El proyecto compila correctamente" -ForegroundColor Green
    } else {
        Write-Host "   ✗ Error al compilar el proyecto" -ForegroundColor Red
    }
} finally {
    Pop-Location
}

# Resumen final
Write-Host ""
Write-Host "=== Limpieza Fase 3 completada exitosamente ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Resumen de cambios:" -ForegroundColor Yellow
Write-Host "- Scripts reorganizados: $movedCount archivos" -ForegroundColor White
Write-Host "- Archivos eliminados: $deletedCount" -ForegroundColor White
Write-Host "- Nueva estructura de carpetas creada" -ForegroundColor White
Write-Host "- README.md de documentación creado" -ForegroundColor White
Write-Host ""
Write-Host "La nueva estructura de scripts está en:" -ForegroundColor Yellow
Write-Host "  scripts/" -ForegroundColor White
Write-Host "  ├── db/         # Base de datos" -ForegroundColor DarkGray
Write-Host "  ├── dev/        # Desarrollo" -ForegroundColor DarkGray
Write-Host "  ├── docker/     # Docker" -ForegroundColor DarkGray
Write-Host "  ├── ops/        # Operaciones" -ForegroundColor DarkGray
Write-Host "  ├── verification/  # Verificación email" -ForegroundColor DarkGray
Write-Host "  └── README.md   # Documentación" -ForegroundColor DarkGray

#!/bin/bash

# Script para limpiar deuda técnica - Fase 3

echo "=== Limpieza de Deuda Técnica - Fase 3 ==="
echo "    Reorganización de scripts"
echo ""

# Función para mover archivo preservando su estado
move_script_file() {
    local source=$1
    local destination=$2
    
    if [ -f "$source" ]; then
        # Crear directorio destino si no existe
        dest_dir=$(dirname "$destination")
        mkdir -p "$dest_dir"
        
        mv "$source" "$destination"
        echo "   ✓ Movido: $source → $destination"
        return 0
    else
        echo "   ⚠ No encontrado: $source"
        return 1
    fi
}

moved_count=0
base_dir="scripts"

# 1. Crear estructura de carpetas
echo "1. Creando estructura de carpetas..."
folders=("db" "dev" "docker" "ops" "ops/gcp" "verification")
for folder in "${folders[@]}"; do
    path="$base_dir/$folder"
    if [ ! -d "$path" ]; then
        mkdir -p "$path"
        echo "   ✓ Carpeta creada: $path"
    fi
done

# 2. Mover scripts de base de datos
echo ""
echo "2. Moviendo scripts de base de datos..."
db_scripts=(
    "migrate.ps1:db/migrate.ps1"
    "backup-db.ps1:db/backup-db.ps1"
    "restore-db.ps1:db/restore-db.ps1"
    "apply-email-required-changes.ps1:db/apply-email-required-changes.ps1"
    "rollback-email-required-changes.ps1:db/rollback-email-required-changes.ps1"
    "run-email-required-migration.ps1:db/run-email-required-migration.ps1"
    "run-production-migrations.ps1:db/run-production-migrations.ps1"
    "run-production-migrations.sh:db/run-production-migrations.sh"
)

for script in "${db_scripts[@]}"; do
    IFS=':' read -r src dst <<< "$script"
    if move_script_file "$base_dir/$src" "$base_dir/$dst"; then
        ((moved_count++))
    fi
done

# 3. Mover scripts de desarrollo
echo ""
echo "3. Moviendo scripts de desarrollo..."
dev_scripts=(
    "cleanup-tech-debt-phase1.ps1:dev/cleanup-tech-debt-phase1.ps1"
    "cleanup-tech-debt-phase1.sh:dev/cleanup-tech-debt-phase1.sh"
    "cleanup-tech-debt-phase2.ps1:dev/cleanup-tech-debt-phase2.ps1"
    "cleanup-tech-debt-phase2.sh:dev/cleanup-tech-debt-phase2.sh"
    "install-hooks.bat:dev/install-hooks.bat"
    "install-hooks.sh:dev/install-hooks.sh"
    "regenerate-graphql.ps1:dev/regenerate-graphql.ps1"
    "regenerate-graphql.sh:dev/regenerate-graphql.sh"
    "format.bat:dev/format.bat"
    "lint.bat:dev/lint.bat"
    "lint.sh:dev/lint.sh"
    "test.ps1:dev/test.ps1"
    "pre-commit:dev/pre-commit"
    "check-config.go:dev/check-config.go"
)

for script in "${dev_scripts[@]}"; do
    IFS=':' read -r src dst <<< "$script"
    if move_script_file "$base_dir/$src" "$base_dir/$dst"; then
        ((moved_count++))
    fi
done

# 4. Mover scripts de Docker
echo ""
echo "4. Moviendo scripts de Docker..."
docker_scripts=(
    "docker-down.ps1:docker/docker-down.ps1"
    "docker-restart.ps1:docker/docker-restart.ps1"
    "docker-up.ps1:docker/docker-up.ps1"
    "logs.ps1:docker/logs.ps1"
)

for script in "${docker_scripts[@]}"; do
    IFS=':' read -r src dst <<< "$script"
    if move_script_file "$base_dir/$src" "$base_dir/$dst"; then
        ((moved_count++))
    fi
done

# 5. Mover scripts operacionales
echo ""
echo "5. Moviendo scripts operacionales..."
ops_scripts=(
    "build.ps1:ops/build.ps1"
    "clean.ps1:ops/clean.ps1"
    "run.ps1:ops/run.ps1"
)

for script in "${ops_scripts[@]}"; do
    IFS=':' read -r src dst <<< "$script"
    if move_script_file "$base_dir/$src" "$base_dir/$dst"; then
        ((moved_count++))
    fi
done

# 6. Mover scripts de GCP
echo ""
echo "6. Moviendo scripts de GCP..."
gcp_dir="$base_dir/gcp"
if [ -d "$gcp_dir" ]; then
    for file in "$gcp_dir"/*; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            mv "$file" "$base_dir/ops/gcp/$filename"
            echo "   ✓ Movido: gcp/$filename → ops/gcp/$filename"
            ((moved_count++))
        fi
    done
    # Eliminar carpeta gcp vacía
    rmdir "$gcp_dir" 2>/dev/null
fi

# 7. Mover scripts de verificación
echo ""
echo "7. Moviendo scripts de verificación de email..."
verification_scripts=(
    "check-email-verification-status.ps1:verification/check-email-verification-status.ps1"
    "check-token-info.ps1:verification/check-token-info.ps1"
    "cleanup-verification-tokens.ps1:verification/cleanup-verification-tokens.ps1"
    "manually-verify-user.ps1:verification/manually-verify-user.ps1"
    "verify-email-manual.sql:verification/verify-email-manual.sql"
)

for script in "${verification_scripts[@]}"; do
    IFS=':' read -r src dst <<< "$script"
    if move_script_file "$base_dir/$src" "$base_dir/$dst"; then
        ((moved_count++))
    fi
done

# 8. Mover scripts de test a test/manual
echo ""
echo "8. Moviendo scripts de test de email..."
test_scripts=("test-email.go" "test-sendgrid-simple.go" "verify-sendgrid.go" "verify-sendgrid.sh")
test_manual_dir="test/manual"

for script in "${test_scripts[@]}"; do
    source="$base_dir/$script"
    dest="$test_manual_dir/$script"
    if move_script_file "$source" "$dest"; then
        ((moved_count++))
    fi
done

# 9. Eliminar archivos innecesarios
echo ""
echo "9. Eliminando archivos innecesarios..."
to_delete=("go.mod" "go.sum")
deleted_count=0

for file in "${to_delete[@]}"; do
    path="$base_dir/$file"
    if [ -f "$path" ]; then
        rm "$path"
        echo "   ✓ Eliminado: $file"
        ((deleted_count++))
    fi
done

# 10. Mover este script a su ubicación final
echo ""
echo "10. Moviendo este script a su ubicación final..."
this_script="cleanup-tech-debt-phase3.sh"
source="$base_dir/$this_script"
dest="$base_dir/dev/$this_script"
if [ ! -f "$dest" ]; then
    cp "$0" "$dest"
    chmod +x "$dest"
    echo "   ✓ Script copiado a: $dest"
fi

# 11. Crear README.md para scripts
echo ""
echo "11. Creando documentación de scripts..."
readme_path="$base_dir/README.md"
cat > "$readme_path" << 'EOF'
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
```bash
# Local development
./scripts/db/migrate.ps1

# Production
./scripts/db/run-production-migrations.sh
```

### Development Workflow
```bash
# Install git hooks
./scripts/dev/install-hooks.sh

# Regenerate GraphQL code
./scripts/dev/regenerate-graphql.sh

# Run tests
./scripts/dev/test.ps1
```

### Docker Operations
```bash
# Start services
./scripts/docker/docker-up.ps1

# View logs
./scripts/docker/logs.ps1

# Restart services
./scripts/docker/docker-restart.ps1
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
EOF

echo "   ✓ README.md creado"

# 12. Verificar que el proyecto sigue compilando
echo ""
echo "12. Verificando que el proyecto compila..."
cd ..
if go build ./...; then
    echo "   ✓ El proyecto compila correctamente"
else
    echo "   ✗ Error al compilar el proyecto"
fi
cd - > /dev/null

# Resumen final
echo ""
echo "=== Limpieza Fase 3 completada exitosamente ==="
echo ""
echo "Resumen de cambios:"
echo "- Scripts reorganizados: $moved_count archivos"
echo "- Archivos eliminados: $deleted_count"
echo "- Nueva estructura de carpetas creada"
echo "- README.md de documentación creado"
echo ""
echo "La nueva estructura de scripts está en:"
echo "  scripts/"
echo "  ├── db/         # Base de datos"
echo "  ├── dev/        # Desarrollo"
echo "  ├── docker/     # Docker"
echo "  ├── ops/        # Operaciones"
echo "  ├── verification/  # Verificación email"
echo "  └── README.md   # Documentación"

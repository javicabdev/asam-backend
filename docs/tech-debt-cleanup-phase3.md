# Limpieza de Deuda Técnica - Fase 3

## Resumen de Cambios

Esta es la tercera fase de limpieza de deuda técnica, enfocada en la organización de scripts y limpieza final.

### Cambios a Realizar

#### 1. Reorganizar scripts por categorías

**Estructura actual vs nueva**:

```
scripts/
├── db/                    # Scripts relacionados con base de datos
│   ├── migrate.ps1
│   ├── backup-db.ps1
│   ├── restore-db.ps1
│   ├── apply-email-required-changes.ps1
│   ├── rollback-email-required-changes.ps1
│   ├── run-email-required-migration.ps1
│   ├── run-production-migrations.ps1
│   └── run-production-migrations.sh
├── dev/                   # Utilidades de desarrollo
│   ├── cleanup-tech-debt-phase1.ps1
│   ├── cleanup-tech-debt-phase1.sh
│   ├── cleanup-tech-debt-phase2.ps1
│   ├── cleanup-tech-debt-phase2.sh
│   ├── cleanup-tech-debt-phase3.ps1
│   ├── cleanup-tech-debt-phase3.sh
│   ├── install-hooks.bat
│   ├── install-hooks.sh
│   ├── regenerate-graphql.ps1
│   ├── regenerate-graphql.sh
│   ├── format.bat
│   ├── lint.bat
│   ├── lint.sh
│   ├── test.ps1
│   ├── pre-commit
│   └── check-config.go
├── docker/                # Scripts de Docker
│   ├── docker-down.ps1
│   ├── docker-restart.ps1
│   ├── docker-up.ps1
│   └── logs.ps1
├── ops/                   # Scripts operacionales
│   ├── build.ps1
│   ├── clean.ps1
│   ├── run.ps1
│   └── gcp/              # (mover contenido de gcp/ aquí)
├── verification/          # Scripts de verificación de email
│   ├── check-email-verification-status.ps1
│   ├── check-token-info.ps1
│   ├── cleanup-verification-tokens.ps1
│   ├── manually-verify-user.ps1
│   └── verify-email-manual.sql
├── user-management/       # (mantener como está)
└── README.md             # Nueva documentación de scripts
```

#### 2. Scripts a eliminar o mover

**A mover a test/manual/:**
- `test-email.go`
- `test-sendgrid-simple.go`
- `verify-sendgrid.go`
- `verify-sendgrid.sh`

**A evaluar para eliminación:**
- Scripts de email obsoletos (si se confirma que ya no se usan)
- `go.mod` y `go.sum` en scripts/ (no deberían estar ahí)

#### 3. Documentación nueva

Crear `scripts/README.md` con:
- Índice de todos los scripts organizados por categoría
- Descripción de cada script
- Dependencias y orden de ejecución
- Ejemplos de uso

### Impacto

- **Scripts reorganizados**: ~40 archivos
- **Nueva estructura**: 5 carpetas principales + subcarpetas
- **Scripts eliminados/movidos**: 4-6 archivos
- **Documentación nueva**: README completo para scripts

### Prerequisitos

**IMPORTANTE**: Antes de ejecutar este script, asegúrate de:
1. Haber ejecutado las Fases 1 y 2 de limpieza
2. Hacer commit de todos los cambios pendientes
3. Estar en una rama limpia

### Cómo Ejecutar

Para aplicar estos cambios, ejecuta uno de los siguientes scripts desde la raíz del proyecto:

**En Windows (PowerShell):**
```powershell
.\scripts\cleanup-tech-debt-phase3.ps1
```

**En Linux/Mac:**
```bash
chmod +x scripts/cleanup-tech-debt-phase3.sh
./scripts/cleanup-tech-debt-phase3.sh
```

### Validación Post-Ejecución

Después de ejecutar el script:
1. Verifica que todos los scripts estén en sus nuevas ubicaciones
2. Revisa que los paths en la documentación estén actualizados
3. Confirma que los scripts sigan siendo ejecutables
4. Actualiza cualquier referencia a scripts en otros archivos

### Resultado Final

La nueva estructura facilita:
- Encontrar rápidamente el script necesario
- Entender el propósito de cada categoría
- Mantener los scripts organizados
- Onboarding más eficiente para nuevos desarrolladores

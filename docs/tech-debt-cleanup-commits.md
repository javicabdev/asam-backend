# Commits Convencionales - Limpieza de Deuda Técnica

## Resumen de Fases y Commits

Este documento detalla los commits convencionales para cada fase de limpieza de deuda técnica.

---

## Fase 1: Prioridad Alta
**Enfoque**: Violaciones de Clean Architecture y dependencias no utilizadas

### Commit:
```
refactor: remove infrastructure layer and clean unused dependencies

- Remove internal/infrastructure directory (violates Clean Architecture)
- Delete unused smtp_email_service.go implementation
- Run go mod tidy to remove unused dependencies (sqlx, lib/pq)
- Keep email interfaces correctly separated in ports/input and ports/output

BREAKING CHANGE: None - removed code was not being used
```

**Archivos afectados**:
- Eliminados: `internal/infrastructure/` (carpeta completa)
- Modificado: `go.mod`, `go.sum` (después de go mod tidy)

---

## Fase 2: Prioridad Media
**Enfoque**: Organización de código y consolidación de configuración

### Commit:
```
refactor: move mocks to test directory and consolidate env files

- Move mock_notification_adapter.go from internal/adapters/email to test/mocks/email
- Update package from 'email' to 'mocks' in moved file
- Move test-mock-email.go script to test/manual directory
- Consolidate .env files removing redundant configurations:
  - Remove: .env.production.free, .env.production.test
  - Remove: .env.docker.example, .env.email.example
  - Remove: .env.complete.example, .env.local
  - Keep only: .env.example, .env.development, .env.production, .env.test, .env.aiven
  
This improves code organization by keeping test code separate from production code
```

**Archivos afectados**:
- Movido: `internal/adapters/email/mock_notification_adapter.go` → `test/mocks/email/`
- Movido: `scripts/test-mock-email.go` → `test/manual/`
- Eliminados: 6-7 archivos .env redundantes

---

## Fase 3: Prioridad Baja
**Enfoque**: Organización de scripts y limpieza final

### Commit:
```
refactor: organize utility scripts and remove obsolete files

- Create organized directory structure for scripts:
  - db/ for database-related scripts (8 files)
  - dev/ for development utilities (14 files)
  - docker/ for Docker operations (4 files)
  - ops/ for operational scripts including GCP (12 files)
  - verification/ for email verification (5 files)
- Move test email scripts to test/manual directory (4 files)
- Remove unnecessary go.mod and go.sum from scripts directory
- Create comprehensive README.md documentation for all scripts
- Total: ~40 scripts reorganized into logical categories

This improves project maintainability by organizing scripts logically
```

**Archivos afectados**:
- Movidos: ~40 scripts a sus nuevas ubicaciones
- Eliminados: `scripts/go.mod`, `scripts/go.sum`
- Creado: `scripts/README.md` con documentación completa

---

## Orden de Ejecución Recomendado

1. **Hacer commit de cambios pendientes actuales**
   ```bash
   git add .
   git commit -m "feat: add email verification and Apollo compatibility"
   ```

2. **Ejecutar Fase 1**
   ```bash
   ./scripts/cleanup-tech-debt-phase1.ps1  # o .sh
   git add -A
   git commit -m "refactor: remove infrastructure layer and clean unused dependencies"
   ```

3. **Ejecutar Fase 2**
   ```bash
   ./scripts/cleanup-tech-debt-phase2.ps1  # o .sh
   git add -A
   git commit -m "refactor: move mocks to test directory and consolidate env files"
   ```

4. **Ejecutar Fase 3** (cuando esté lista)
   ```bash
   ./scripts/cleanup-tech-debt-phase3.ps1  # o .sh
   git add -A
   git commit -m "refactor: organize utility scripts and remove obsolete files"
   ```

---

## Notas Importantes

- Cada fase debe ejecutarse y hacer commit por separado
- Verificar que los tests pasen después de cada fase
- Si alguna fase falla, es más fácil revertir cambios individuales
- Los commits siguen el estándar [Conventional Commits](https://www.conventionalcommits.org/)

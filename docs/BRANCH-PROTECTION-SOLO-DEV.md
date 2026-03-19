# Branch Protection para Desarrollador Solo

Esta es una guía simplificada de branch protection para proyectos donde **una sola persona** desarrolla.

## ¿Por qué Branch Protection si trabajo solo?

Aunque trabajes solo, branch protection te ayuda a:

1. ✅ **Evitar commits rotos en main**: Tests deben pasar antes de merge
2. ✅ **Forzar buenas prácticas**: Usar ramas feature en lugar de commits directos
3. ✅ **Seguridad automática**: El security scan debe pasar
4. ✅ **Historial limpio**: Documentar cambios via PRs con descripciones

**Lo que NO necesitas**:
- ❌ Múltiples aprobaciones
- ❌ CODEOWNERS
- ❌ Reglas super estrictas
- ❌ Revisiones obligatorias

## Configuración Minimalista Recomendada

### Opción 1: Solo CI/CD (Recomendada para Solo Dev)

Configuración perfecta para trabajar solo pero con calidad:

1. Ve a **Settings** → **Branches** → **Add rule**
2. Branch name pattern: `main`
3. Configura **solo esto**:

```
✅ Require status checks to pass before merging
   └─ Status checks required:
      • lint     (código limpio)
      • security (sin vulnerabilidades)
      • test     (tests pasan)

❌ NO marcar "Require a pull request" (puedes pero no es necesario)
❌ NO marcar "Require approvals"
✅ Do not allow force pushes (previene errores)
```

**Ventajas**:
- Puedes mergear directamente si los checks pasan
- No necesitas aprobaciones
- Todavía tienes seguridad automática
- Workflow rápido para solo dev

**Workflow**:
```bash
# Trabajas en rama feature
git checkout -b feature/nueva-funcionalidad
git commit -m "feat: nueva funcionalidad"
git push origin feature/nueva-funcionalidad

# Crear PR (opcional pero recomendado para documentación)
gh pr create --title "Nueva funcionalidad"

# Si los checks pasan, mergeas (GitHub UI o CLI)
gh pr merge --squash

# O directamente desde la rama si prefieres
git checkout main
git merge feature/nueva-funcionalidad --squash
git push origin main
# ⚠️ Solo funciona si los checks están en green
```

### Opción 2: Sin Branch Protection (Máxima Libertad)

Si prefieres máxima flexibilidad:

```
❌ No configurar branch protection

Pero SÍ confiar en:
✅ CI/CD automático (corre en cada push)
✅ make security antes de push manualmente
✅ make test antes de push manualmente
```

**Ventajas**:
- Total libertad
- Workflow ultra rápido
- Sin restricciones

**Desventajas**:
- Puedes pushear código roto a main por error
- Requiere más disciplina manual

**Workflow**:
```bash
# Antes de push a main
make lint
make security
make test

# Si todo pasa, push directo
git push origin main
```

### Opción 3: Balance (Recomendada si quieres disciplina)

Configuración que te obliga a usar PRs pero sin aprobaciones:

```
✅ Require a pull request before merging
   └─ Required approvals: 0  ← Sin aprobaciones necesarias
   └─ ✅ Allow specified actors to bypass (tú mismo)

✅ Require status checks to pass before merging
   └─ Status checks required: lint, security, test

✅ Do not allow force pushes
```

**Ventajas**:
- Te obliga a usar PRs (mejor documentación)
- No necesitas aprobarte a ti mismo
- Checks deben pasar
- Puedes mergear inmediatamente si checks pasan

**Workflow**:
```bash
git checkout -b feature/mi-feature
git push origin feature/mi-feature
gh pr create --title "Mi feature" --fill
# Esperar a que checks pasen
gh pr merge --squash  # No necesita aprobación
```

## Configuración Práctica para Este Proyecto

Dado que tienes SAST (gosec) y DAST (ZAP) configurados:

### Configuración Mínima Efectiva

```yaml
Branch: main

✅ Require status checks to pass before merging
   Required checks:
   - setup     # Genera código GraphQL
   - lint      # golangci-lint
   - security  # gosec (SAST)
   - test      # Unit + Integration tests

   ⚠️ NO marcar "Require branches to be up to date"
   (Esto es tedioso para solo dev, puedes rebase si quieres)

❌ Require pull request: NO
   (Puedes crear PRs cuando quieras documentar, pero no son obligatorios)

❌ Require approvals: NO
   (Eres solo tú)

✅ Do not allow force pushes: SÍ
   (Previene errores)

❌ Do not allow bypassing: NO
   (Eres admin, necesitas poder saltarte reglas en emergencias)
```

## Workflows Simplificados

### Workflow 1: Feature pequeña (rápida)

```bash
# Crear rama
git checkout -b fix/pequeño-bug

# Hacer cambios
git add .
git commit -m "fix: corrección de bug"

# Push
git push origin fix/pequeño-bug

# Mergear directo desde GitHub UI si checks pasan
# O desde CLI
gh pr create --fill
gh pr merge --squash
```

### Workflow 2: Experimentación

```bash
# Trabajas directo en rama de experimento
git checkout -b experiment/nueva-idea

# Muchos commits experimentales
git commit -am "wip: probando esto"
git commit -am "wip: probando aquello"

# Cuando funciona, limpias
git rebase -i main  # Squash todos los commits
git push origin experiment/nueva-idea

# Mergeas
gh pr create --title "Nueva idea implementada"
gh pr merge --squash
```

### Workflow 3: Hotfix urgente

```bash
# Si branch protection está activado pero necesitas hotfix urgente

# Opción A: Desactivar temporalmente
# Settings → Branches → Edit rule → Desactivar
# Hacer hotfix
# Reactivar

# Opción B: PR ultra rápido
git checkout -b hotfix/critical
git commit -am "fix: critical security issue"
git push
gh pr create --fill
# Esperar 2-3 min a que pasen checks
gh pr merge --squash

# Opción C: Si configuraste bypass para ti
# Push directo, checks corren pero no bloquean
```

## Comandos Útiles para Solo Dev

```bash
# Ver estado de checks del último commit
gh pr checks

# Mergear sin esperar (si configuraste bypass)
gh pr merge --admin --squash

# Saltar checks en emergencia (no recomendado)
git push --no-verify  # Solo salta pre-commit hooks locales

# Ver ramas con checks fallidos
gh pr list --state open --json number,title,statusCheckRollup

# Limpiar ramas mergeadas
git branch --merged | grep -v "main" | xargs git branch -d
```

## Recomendación Final para Tu Proyecto

**Configuración Sugerida**:

```
✅ Require status checks: lint, security, test
❌ Require PR: NO (usa PRs cuando quieras, no obligatorio)
❌ Require approvals: NO
✅ Do not allow force pushes: SÍ
```

**Razones**:
1. Mantienes calidad (checks automáticos)
2. Workflow rápido (no necesitas PR para todo)
3. Flexibilidad (puedes saltarte si es urgente)
4. Seguridad (gosec debe pasar)

## Automatización Adicional para Solo Dev

### Pre-commit hooks (Opcional)

Crea `.git/hooks/pre-push`:

```bash
#!/bin/bash
# Corre checks antes de push

echo "🔍 Running checks before push..."

# Lint
make lint || exit 1

# Security
make security || exit 1

# Tests
make test || exit 1

echo "✅ All checks passed, pushing..."
```

Haz ejecutable:
```bash
chmod +x .git/hooks/pre-push
```

Ahora antes de cada push se corren los checks localmente.

### GitHub CLI aliases

Agrega a `~/.gitconfig`:

```ini
[alias]
    quick-pr = !gh pr create --fill && gh pr merge --squash
    force-merge = !gh pr merge --admin --squash
```

Uso:
```bash
git quick-pr    # Crear PR y mergear en un comando
git force-merge # Mergear aunque checks fallen (emergencias)
```

## Cuándo SÍ usar configuración estricta

Usa configuración estricta (aprobaciones, etc.) si:

- ❌ El proyecto crece y contratas/agregas colaboradores
- ❌ Es un proyecto open source con contribuidores externos
- ❌ Necesitas auditoría estricta (compliance, regulaciones)
- ❌ El proyecto es crítico (infraestructura, finanzas, salud)

Para tu proyecto actual (backend de gestión):
- ✅ Configuración minimalista es suficiente
- ✅ Confía en los checks automáticos
- ✅ Mantén flexibilidad para iterar rápido

## Resumen

**Para desarrollador solo**:

```
┌─────────────────────────────────────────┐
│  Configuración Minimalista              │
├─────────────────────────────────────────┤
│ ✅ Status checks required (automated)  │
│ ❌ Manual approvals (no needed)        │
│ ✅ Force push prevention (safety)      │
│ ⚡ Fast workflow                        │
│ 🛡️ Security via automation             │
└─────────────────────────────────────────┘
```

¿Quieres que actualice la configuración recomendada en la documentación principal para reflejar esto?

# Guía de Configuración de Branch Protection

Esta guía te ayudará a configurar branch protection rules en GitHub para asegurar que el código que llega a `main` cumpla con todos los estándares de calidad y seguridad.

## ¿Qué es Branch Protection?

Branch Protection es una funcionalidad de GitHub que permite establecer reglas sobre qué puede fusionarse (merge) en ramas específicas. Esto previene:

- Merges directos sin revisión
- Código que no pasa los tests
- Código con vulnerabilidades de seguridad
- Commits sin firma
- Force pushes accidentales

## Configuración Recomendada para `main`

### Paso 1: Acceder a Branch Protection

1. Ve a tu repositorio en GitHub
2. Haz clic en **Settings** (Configuración)
3. En el menú lateral, selecciona **Branches** (Ramas)
4. En "Branch protection rules", haz clic en **Add rule** (Agregar regla)
5. En "Branch name pattern", escribe: `main`

### Paso 2: Configurar Reglas Básicas

#### ✅ Require a pull request before merging
**Descripción**: Obliga a que todo código pase por un Pull Request antes de llegar a main.

**Configuración**:
- ✅ Marcar checkbox
- **Opciones adicionales**:
  - ✅ `Require approvals`: Mínimo 1 aprobación
  - ✅ `Dismiss stale pull request approvals when new commits are pushed`
  - ✅ `Require review from Code Owners` (si tienes un archivo CODEOWNERS)

**Razón**: Previene merges accidentales y asegura revisión de código.

#### ✅ Require status checks to pass before merging
**Descripción**: Obliga a que todos los checks (tests, linting, security) pasen antes del merge.

**Configuración**:
- ✅ Marcar checkbox
- ✅ `Require branches to be up to date before merging`

**Status checks a requerir**:
- `setup` - Setup Environment & Generate Code
- `lint` - Lint Code
- `security` - Security Scan (SAST)
- `test` - Run Tests

Para agregar checks:
1. Haz un commit para disparar el workflow
2. Los checks aparecerán en la lista
3. Búscalos por nombre y márcalos

**Razón**: Asegura que el código funciona y es seguro antes de merge.

#### ✅ Require conversation resolution before merging
**Descripción**: Obliga a resolver todos los comentarios en el PR antes de merge.

**Configuración**:
- ✅ Marcar checkbox

**Razón**: Asegura que todos los problemas identificados sean abordados.

### Paso 3: Configurar Reglas de Seguridad (Opcional pero Recomendado)

#### 🔒 Require signed commits
**Descripción**: Solo permite commits firmados con GPG/SSH.

**Configuración**:
- ✅ Marcar checkbox si tu equipo usa firma de commits

**Razón**: Verifica la identidad del autor del commit.

#### 🔒 Require linear history
**Descripción**: Previene merge commits, obliga a usar rebase o squash.

**Configuración**:
- ✅ Marcar checkbox (recomendado para mantener historial limpio)

**Razón**: Historial de git más limpio y fácil de seguir.

### Paso 4: Configurar Restricciones Administrativas

#### 🚫 Do not allow bypassing the above settings
**Descripción**: Ni siquiera los admins pueden saltarse las reglas.

**Configuración**:
- ✅ Marcar checkbox

**Razón**: Asegura que TODOS sigan las reglas, incluso administradores.

#### 🚫 Do not allow force pushes
**Descripción**: Previene `git push --force` en main.

**Configuración**:
- ✅ Marcar checkbox

**Razón**: Previene reescritura accidental del historial.

#### 🚫 Allow deletions
**Descripción**: Previene borrado accidental de la rama.

**Configuración**:
- ❌ **NO** marcar checkbox (queremos prevenir borrado)

**Razón**: Protege contra borrado accidental de main.

## Configuración Completa Paso a Paso

### 1. Branch name pattern
```
main
```

### 2. Protect matching branches
```
✅ Require a pull request before merging
   └─ Required approvals: 1
   └─ ✅ Dismiss stale pull request approvals when new commits are pushed
   └─ ✅ Require review from Code Owners (si aplica)

✅ Require status checks to pass before merging
   └─ ✅ Require branches to be up to date before merging
   └─ Status checks that are required:
      • setup
      • lint
      • security
      • test

✅ Require conversation resolution before merging

✅ Require signed commits (opcional)

✅ Require linear history (recomendado)

✅ Do not allow bypassing the above settings

✅ Do not allow force pushes

❌ Allow deletions (NO marcar)
```

### 3. Rules applied to everyone including administrators
```
✅ Include administrators
```

## Verificar Configuración

Una vez configurado, intenta:

1. **Hacer push directo a main** → Debería ser rechazado
   ```bash
   git push origin main
   # Error: protected branch
   ```

2. **Crear PR sin pasar tests** → No se puede mergear
   - Los checks deben pasar primero

3. **Crear PR y aprobar inmediatamente** → Debería poder mergearse
   - Si tienes 1 aprobación y todos los checks pasan

## Troubleshooting

### Error: "Required status check is not available"

**Problema**: Agregaste un check que no existe en el workflow.

**Solución**:
1. Verifica nombres exactos en `.github/workflows/ci.yml`
2. Los nombres deben coincidir con `jobs.<job-id>.name`

### Error: "Branch is out of date"

**Problema**: La rama PR no tiene los últimos commits de main.

**Solución**:
```bash
git checkout your-feature-branch
git rebase origin/main
git push --force-with-lease
```

### No puedo mergear aunque todo pasó

**Problema**: Tienes conversaciones sin resolver.

**Solución**: Resuelve todos los comentarios en el PR.

## Excepciones y Bypasses

### Hotfix Urgente

Si necesitas hacer un hotfix urgente y saltarte las reglas:

**Opción 1: Crear regla de bypass temporal**
1. Settings → Branches
2. Edita la regla de `main`
3. Desmarca temporalmente las reglas necesarias
4. Haz el hotfix
5. **Reestablece las reglas inmediatamente**

**Opción 2: Usar rama de emergencia**
```bash
# Crear rama de emergencia sin protección
git checkout -b hotfix-emergency
# Hacer cambios
git commit -am "hotfix: critical security patch"
git push origin hotfix-emergency
# Mergear rápidamente
# Después pasar por proceso normal
```

## Configuración Avanzada

### CODEOWNERS

Crea `.github/CODEOWNERS` para requerir aprobación de dueños específicos:

```
# Cambios en seguridad requieren aprobación de security team
/pkg/auth/** @security-team
/pkg/monitoring/** @security-team

# Cambios en CI/CD requieren aprobación de devops
/.github/workflows/** @devops-team

# Todo lo demás requiere aprobación de cualquier maintainer
* @maintainers
```

### Rulesets (Beta)

GitHub Rulesets es la nueva forma de configurar protección de ramas:

**Ventajas**:
- Aplicar reglas a múltiples ramas
- Más granularidad
- Mejor UI

**Cómo acceder**:
1. Settings → Rules → Rulesets
2. Create ruleset
3. Configurar igual que branch protection

## Mejores Prácticas

### 1. Empezar Estricto
```
Mejor: Empezar con reglas estrictas y relajar si es necesario
Evitar: Empezar permisivo e intentar restringir después
```

### 2. Documentar Excepciones
Si desactivas una regla temporalmente:
```markdown
## Cambios en Branch Protection
- **Fecha**: 2025-01-15
- **Razón**: Hotfix crítico de seguridad
- **Reglas desactivadas**: Required reviews
- **Duración**: 1 hora
- **Restaurado por**: @usuario
```

### 3. Revisar Regularmente
- Cada 3 meses: Revisar si las reglas siguen siendo apropiadas
- Ajustar según el tamaño del equipo
- Actualizar checks requeridos cuando cambies workflows

### 4. Comunicar al Equipo
```markdown
# Anuncio: Branch Protection Activado

A partir del 2025-01-20, `main` estará protegido:

✅ PRs obligatorios
✅ 1 aprobación mínima
✅ Tests deben pasar
✅ Security scan debe pasar

Por favor actualicen su flujo de trabajo.
```

## Monitoreo

### Ver intentos de bypass

GitHub logs:
1. Settings → Audit log
2. Filtrar por: `action:protected_branch.*`

### Métricas a seguir

- Número de PRs mergeados/semana
- Tiempo promedio de review
- Número de rechazos por security checks
- Número de rechazos por tests fallidos

## Referencias

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [Status Checks](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks)
- [CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)

## Resumen de Comandos

```bash
# Ver reglas actuales (requiere gh CLI)
gh api repos/:owner/:repo/branches/main/protection

# Crear PR en lugar de push directo
git checkout -b feature/my-feature
git push origin feature/my-feature
gh pr create --title "My Feature" --body "Description"

# Actualizar rama antes de merge
git checkout feature/my-feature
git rebase origin/main
git push --force-with-lease
```

---

**Nota**: Esta configuración asume que ya tienes los workflows de CI configurados (ver commit anterior con SAST/DAST).

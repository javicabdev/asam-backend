# Guía para Habilitar GitHub Security Tab

Esta guía te ayudará a habilitar y configurar GitHub Security para visualizar las alertas de seguridad generadas por gosec (SAST).

## Paso 1: Habilitar GitHub Advanced Security (si es necesario)

### Para repositorios públicos:
GitHub Advanced Security (incluyendo Code Scanning) está **GRATIS y habilitado automáticamente** en repositorios públicos.

### Para repositorios privados:
Necesitas una licencia de GitHub Advanced Security. Si no la tienes:
1. Ve a **Settings** del repositorio
2. En el menú lateral, selecciona **Code security and analysis**
3. Si ves la opción, habilita **GitHub Advanced Security**

## Paso 2: Habilitar Code Scanning

1. Ve a tu repositorio en GitHub
2. Haz clic en la pestaña **Security** (en la parte superior del repo)
3. En el panel lateral izquierdo, haz clic en **Code scanning**
4. Haz clic en **Set up code scanning**
5. Selecciona **Advanced** (no uses el Default setup, ya que tenemos nuestro propio workflow)

**Nota**: Si ya tienes el workflow configurado (como en este proyecto), GitHub detectará automáticamente el workflow `ci.yml` que sube reportes SARIF.

## Paso 3: Verificar configuración del workflow

Ya hemos configurado el workflow correctamente en `.github/workflows/ci.yml`:

```yaml
permissions:
  contents: read
  security-events: write  # ← Necesario para subir SARIF
  actions: read

jobs:
  security:
    name: Security Scan (SAST)
    steps:
      - name: Run gosec security scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: gosec-results.sarif
```

## Paso 4: Hacer push de los cambios

Para activar GitHub Security, necesitas hacer push de los cambios del workflow:

```bash
git add .github/workflows/ci.yml
git commit -m "feat: enable GitHub Security with SARIF upload"
git push origin main
```

## Paso 5: Esperar a que se ejecute el workflow

1. Ve a la pestaña **Actions** en GitHub
2. Espera a que el workflow "Continuous Integration" termine
3. Verifica que el job **security** se complete exitosamente

## Paso 6: Ver las alertas en GitHub Security

Una vez que el workflow se complete:

1. Ve a la pestaña **Security** de tu repositorio
2. En el menú lateral, haz clic en **Code scanning**
3. Deberías ver las alertas generadas por gosec

### Estructura de alertas:

```
Security > Code scanning > Alerts

┌────────────────────────────────────────┐
│ Tool: gosec                            │
│ Status: Open / Fixed / Dismissed       │
│ Severity: Critical / High / Medium     │
├────────────────────────────────────────┤
│ Alert 1: Potential file inclusion     │
│ File: pkg/monitoring/memory.go:163    │
│ CWE-22: Improper Input Validation     │
│ [View Details] [Dismiss]               │
└────────────────────────────────────────┘
```

## Paso 7: Configurar notificaciones (opcional)

Para recibir notificaciones de nuevas alertas:

1. Ve a **Settings** del repositorio
2. Navega a **Code security and analysis**
3. Bajo **Code scanning**, configura:
   - **Alert notifications**: Email o GitHub notifications
   - **Failed analysis notifications**: Para saber si el análisis falla

## Verificar que funciona

### Opción 1: Ver en GitHub UI
```
https://github.com/[tu-usuario]/asam-backend/security/code-scanning
```

### Opción 2: Usar GitHub CLI
```bash
# Instalar GitHub CLI si no lo tienes
# https://cli.github.com/

# Ver alertas de code scanning
gh api repos/:owner/:repo/code-scanning/alerts
```

### Opción 3: Ver en el workflow
Después de cada ejecución del workflow, deberías ver:
```
✓ Run gosec security scanner
✓ Upload SARIF file
  Code scanning results uploaded successfully
```

## Troubleshooting

### Error: "Advanced Security must be enabled"
**Solución**: Si es un repositorio privado, necesitas habilitar GitHub Advanced Security primero (Paso 1).

### Error: "Resource not accessible by integration"
**Solución**: Verifica que el workflow tenga los permisos correctos:
```yaml
permissions:
  security-events: write
```

### No veo alertas en Security tab
**Posibles causas**:
1. El workflow no se ha ejecutado todavía → Espera o ejecuta manualmente
2. No hay vulnerabilidades → ¡Excelente! Tu código está limpio
3. Las alertas están filtradas → Revisa los filtros en la UI

### El archivo SARIF no se sube
**Solución**: Verifica que:
1. El archivo `gosec-results.sarif` se genera correctamente
2. La acción `github/codeql-action/upload-sarif@v3` está actualizada
3. Los permisos están configurados correctamente

## Características adicionales

### 1. Filtrar alertas
En la página de Code scanning, puedes filtrar por:
- Severidad (Critical, High, Medium, Low)
- Estado (Open, Fixed, Dismissed)
- Rama (main, develop, etc.)
- Tool (gosec, CodeQL, etc.)

### 2. Dismissar falsos positivos
Si una alerta es un falso positivo:
1. Haz clic en la alerta
2. Selecciona **Dismiss alert**
3. Elige la razón (False positive, Won't fix, Used in tests)
4. Añade un comentario explicativo

### 3. Integración con Pull Requests
Las nuevas alertas aparecerán automáticamente en los PRs:
```
Security
  ⚠️ gosec found 1 new alert

  Potential file inclusion via variable
  pkg/monitoring/memory.go:163
```

### 4. Webhooks y API
Puedes configurar webhooks para alertas de seguridad:
```
Settings > Webhooks > Add webhook
Events: Code scanning alerts
```

## Próximos pasos

Una vez habilitado GitHub Security:

1. **Revisa todas las alertas** y categorízalas
2. **Corrige las vulnerabilidades reales**
3. **Documenta los falsos positivos** con comentarios
4. **Configura branch protection** para requerir que el security scan pase
5. **Monitorea regularmente** las nuevas alertas

## Referencias

- [GitHub Code Scanning Documentation](https://docs.github.com/en/code-security/code-scanning)
- [SARIF Format Specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
- [gosec GitHub Action](https://github.com/securego/gosec)
- [CodeQL Action](https://github.com/github/codeql-action)

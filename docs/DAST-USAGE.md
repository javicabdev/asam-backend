# Guía de DAST (Dynamic Application Security Testing)

Esta guía explica cómo usar el análisis de seguridad dinámico con OWASP ZAP en este proyecto.

## ¿Qué es DAST?

DAST (Dynamic Application Security Testing) es un "pentester automático" que ejecuta tu aplicación y la ataca desde fuera, como lo haría un hacker real. **No tiene acceso al código fuente** (análisis de caja negra).

### Diferencia con SAST:
- **SAST (gosec)**: Analiza el código fuente sin ejecutarlo
- **DAST (OWASP ZAP)**: Ataca la aplicación en ejecución

## Tipos de Escaneo Disponibles

### 1. Baseline Scan (Rápido - 2-3 minutos)
- Escaneo pasivo de la aplicación
- No envía ataques activos
- Identifica problemas obvios
- **Ideal para**: CI/CD, verificaciones rápidas

### 2. Full Scan (Completo - 10+ minutos)
- Escaneo activo con ataques reales
- Prueba inyecciones, XSS, etc.
- Más exhaustivo
- **Ideal para**: Antes de releases importantes

### 3. API Scan (Específico para GraphQL)
- Optimizado para APIs
- Prueba endpoints GraphQL
- Valida autenticación y autorización
- **Ideal para**: Testing de GraphQL API

## Cómo Ejecutar DAST

### Opción 1: GitHub Actions (Recomendado)

#### A. Manualmente desde GitHub UI

1. Ve a la pestaña **Actions** en GitHub
2. Selecciona el workflow **"DAST Security Scan"**
3. Haz clic en **"Run workflow"**
4. Configura:
   - **Target URL**: URL de tu aplicación en staging/producción
   - **Scan Type**: baseline, full, o api
5. Haz clic en **"Run workflow"**

#### B. Automáticamente (Programado)

El workflow se ejecuta automáticamente **cada domingo a las 2 AM** contra la URL de producción.

### Opción 2: Localmente con Docker

```bash
# Baseline scan
docker run -t zaproxy/zap-stable zap-baseline.py \
  -t https://your-app-url.com \
  -r zap-report.html

# Full scan
docker run -t zaproxy/zap-stable zap-full-scan.py \
  -t https://your-app-url.com \
  -r zap-report.html

# API scan (GraphQL)
docker run -t zaproxy/zap-stable zap-api-scan.py \
  -t https://your-app-url.com/graphql \
  -f openapi \
  -r zap-report.html
```

## Interpretar Resultados

Los resultados de ZAP se clasifican por nivel de riesgo:

### 🔴 High (Alto)
- **Urgencia**: Crítico - Arreglar inmediatamente
- **Ejemplos**: SQL Injection, XSS, Command Injection
- **Acción**: Bloquear deploy hasta corregir

### 🟡 Medium (Medio)
- **Urgencia**: Importante - Arreglar pronto
- **Ejemplos**: Headers de seguridad faltantes, cookies sin secure flag
- **Acción**: Corregir antes del próximo release

### 🟢 Low (Bajo)
- **Urgencia**: Informativo
- **Ejemplos**: Information disclosure, version leakage
- **Acción**: Revisar y considerar

### 🔵 Informational
- **Urgencia**: Informativo
- **Ejemplos**: Recomendaciones generales
- **Acción**: Opcional

## Configuración Avanzada

### Archivo de Reglas (`.zap/rules.tsv`)

Puedes personalizar cómo ZAP trata ciertos hallazgos:

```tsv
# Ignorar X-Frame-Options faltante (si usas CSP)
10020	IGNORE

# Reportar como WARNING en lugar de FAIL
10015	WARN
```

### Formatos disponibles:
- `IGNORE`: No reportar
- `INFO`: Solo informativo
- `WARN`: Advertencia (no falla el build)
- `FAIL`: Falla el build

## Workflow de Seguridad Recomendado

```
┌─────────────┐
│  Developer  │
│  escribe    │
│   código    │
└──────┬──────┘
       │
       ↓
┌─────────────┐
│    SAST     │  ← gosec (CI/CD)
│  (gosec)    │
└──────┬──────┘
       │ Pass
       ↓
┌─────────────┐
│   Tests     │  ← Unit + Integration
└──────┬──────┘
       │ Pass
       ↓
┌─────────────┐
│  Deploy to  │
│   Staging   │
└──────┬──────┘
       │
       ↓
┌─────────────┐
│    DAST     │  ← OWASP ZAP
│ (OWASP ZAP) │
└──────┬──────┘
       │ Pass
       ↓
┌─────────────┐
│ Production  │
└─────────────┘
```

## Falsos Positivos Comunes

### 1. Headers de Seguridad Faltantes
**Alerta**: X-Frame-Options, X-Content-Type-Options faltantes

**Razón**: Si estás detrás de un CDN (Cloudflare), estos headers podrían estar configurados ahí.

**Acción**: Verificar en producción con:
```bash
curl -I https://your-app.com
```

### 2. Content Security Policy (CSP)
**Alerta**: CSP no configurado

**Razón**: APIs GraphQL no siempre necesitan CSP (es más relevante para frontend).

**Acción**: Considerar si es necesario para tu caso de uso.

### 3. Cookie sin Secure Flag
**Alerta**: Cookies sin atributo Secure

**Razón**: Cookies deben tener `Secure` en producción HTTPS.

**Acción**: Verificar configuración de cookies en pkg/auth.

## Integración con otros Escaneos

### SAST + DAST = Cobertura Completa

| Tipo | Encuentra | Ejemplo |
|------|-----------|---------|
| **SAST** | Vulnerabilidades en código | `sql.Exec("SELECT * FROM users WHERE id = " + id)` |
| **DAST** | Vulnerabilidades en runtime | Inyección SQL que SAST no detectó por estar en query compleja |

Ambos son necesarios para seguridad completa.

## Troubleshooting

### Error: "Target is not accessible"
**Solución**: Verifica que la URL es accesible públicamente. Si es staging, asegúrate de que no esté detrás de firewall/VPN.

### Error: "Scan timed out"
**Solución**: La aplicación podría estar muy lenta o no responder. Verifica que esté corriendo.

### Demasiados falsos positivos
**Solución**: Personaliza `.zap/rules.tsv` para ignorar alertas específicas.

## Mejores Prácticas

1. **Ejecuta DAST regularmente**
   - Mínimo: Antes de cada release a producción
   - Recomendado: Semanalmente en staging

2. **No ejecutes Full Scan en producción**
   - Usa Baseline en producción
   - Usa Full Scan solo en staging/QA

3. **Revisa todos los High y Medium**
   - No ignores alertas sin investigar
   - Documenta por qué ignoras algo

4. **Combina con SAST**
   - SAST encuentra ~70% de vulnerabilidades
   - DAST encuentra ~30% que SAST no puede
   - Juntos = ~90% cobertura

5. **Monitorea tendencias**
   - ¿Están aumentando las vulnerabilidades?
   - ¿Se están corrigiendo rápidamente?

## Recursos

- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP ZAP GitHub Actions](https://github.com/marketplace/actions/owasp-zap-baseline-scan)

## Siguientes Pasos

Una vez que DAST esté funcionando:

1. **Configura alertas** para alertas High/Critical
2. **Integra con Slack/Email** para notificaciones
3. **Haz obligatorio el DAST** en branch protection antes de merge a main
4. **Considera un pentest manual** una vez al año por expertos

# Security Policy

## SAST (Static Application Security Testing)

Este proyecto implementa análisis de seguridad estático utilizando **gosec**, una herramienta especializada en encontrar vulnerabilidades de seguridad en código Go.

### ¿Qué es SAST?

SAST (Static Application Security Testing) es un "linter de seguridad" que analiza tu código fuente **sin ejecutarlo**. Busca patrones de código peligrosos conocidos (análisis de caja blanca).

### ¿Cuándo se ejecuta?

1. **En CI/CD**: Automáticamente en cada push y pull request a `main`
2. **Localmente**: Ejecutando `make security`
3. **Durante desarrollo**: Puedes integrarlo en tu editor/IDE

### Comandos disponibles

```bash
# Ejecutar análisis de seguridad completo
make security

# Ejecutar análisis de seguridad para CI/CD (formato SARIF)
make security-ci

# Instalar herramientas (incluye gosec)
make tools
```

### Lo que encuentra gosec

- **Uso de funciones débiles** (ej. `math/rand` para criptografía en lugar de `crypto/rand`)
- **Credenciales hardcodeadas** en el código
- **Posibles inyecciones SQL** (si detecta concatenación de strings en queries)
- **Configuraciones TLS inseguras** (como `InsecureSkipVerify: true`)
- **Path traversal** (uso de rutas de archivos sin validación)
- **Integer overflow** (conversiones de tipos que pueden causar desbordamiento)

### Configuración

El archivo `.gosec.json` configura:
- **Exclusiones**: Código generado automáticamente y test utilities
- **Severidad mínima**: MEDIUM
- **Confianza mínima**: MEDIUM
- **Reglas excluidas**:
  - `G115`: Integer overflow en código generado
  - `G404`: Weak random en test generators (aceptable para testing)
  - `G101`: False positives de credenciales (constantes de tipo)

### Interpretación de resultados

#### Severidad
- **HIGH**: Vulnerabilidad crítica que debe corregirse inmediatamente
- **MEDIUM**: Vulnerabilidad importante que debe revisarse
- **LOW**: Problema menor o informativo

#### Confianza
- **HIGH**: Muy probable que sea un problema real
- **MEDIUM**: Puede requerir revisión manual
- **LOW**: Probable falso positivo

### Integración con GitHub

Los resultados se suben automáticamente a **GitHub Security** en formato SARIF, donde puedes:
- Ver alertas de seguridad en la pestaña "Security"
- Recibir notificaciones de nuevas vulnerabilidades
- Hacer seguimiento de vulnerabilidades corregidas

### Buenas prácticas

1. **Ejecuta `make security` antes de cada commit importante**
2. **Revisa las alertas nuevas** en tus Pull Requests
3. **No ignores alertas sin documentación**: Si necesitas ignorar una alerta, usa `//nolint:gosec // Razón clara`
4. **Mantén actualizado gosec**: `go install github.com/securego/gosec/v2/cmd/gosec@latest`

### Reportar una vulnerabilidad

Si encuentras una vulnerabilidad de seguridad en este proyecto, por favor repórtala de forma privada a través de:
- GitHub Security Advisories
- Email directo al equipo de desarrollo

**NO** abras issues públicos para vulnerabilidades de seguridad.

---

## Estado actual de seguridad

Último scan: Consultar el último workflow de GitHub Actions
Vulnerabilidades conocidas: Ninguna (código de producción limpio)

### Exclusiones documentadas

- **Código generado** (`internal/adapters/gql/generated/`): Generado automáticamente por gqlgen
- **Test generators** (`test/seed/generators/`): Uso de `math/rand` es aceptable para datos de prueba
- **File operations** (`pkg/monitoring/memory.go`): Rutas controladas y validadas

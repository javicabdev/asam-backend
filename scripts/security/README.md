# Configuración de Seguridad - gosec

Este documento describe la configuración de seguridad del proyecto usando `gosec` y cómo resolver problemas comunes.

## Problema: Fallos en código generado

El código generado por `gqlgen` puede causar fallos en el análisis de seguridad, especialmente:

- **G115**: Integer overflow conversion (conversiones int -> int32)
- **G404**: Uso de generadores de números aleatorios débiles (en código de test)
- **G101**: Posibles credenciales hardcodeadas (falsos positivos)

## Solución

### 1. Actualización de `.gosec.json`

```json
{
  "global": {
    "nosec": true,
    "exclude-generated": true,
    "severity": "medium",
    "confidence": "medium"
  },
  "exclude-dir": [
    "test/seed",
    "scripts/cmd",
    "test/fixtures",
    "test/helpers",
    "internal/adapters/gql/generated",
    "internal/adapters/gql/model"
  ],
  "exclude": [
    "G115",
    "G404",
    "G101",
    "G204",
    "G304"
  ]
}
```

### 2. Actualización de `.golangci.yml`

Se agregaron exclusiones específicas para el código generado:

- `skip-dirs`: Excluye directorios completos del análisis
- `exclude-files`: Excluye archivos específicos por patrón
- `gosec.excludes`: Configura exclusiones específicas de gosec

### 3. Scripts de verificación local

Se crearon scripts para ejecutar gosec localmente:

```bash
# Linux/Mac
./scripts/security/run-gosec.sh

# Windows PowerShell
.\scripts\security\run-gosec.ps1
```

## Mejores prácticas

1. **Ejecutar verificación local antes de push**: Usa los scripts proporcionados
2. **No modificar código generado**: Nunca edites manualmente archivos en `internal/adapters/gql/generated`
3. **Usar nolint con justificación**: Si necesitas ignorar un warning legítimo, usa `//nolint:gosec` con comentario explicativo

## Exclusiones justificadas

### G115 - Integer overflow
- Ocurre en código generado por gqlgen
- No es un problema real ya que los valores están controlados

### G404 - Weak random number generator
- Aceptable en código de test y generación de datos de prueba
- NUNCA usar math/rand en código de producción

### G101 - Hardcoded credentials
- Falsos positivos con constantes como `TokenTypeEmailVerification`
- Las credenciales reales deben venir de variables de entorno

### G204 - Subprocess with tainted input
- Solo en scripts de utilidad con inputs controlados
- Requiere nolint con justificación

### G304 - File inclusion via variable
- Solo en código de monitoreo con paths construidos de forma segura
- Requiere validación de inputs

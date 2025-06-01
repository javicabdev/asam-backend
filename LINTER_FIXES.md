# Correcciones del Linter

## Problemas Encontrados y Solucionados

### 1. Errores del Linter en main.go

**Problema 1**: Variables exportadas sin comentarios
```go
// Error: exported var Commit should have comment or be unexported
Commit = "unknown"
```

**Solución**: Añadidos comentarios a todas las variables exportadas
```go
// Version is the application version (set by build flags)
Version = "unknown"
// Commit is the git commit hash (set by build flags)
Commit = "unknown"
// BuildTime is the build timestamp (set by build flags)
BuildTime = "unknown"
```

**Problema 2**: Uso de string básico como clave de contexto
```go
// Error: should not use basic type untyped string as key in context.WithValue
ctx := context.WithValue(r.Context(), "requestID", requestID)
```

**Solución**: Creado tipo personalizado para claves de contexto
```go
// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
    // requestIDKey is the context key for request ID
    requestIDKey contextKey = "requestID"
)

// Uso correcto
ctx := context.WithValue(r.Context(), requestIDKey, requestID)
```

### 2. Error de configuración con typecheck

**Problema**: `typecheck is not a linter, it cannot be enabled or disabled`

**Solución**: 
- Eliminado `typecheck` de la lista de linters (no es un linter válido en v2.1.6)
- Actualizada la configuración de `.golangci.yml` para usar la sintaxis correcta:
  - `skip-dirs` y `skip-files` en lugar de estructuras anidadas
  - `exclude-rules` en lugar de `exclusions`
  - Estructura simplificada y compatible con v2.1.6

### 3. Actualización a las versiones más recientes

**Versiones utilizadas**:

1. **golangci-lint v2.1.6** (última versión)
   - GitHub Actions usando golangci-lint-action@v8
   - Configuración en `.golangci.yml` optimizada para v2.1.6

2. **Go 1.24** (última versión estable)
   - Actualizado en todos los archivos:
     - `Dockerfile`: `FROM golang:1.24-alpine`
     - `go.mod`: `go 1.24`
     - Workflows de GitHub Actions: `GO_VERSION: '1.24'`

## Archivos Modificados

- `cmd/api/main.go` - Correcciones de código
- `.github/workflows/cloud-run-deploy.yml` - Actualización de versiones
- `.github/workflows/ci.yml` - Actualización de versiones
- `.golangci.yml` - Configuración corregida para v2.1.6
- `Dockerfile` - Actualizado a Go 1.24
- `go.mod` - Actualizado a Go 1.24

## Configuración Final de .golangci.yml

```yaml
run:
  go: '1.24'
  timeout: 5m
  skip-dirs:
    - internal/adapters/gql/generated
  skip-files:
    - ".*_generated\\.go$"
    - ".*\\.pb\\.go$"

linters:
  enable:
    - bodyclose
    - errcheck
    - govet
    - ineffassign
    - noctx
    - staticcheck
    - unused
    - revive
    - unconvert
    - gocyclo
    - gosimple

linters-settings:
  gocyclo:
    min-complexity: 15
  # ... más configuraciones
```

## Verificación Local

Para verificar que todo funciona correctamente antes de hacer push:

```bash
# Instalar golangci-lint v2.1.6 localmente
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6

# Ejecutar el linter
golangci-lint run --timeout=5m

# Verificar versión
golangci-lint --version

# O usar el script de prueba incluido
chmod +x test-linter.sh
./test-linter.sh
```

## Comandos para Commit

```bash
# Añadir todos los cambios
git add .

# Commit con mensaje descriptivo
git commit -m "fix: resolve golangci-lint v2.1.6 configuration issues

- Remove typecheck from linters (not valid in v2.1.6)
- Update .golangci.yml to use correct v2.1.6 syntax
- Use skip-dirs and skip-files instead of nested structures
- Use exclude-rules instead of exclusions
- Keep all previous fixes for exported variables and context keys"

# Push los cambios
git push origin main
```

## Notas Importantes

1. **Compatibilidad con v2.1.6**:
   - `typecheck` no es un linter separado en v2.1.6
   - La sintaxis de configuración es diferente a versiones anteriores
   - Usar `skip-dirs` y `skip-files` en la sección `run`

2. **Best Practices de Go**:
   - Siempre usar tipos personalizados para claves de contexto
   - Todas las variables/funciones/tipos exportados deben tener comentarios
   - Los comentarios deben empezar con el nombre del elemento

3. **Versiones**:
   - golangci-lint v2.1.6 es la última versión disponible
   - Go 1.24 es la última versión estable (lanzada en febrero 2025)
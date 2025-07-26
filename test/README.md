# Estructura de Pruebas

## Estructura de Directorios

Todos los tests están consolidados en la carpeta `test/` siguiendo la misma estructura del código fuente para mantener claridad y facilitar el mantenimiento.

```
test/
  internal/               <- Refleja el paquete internal del código fuente
    domain/               <- Tests para la capa de dominio
      services/           <- Tests para servicios
      models/             <- Tests para modelos
    adapters/             <- Tests para adaptadores
      gql/                <- Tests para GraphQL
        resolvers/        <- Tests de resolvers y control de acceso
      middleware/         <- Tests para middleware
```

## Ejecutando Tests

### Ejecutar todos los tests
```bash
go test ./test/... -v
```

### Ejecutar tests con cobertura
```bash
go test ./test/... -coverprofile=coverage.out -coverpkg=./...
```

### Generar reporte HTML de cobertura
```bash
go tool cover -html=coverage.out -o coverage.html
```

### Script automatizado
Puedes usar el script incluido para ejecutar todos los tests con cobertura:
```bash
./run_consolidated_tests.sh
```

## Estructura de Tests

### Tests Unitarios vs Tests de Integración

- **Tests Unitarios**: Prueban componentes individuales en aislamiento
- **Tests de Integración**: Prueban flujos completos y la interacción entre componentes

### Framework de Testing

Usamos **Ginkgo/Gomega** para tests BDD (Behavior-Driven Development):
- Ginkgo: Framework de testing BDD para Go
- Gomega: Librería de assertions que complementa Ginkgo

### Tests de Control de Acceso

Los tests de control de acceso verifican:
1. **Autenticación**: Usuarios no autenticados no pueden acceder a recursos protegidos
2. **Autorización por Rol**: 
   - ADMIN: Acceso completo a todos los recursos
   - USER: Acceso solo a sus propios recursos
3. **Autorización por Recurso**:
   - Members: Los usuarios solo pueden ver su propio registro
   - Families: Los usuarios solo pueden ver familias donde son miembro origen
   - Payments: Los usuarios solo pueden ver sus propios pagos

## Mantenimiento de Tests

### Al agregar nuevos tests:

1. **Ubicación**: Coloca los tests en el directorio correspondiente dentro de `test/` que refleje la estructura del código
2. **Nomenclatura**: Usa el sufijo `_test` para archivos de test
3. **Suite**: Si es un nuevo paquete, crea un archivo `*_suite_test.go`
4. **Mocks**: Usa los mocks definidos en `test/mocks/`

### Buenas prácticas:

1. **Aislamiento**: Cada test debe ser independiente
2. **Claridad**: Los nombres de tests deben describir claramente qué se está probando
3. **Cobertura**: Apunta a cubrir casos normales y casos límite
4. **Mantenibilidad**: Evita duplicación usando helpers y setup compartido

## Tests Consolidados

Los siguientes tests han sido consolidados desde el código fuente:

- `access_control_test.go`: Tests de control de acceso en resolvers GraphQL
- `authorization_member_test.go`: Tests del middleware de autorización de miembros

Estos tests ahora siguen el estilo BDD con Ginkgo/Gomega para mantener consistencia con el resto de la suite de tests.

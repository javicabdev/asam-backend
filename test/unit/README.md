# Tests Unitarios ASAM Backend

Este directorio contiene todos los tests unitarios del proyecto ASAM Backend. Los tests unitarios se ejecutan sin necesidad de base de datos ni otros servicios externos.

## Estructura

```
test/unit/
├── services/          # Tests de servicios de dominio (con mocks en mocks.go)
└── README.md          # Esta documentación
```

## Convenciones

### 1. Nomenclatura

- Los archivos de test deben terminar en `_test.go`
- Los tests deben seguir el patrón: `Test<Tipo>_<Método>_<Escenario>`
- Ejemplos:
  - `TestMember_Validate_WithValidData`
  - `TestMemberService_CreateMember_DuplicateNumber`

### 2. Estructura de Tests (AAA)

Todos los tests siguen el patrón **Arrange-Act-Assert**:

```go
func TestMember_Validate_WithMissingName(t *testing.T) {
    // Arrange
    member := &models.Member{MembershipNumber: "A00001", Name: ""}
    
    // Act
    err := member.Validate()
    
    // Assert
    if err == nil {
        t.Fatal("expected a validation error for an empty name, got nil")
    }
}
```

### 3. Aserciones

Los tests unitarios usan la librería estándar: `t.Fatal`/`t.Fatalf` cuando el test no puede seguir, y `t.Errorf` para seguir comprobando. Los tests de integración (`test/integration/`) usan [testify](https://github.com/stretchr/testify) (`require`/`assert`). No hay un paquete de aserciones propio: el antiguo `test/unit/testutils` se eliminó porque nadie lo usaba.

### 4. Tests de Tabla (Table-Driven Tests)

Para múltiples casos similares, usa table-driven tests:

```go
func TestMember_Validate_InvalidStates(t *testing.T) {
    tests := []struct {
        name     string
        state    string
        wantErr  bool
    }{
        {"empty state", "", true},
        {"invalid state", "invalid", true},
        {"active state", models.EstadoActivo, false},
        {"inactive state", models.EstadoInactivo, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            member := &models.Member{MembershipNumber: "A00001", Name: "Ana", State: tt.state}
            
            err := member.Validate()
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 5. Mocking

Los mocks de los repositorios están escritos a mano en `services/mocks.go`. Cada uno tiene un campo `<Método>Func` por método: el test asigna solo los que necesita, y los demás devuelven valores vacíos.

```go
func TestCreateFamilyAtomic_Example(t *testing.T) {
    // Arrange
    mockFamilyRepo := &MockFamilyRepository{
        CreateFunc: func(ctx context.Context, family *models.Family) error {
            return nil
        },
    }
    familyService := services.NewFamilyService(
        mockFamilyRepo, &MockMemberRepository{}, &MockPaymentRepository{}, &MockMembershipFeeRepository{},
    )

    // Act
    _, err := familyService.CreateFamilyAtomic(context.Background(), req)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### 6. Punteros

Para campos opcionales (`*time.Time`, `*string`, …) usa `new` con una expresión (Go 1.26+), sin helpers:

```go
fechaNacimiento := new(time.Now().AddDate(-10, 0, 0)) // *time.Time
nombre := new("Ana")                                  // *string
```

### 7. Ejecución de Tests

```bash
# Ejecutar todos los tests unitarios
make test-unit

# Ejecutar tests de un paquete específico
go test ./test/unit/services/...

# Ejecutar con cobertura
go test -cover ./test/unit/...

# Ejecutar un test específico
go test -run TestCreateFamilyAtomic ./test/unit/services/
```

### 8. Mejores Prácticas

1. **No uses base de datos**: Los tests unitarios deben ser rápidos y no depender de servicios externos
2. **Aísla las dependencias**: Usa mocks para todas las dependencias externas
3. **Tests independientes**: Cada test debe poder ejecutarse de forma aislada
4. **Nombres descriptivos**: El nombre del test debe describir claramente qué se está probando
5. **Un assert por test**: Idealmente, cada test debe verificar una sola cosa
6. **Usa t.Helper()**: En funciones helper para que los errores apunten al test correcto
7. **Datos realistas**: Usa valores válidos para el dominio (p. ej. un DNI con letra de control correcta como `12345678Z`, números de socio con el formato `A00123`)

### 9. Cobertura

Objetivos de cobertura:
- Modelos de dominio: 90%+
- Servicios de dominio: 80%+
- Validaciones críticas: 100%

Para verificar la cobertura:
```bash
go test -coverprofile=coverage.out ./test/unit/...
go tool cover -html=coverage.out
```

## Ejemplos

Ver los tests existentes en:
- `test/unit/services/family_service_test.go` - Tests de servicio con mocks
- `test/unit/services/payment_service_fees_test.go` - Tests de servicio con mocks

## Notas

- Los tests unitarios no deben hacer llamadas a BD, APIs externas o sistema de archivos
- Para tests que requieren estos recursos, usar tests de integración en `test/integration/`

## Decisiones de Diseño

### Conversiones de Tipos

Algunas conversiones int->uint generan advertencias G115. Estas están controladas y documentadas cuando:
- El rango de valores está garantizado
- La conversión es segura dentro del contexto de test

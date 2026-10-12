# Tests Unitarios ASAM Backend

Este directorio contiene todos los tests unitarios del proyecto ASAM Backend. Los tests unitarios se ejecutan sin necesidad de base de datos ni otros servicios externos.

## Estructura

```
test/unit/
├── testutils/         # Utilidades compartidas para tests
│   ├── helpers.go     # Funciones helper genéricas
│   └── assertions.go  # Assertions personalizadas
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
    testutils.AssertValidationError(t, err, "name")
}
```

### 3. Assertions Personalizadas

Usa las assertions personalizadas para validaciones comunes:

```go
// En lugar de:
if err := member.Validate(); err != nil {
    t.Fatalf("expected valid member: %v", err)
}

// Usa:
testutils.AssertMemberValid(t, member)
```

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
            
            if tt.wantErr {
                testutils.AssertError(t, err)
            } else {
                testutils.AssertNoError(t, err)
            }
        })
    }
}
```

### 5. Mocking

Para tests de servicios, usa mocks de las interfaces:

```go
func TestMemberService_CreateMember_Success(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockMemberRepository(t)
    mockLogger := mocks.NewMockLogger(t)
    service := services.NewMemberService(mockRepo, mockLogger, mockLogger)
    
    member := &models.Member{MembershipNumber: "A00001", Name: "Ana"}
    
    mockRepo.EXPECT().
        GetByNumeroSocio(mock.Anything, member.MembershipNumber).
        Return(nil, nil)
    
    mockRepo.EXPECT().
        Create(mock.Anything, member).
        Return(nil)
    
    // Act
    err := service.CreateMember(context.Background(), member)
    
    // Assert
    testutils.AssertNoError(t, err)
}
```

### 6. Helpers Disponibles

#### Punteros
```go
strPtr := testutils.StringPtr("value")
timePtr := testutils.TimePtr(time.Now())
uintPtr := testutils.UintPtr(123)
```

#### Tiempos
```go
date := testutils.ParseTime("2024-01-15")
datePtr := testutils.ParseTimePtr("2024-01-15")
```

#### DNI/NIE Válidos
```go
dni := testutils.ValidSpanishDNI()    // "12345678Z"
nie := testutils.ValidSpanishNIE()    // "X1234567L"
invalid := testutils.InvalidSpanishDNI() // "12345678A"
```

### 7. Ejecución de Tests

```bash
# Ejecutar todos los tests unitarios
make test-unit

# Ejecutar tests de un paquete específico
go test ./test/unit/domain/models/...

# Ejecutar con cobertura
go test -cover ./test/unit/...

# Ejecutar un test específico
go test -run TestMember_Validate ./test/unit/domain/models/
```

### 8. Mejores Prácticas

1. **No uses base de datos**: Los tests unitarios deben ser rápidos y no depender de servicios externos
2. **Aísla las dependencias**: Usa mocks para todas las dependencias externas
3. **Tests independientes**: Cada test debe poder ejecutarse de forma aislada
4. **Nombres descriptivos**: El nombre del test debe describir claramente qué se está probando
5. **Un assert por test**: Idealmente, cada test debe verificar una sola cosa
6. **Usa t.Helper()**: En funciones helper para que los errores apunten al test correcto
7. **Datos realistas**: Usa valores válidos para el dominio (p. ej. `testutils.ValidSpanishDNI()`, números de socio con `testutils.GenerateMemberNumber`)

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
- `test/unit/testutils/helpers_test.go` - Tests de los propios helpers

## Notas

- Los tests unitarios no deben hacer llamadas a BD, APIs externas o sistema de archivos
- Para tests que requieren estos recursos, usar tests de integración en `test/integration/`

## Decisiones de Diseño

### Conversiones de Tipos

Algunas conversiones int->uint generan advertencias G115. Estas están controladas y documentadas cuando:
- El rango de valores está garantizado
- La conversión es segura dentro del contexto de test

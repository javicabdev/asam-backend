# Tests Unitarios ASAM Backend

Este directorio contiene todos los tests unitarios del proyecto ASAM Backend. Los tests unitarios se ejecutan sin necesidad de base de datos ni otros servicios externos.

## Estructura

```
test/unit/
├── testutils/         # Utilidades compartidas para tests
│   ├── helpers.go     # Funciones helper genéricas
│   ├── assertions.go  # Assertions personalizadas
│   └── builders.go    # Builders para crear datos de prueba
├── domain/
│   ├── models/       # Tests de modelos de dominio
│   └── services/     # Tests de servicios de dominio
└── README.md         # Esta documentación
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
    builder := testutils.NewTestDataBuilder()
    member := builder.NewMemberBuilder().
        WithName("").
        Build()
    
    // Act
    err := member.Validate()
    
    // Assert
    testutils.AssertValidationError(t, err, "name")
}
```

### 3. Uso de Builders

Utiliza los builders para crear datos de prueba consistentes:

```go
// Crear un miembro válido
member := builder.BuildValidMember()

// Crear con builder fluent
member := builder.NewMemberBuilder().
    WithMembershipNumber("B00001").
    WithEmail("test@example.com").
    AsInactive().
    Build()
```

### 4. Assertions Personalizadas

Usa las assertions personalizadas para validaciones comunes:

```go
// En lugar de:
if err := member.Validate(); err != nil {
    t.Fatalf("expected valid member: %v", err)
}

// Usa:
testutils.AssertMemberValid(t, member)
```

### 5. Tests de Tabla (Table-Driven Tests)

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
            member := builder.NewMemberBuilder().
                WithState(tt.state).
                Build()
            
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

### 6. Mocking

Para tests de servicios, usa mocks de las interfaces:

```go
func TestMemberService_CreateMember_Success(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockMemberRepository(t)
    mockLogger := mocks.NewMockLogger(t)
    service := services.NewMemberService(mockRepo, mockLogger, mockLogger)
    
    member := builder.BuildValidMember()
    
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

### 7. Helpers Disponibles

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

### 8. Ejecución de Tests

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

### 9. Mejores Prácticas

1. **No uses base de datos**: Los tests unitarios deben ser rápidos y no depender de servicios externos
2. **Aísla las dependencias**: Usa mocks para todas las dependencias externas
3. **Tests independientes**: Cada test debe poder ejecutarse de forma aislada
4. **Nombres descriptivos**: El nombre del test debe describir claramente qué se está probando
5. **Un assert por test**: Idealmente, cada test debe verificar una sola cosa
6. **Usa t.Helper()**: En funciones helper para que los errores apunten al test correcto
7. **Datos realistas**: Usa los generadores del seeder para crear datos consistentes

### 10. Cobertura

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
- `test/unit/domain/models/member_test.go` - Tests de modelo
- `test/unit/domain/services/member_service_test.go` - Tests de servicio

## Notas

- Los tests unitarios no deben hacer llamadas a BD, APIs externas o sistema de archivos
- Para tests que requieren estos recursos, usar tests de integración en `test/integration/`
- Los builders reutilizan los generadores del sistema de seeding para consistencia

## Decisiones de Diseño

### Uso de math/rand en Tests

Los builders de test usan intencionalmente `math/rand` en lugar de `crypto/rand` por las siguientes razones:

1. **Reproducibilidad**: Los tests necesitan generar los mismos datos cuando se ejecutan con la misma semilla
2. **Rendimiento**: No necesitamos seguridad criptográfica para datos de prueba
3. **Debugging**: Poder reproducir exactamente los mismos datos facilita la depuración

Esto genera advertencias de gosec (G404) que son suprimidas con `//nolint:gosec` ya que es un uso legítimo en contexto de testing.

### Conversiones de Tipos

Algunas conversiones int->uint generan advertencias G115. Estas están controladas y documentadas cuando:
- El rango de valores está garantizado (ej: `rand.Intn(100) + 1` siempre produce 1-100)
- La conversión es segura dentro del contexto de test

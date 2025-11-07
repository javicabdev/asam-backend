# Testing - Generación de Cuotas Anuales

## Índice

1. [Tests Backend](#tests-backend)
2. [Tests Frontend](#tests-frontend)
3. [Tests de Integración](#tests-de-integración)
4. [Tests Manuales](#tests-manuales)

---

## Tests Backend

### Test Unitarios del Servicio

**Archivo**: `test/unit/services/payment_service_annual_fees_test.go`

```go
package services_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/javicabdev/asam-backend/internal/domain/models"
    "github.com/javicabdev/asam-backend/internal/domain/services"
    "github.com/javicabdev/asam-backend/internal/ports/input"
    "github.com/javicabdev/asam-backend/internal/ports/output"
)

func TestGenerateAnnualFees_Success(t *testing.T) {
    t.Run("Generate fees for active members", func(t *testing.T) {
        ctx := context.Background()

        // Setup mocks
        memberRepo := &MockMemberRepository{
            GetAllActiveFunc: func(ctx context.Context) ([]*models.Member, error) {
                return []*models.Member{
                    {
                        ID:               1,
                        MembershipNumber: "B00001",
                        MembershipType:   models.TipoMembresiaPIndividual,
                        Name:             "Juan",
                        Surnames:         "Pérez",
                        State:            models.EstadoActivo,
                    },
                    {
                        ID:               2,
                        MembershipNumber: "A00001",
                        MembershipType:   models.TipoMembresiaPFamiliar,
                        Name:             "María",
                        Surnames:         "García",
                        State:            models.EstadoActivo,
                    },
                }, nil
            },
        }

        membershipFeeRepo := &MockMembershipFeeRepository{
            FindByYearFunc: func(ctx context.Context, year int) (*models.MembershipFee, error) {
                return nil, nil // No existe aún
            },
            CreateFunc: func(ctx context.Context, fee *models.MembershipFee) error {
                fee.ID = 1
                return nil
            },
        }

        paymentRepo := &MockPaymentRepository{
            FindByMemberFunc: func(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
                return []models.Payment{}, nil // No hay pagos existentes
            },
            CreateFunc: func(ctx context.Context, payment *models.Payment) error {
                payment.ID = uint(time.Now().UnixNano())
                return nil
            },
        }

        service := services.NewPaymentService(
            paymentRepo,
            membershipFeeRepo,
            memberRepo,
            nil, // familyRepo
            nil, // cashFlowRepo
            nil, // feeCalculator
        )

        // Execute
        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, 2024, result.Year)
        assert.Equal(t, 2, result.PaymentsGenerated)
        assert.Equal(t, 0, result.PaymentsExisting)
        assert.Equal(t, 2, result.TotalMembers)
        assert.Len(t, result.Details, 2)

        // Verificar montos calculados correctamente
        assert.Equal(t, 40.0, result.Details[0].Amount)  // Individual
        assert.Equal(t, 50.0, result.Details[1].Amount)  // Familiar (40 + 10)
        assert.True(t, result.Details[0].WasCreated)
        assert.True(t, result.Details[1].WasCreated)
    })
}

func TestGenerateAnnualFees_FutureYear(t *testing.T) {
    t.Run("Reject future year", func(t *testing.T) {
        ctx := context.Background()
        service := services.NewPaymentService(nil, nil, nil, nil, nil, nil)

        req := &input.GenerateAnnualFeesRequest{
            Year:           2030, // Futuro
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "futuro")
    })
}

func TestGenerateAnnualFees_InvalidAmount(t *testing.T) {
    t.Run("Reject zero amount", func(t *testing.T) {
        ctx := context.Background()
        service := services.NewPaymentService(nil, nil, nil, nil, nil, nil)

        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  0.0, // Inválido
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "positivo")
    })

    t.Run("Reject negative family extra", func(t *testing.T) {
        ctx := context.Background()
        service := services.NewPaymentService(nil, nil, nil, nil, nil, nil)

        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: -5.0, // Inválido
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "negativo")
    })
}

func TestGenerateAnnualFees_Idempotent(t *testing.T) {
    t.Run("Do not create duplicates", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetAllActiveFunc: func(ctx context.Context) ([]*models.Member, error) {
                return []*models.Member{
                    {
                        ID:               1,
                        MembershipNumber: "B00001",
                        MembershipType:   models.TipoMembresiaPIndividual,
                        Name:             "Juan",
                        Surnames:         "Pérez",
                        State:            models.EstadoActivo,
                    },
                }, nil
            },
        }

        membershipFeeRepo := &MockMembershipFeeRepository{
            FindByYearFunc: func(ctx context.Context, year int) (*models.MembershipFee, error) {
                // Ya existe la cuota
                return &models.MembershipFee{
                    ID:             1,
                    Year:           2024,
                    BaseFeeAmount:  40.0,
                    FamilyFeeExtra: 10.0,
                }, nil
            },
        }

        paymentRepo := &MockPaymentRepository{
            FindByMemberFunc: func(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
                // Ya existe un pago para esta cuota
                feeID := uint(1)
                return []models.Payment{
                    {
                        ID:              100,
                        MemberID:        memberID,
                        MembershipFeeID: &feeID,
                        Amount:          40.0,
                        Status:          models.PaymentStatusPending,
                    },
                }, nil
            },
            CreateFunc: func(ctx context.Context, payment *models.Payment) error {
                t.Error("No debería crear pagos duplicados")
                return nil
            },
        }

        service := services.NewPaymentService(
            paymentRepo,
            membershipFeeRepo,
            memberRepo,
            nil,
            nil,
            nil,
        )

        // Execute
        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, 0, result.PaymentsGenerated) // No se crearon nuevos
        assert.Equal(t, 1, result.PaymentsExisting)  // Uno ya existía
        assert.False(t, result.Details[0].WasCreated)
    })
}

func TestGenerateAnnualFees_NoActiveMembers(t *testing.T) {
    t.Run("Handle no active members", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetAllActiveFunc: func(ctx context.Context) ([]*models.Member, error) {
                return []*models.Member{}, nil // Sin socios activos
            },
        }

        membershipFeeRepo := &MockMembershipFeeRepository{
            FindByYearFunc: func(ctx context.Context, year int) (*models.MembershipFee, error) {
                return nil, nil
            },
            CreateFunc: func(ctx context.Context, fee *models.MembershipFee) error {
                fee.ID = 1
                return nil
            },
        }

        service := services.NewPaymentService(
            nil,
            membershipFeeRepo,
            memberRepo,
            nil,
            nil,
            nil,
        )

        // Execute
        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, 0, result.PaymentsGenerated)
        assert.Equal(t, 0, result.TotalMembers)
        assert.Empty(t, result.Details)
    })
}

func TestGenerateAnnualFees_UpdateExistingFee(t *testing.T) {
    t.Run("Update fee amount if different", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetAllActiveFunc: func(ctx context.Context) ([]*models.Member, error) {
                return []*models.Member{
                    {
                        ID:               1,
                        MembershipNumber: "B00001",
                        MembershipType:   models.TipoMembresiaPIndividual,
                        State:            models.EstadoActivo,
                    },
                }, nil
            },
        }

        updated := false
        membershipFeeRepo := &MockMembershipFeeRepository{
            FindByYearFunc: func(ctx context.Context, year int) (*models.MembershipFee, error) {
                // Ya existe con monto diferente
                return &models.MembershipFee{
                    ID:             1,
                    Year:           2024,
                    BaseFeeAmount:  35.0, // Monto anterior
                    FamilyFeeExtra: 8.0,
                }, nil
            },
            UpdateFunc: func(ctx context.Context, fee *models.MembershipFee) error {
                updated = true
                assert.Equal(t, 40.0, fee.BaseFeeAmount)
                assert.Equal(t, 10.0, fee.FamilyFeeExtra)
                return nil
            },
        }

        paymentRepo := &MockPaymentRepository{
            FindByMemberFunc: func(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
                return []models.Payment{}, nil
            },
            CreateFunc: func(ctx context.Context, payment *models.Payment) error {
                return nil
            },
        }

        service := services.NewPaymentService(
            paymentRepo,
            membershipFeeRepo,
            memberRepo,
            nil,
            nil,
            nil,
        )

        // Execute
        req := &input.GenerateAnnualFeesRequest{
            Year:           2024,
            BaseFeeAmount:  40.0, // Nuevo monto
            FamilyFeeExtra: 10.0,
        }

        result, err := service.GenerateAnnualFees(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.True(t, updated, "Debería actualizar la cuota existente")
    })
}
```

### Test del Resolver GraphQL

**Archivo**: `test/unit/resolvers/payment_resolver_test.go`

```go
func TestGenerateAnnualFeesResolver(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        // Setup
        mockService := &MockPaymentService{
            GenerateAnnualFeesFunc: func(ctx context.Context, req *input.GenerateAnnualFeesRequest) (*input.GenerateAnnualFeesResponse, error) {
                return &input.GenerateAnnualFeesResponse{
                    Year:              req.Year,
                    MembershipFeeID:   1,
                    PaymentsGenerated: 45,
                    PaymentsExisting:  5,
                    TotalMembers:      50,
                    Details:           []input.PaymentGenDetail{},
                }, nil
            },
        }

        resolver := &mutationResolver{
            PaymentService: mockService,
        }

        // Execute
        ctx := context.WithValue(context.Background(), "user_role", "admin")
        input := model.GenerateAnnualFeesInput{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := resolver.GenerateAnnualFees(ctx, input)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, 2024, result.Year)
        assert.Equal(t, 45, result.PaymentsGenerated)
        assert.Equal(t, 5, result.PaymentsExisting)
    })

    t.Run("Unauthorized", func(t *testing.T) {
        resolver := &mutationResolver{}

        // Execute con usuario no admin
        ctx := context.WithValue(context.Background(), "user_role", "user")
        input := model.GenerateAnnualFeesInput{
            Year:           2024,
            BaseFeeAmount:  40.0,
            FamilyFeeExtra: 10.0,
        }

        result, err := resolver.GenerateAnnualFees(ctx, input)

        // Assert
        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "unauthorized")
    })
}
```

---

## Tests Frontend

### Test del Hook

**Archivo**: `src/features/payments/hooks/useGenerateAnnualFees.test.ts`

```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { useGenerateAnnualFees } from './useGenerateAnnualFees'
import { GENERATE_ANNUAL_FEES } from '../api/mutations'

describe('useGenerateAnnualFees', () => {
  it('should generate fees successfully', async () => {
    const mocks = [
      {
        request: {
          query: GENERATE_ANNUAL_FEES,
          variables: {
            input: {
              year: 2024,
              baseFeeAmount: 40,
              familyFeeExtra: 10,
            },
          },
        },
        result: {
          data: {
            generateAnnualFees: {
              year: 2024,
              membershipFeeId: '1',
              paymentsGenerated: 45,
              paymentsExisting: 5,
              totalMembers: 50,
              details: [],
            },
          },
        },
      },
    ]

    const wrapper = ({ children }) => (
      <MockedProvider mocks={mocks}>{children}</MockedProvider>
    )

    const { result } = renderHook(() => useGenerateAnnualFees(), { wrapper })

    await waitFor(() => {
      result.current.generateFees({
        year: 2024,
        baseFeeAmount: 40,
        familyFeeExtra: 10,
      })
    })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBeUndefined()
  })

  it('should handle errors', async () => {
    const mocks = [
      {
        request: {
          query: GENERATE_ANNUAL_FEES,
          variables: {
            input: {
              year: 2030, // Futuro
              baseFeeAmount: 40,
              familyFeeExtra: 10,
            },
          },
        },
        error: new Error('No se pueden generar cuotas para años futuros'),
      },
    ]

    const wrapper = ({ children }) => (
      <MockedProvider mocks={mocks}>{children}</MockedProvider>
    )

    const { result } = renderHook(() => useGenerateAnnualFees(), { wrapper })

    await expect(
      result.current.generateFees({
        year: 2030,
        baseFeeAmount: 40,
        familyFeeExtra: 10,
      })
    ).rejects.toThrow()
  })
})
```

### Test del Componente

**Archivo**: `src/features/payments/components/GenerateFeesDialog.test.tsx`

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MockedProvider } from '@apollo/client/testing'
import { GenerateFeesDialog } from './GenerateFeesDialog'
import { GENERATE_ANNUAL_FEES } from '../api/mutations'

describe('GenerateFeesDialog', () => {
  it('should render form correctly', () => {
    render(
      <MockedProvider>
        <GenerateFeesDialog open={true} onClose={() => {}} />
      </MockedProvider>
    )

    expect(screen.getByLabelText(/año/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/monto base/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/extra familiar/i)).toBeInTheDocument()
  })

  it('should validate future year', () => {
    render(
      <MockedProvider>
        <GenerateFeesDialog open={true} onClose={() => {}} />
      </MockedProvider>
    )

    const yearInput = screen.getByLabelText(/año/i)
    fireEvent.change(yearInput, { target: { value: '2030' } })

    expect(screen.getByText(/no puede ser año futuro/i)).toBeInTheDocument()
    expect(screen.getByText(/generar/i)).toBeDisabled()
  })

  it('should validate amount', () => {
    render(
      <MockedProvider>
        <GenerateFeesDialog open={true} onClose={() => {}} />
      </MockedProvider>
    )

    const amountInput = screen.getByLabelText(/monto base/i)
    fireEvent.change(amountInput, { target: { value: '0' } })

    expect(screen.getByText(/debe ser mayor a 0/i)).toBeInTheDocument()
    expect(screen.getByText(/generar/i)).toBeDisabled()
  })

  it('should submit successfully', async () => {
    const mocks = [
      {
        request: {
          query: GENERATE_ANNUAL_FEES,
          variables: {
            input: {
              year: 2024,
              baseFeeAmount: 40,
              familyFeeExtra: 10,
            },
          },
        },
        result: {
          data: {
            generateAnnualFees: {
              year: 2024,
              membershipFeeId: '1',
              paymentsGenerated: 45,
              paymentsExisting: 5,
              totalMembers: 50,
              details: [],
            },
          },
        },
      },
    ]

    const onClose = jest.fn()
    const onSuccess = jest.fn()

    render(
      <MockedProvider mocks={mocks}>
        <GenerateFeesDialog
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
        />
      </MockedProvider>
    )

    const submitButton = screen.getByText(/generar/i)
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
    })
  })
})
```

---

## Tests de Integración

### Test End-to-End con Base de Datos

**Archivo**: `test/integration/generate_annual_fees_test.go`

```go
func TestGenerateAnnualFees_Integration(t *testing.T) {
    // Setup real database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create test members
    member1 := createTestMember(t, db, "B00001", models.TipoMembresiaPIndividual)
    member2 := createTestMember(t, db, "A00001", models.TipoMembresiaPFamiliar)

    // Setup service with real repositories
    memberRepo := db.NewMemberRepository(db)
    membershipFeeRepo := db.NewMembershipFeeRepository(db)
    paymentRepo := db.NewPaymentRepository(db)

    service := services.NewPaymentService(
        paymentRepo,
        membershipFeeRepo,
        memberRepo,
        nil,
        nil,
        nil,
    )

    // Execute
    req := &input.GenerateAnnualFeesRequest{
        Year:           2024,
        BaseFeeAmount:  40.0,
        FamilyFeeExtra: 10.0,
    }

    result, err := service.GenerateAnnualFees(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 2, result.PaymentsGenerated)
    assert.Equal(t, 0, result.PaymentsExisting)

    // Verify payments in DB
    payments1 := getPaymentsByMember(t, db, member1.ID)
    assert.Len(t, payments1, 1)
    assert.Equal(t, 40.0, payments1[0].Amount)
    assert.Equal(t, models.PaymentStatusPending, payments1[0].Status)

    payments2 := getPaymentsByMember(t, db, member2.ID)
    assert.Len(t, payments2, 1)
    assert.Equal(t, 50.0, payments2[0].Amount) // 40 + 10

    // Verify membership fee in DB
    fee := getMembershipFeeByYear(t, db, 2024)
    assert.NotNil(t, fee)
    assert.Equal(t, 40.0, fee.BaseFeeAmount)
    assert.Equal(t, 10.0, fee.FamilyFeeExtra)

    // Test idempotence: Run again
    result2, err := service.GenerateAnnualFees(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, 0, result2.PaymentsGenerated) // No new payments
    assert.Equal(t, 2, result2.PaymentsExisting)  // 2 already existed

    // Verify no duplicates
    payments1After := getPaymentsByMember(t, db, member1.ID)
    assert.Len(t, payments1After, 1) // Still only 1
}
```

---

## Tests Manuales

### Checklist de Tests Manuales

#### Preparación
- [ ] Backend levantado y corriendo
- [ ] Frontend levantado y corriendo
- [ ] Base de datos limpia o con datos de prueba
- [ ] Usuario admin creado y con sesión activa

#### Tests Funcionales

**Test 1: Generación Básica**
- [ ] Navegar a Pagos
- [ ] Click en "Generar Cuotas Anuales"
- [ ] Año: 2024, Monto Base: 40€, Extra Familiar: 10€
- [ ] Click "Generar"
- [ ] ✅ Verificar mensaje de éxito
- [ ] ✅ Verificar que muestra N pagos generados
- [ ] ✅ Verificar que la lista de pagos se actualiza

**Test 2: Validación de Año Futuro**
- [ ] Abrir diálogo
- [ ] Ingresar año 2030
- [ ] ✅ Campo muestra error
- [ ] ✅ Botón "Generar" está deshabilitado

**Test 3: Validación de Montos**
- [ ] Monto base = 0
- [ ] ✅ Botón deshabilitado
- [ ] Extra familiar = -5
- [ ] ✅ Muestra error

**Test 4: Idempotencia**
- [ ] Generar cuotas de 2024 (primera vez)
- [ ] Resultado: "X nuevos, 0 existentes"
- [ ] Generar cuotas de 2024 (segunda vez)
- [ ] ✅ Resultado: "0 nuevos, X existentes"
- [ ] ✅ No hay pagos duplicados en la BD

**Test 5: Múltiples Años (Migración)**
- [ ] Generar 2020 (monto 35€)
- [ ] Generar 2021 (monto 35€)
- [ ] Generar 2022 (monto 38€)
- [ ] Generar 2023 (monto 40€)
- [ ] Generar 2024 (monto 40€)
- [ ] ✅ Cada socio tiene 5 cuotas pendientes
- [ ] ✅ Los montos son correctos

**Test 6: Cálculo de Montos**
- [ ] Generar con Base: 40€, Extra: 10€
- [ ] Verificar en BD:
- [ ] ✅ Socios individuales: 40€
- [ ] ✅ Socios familiares: 50€

**Test 7: Solo Socios Activos**
- [ ] Tener 2 socios activos y 1 inactivo
- [ ] Generar cuotas
- [ ] ✅ Solo se crean 2 pagos (socios activos)
- [ ] ✅ No se crea pago para socio inactivo

**Test 8: Sin Socios Activos**
- [ ] Dar de baja todos los socios
- [ ] Intentar generar cuotas
- [ ] ✅ Resultado: "0 pagos generados"
- [ ] ✅ No hay error, mensaje informativo

#### Tests de Performance

**Test 9: Generación con Muchos Socios**
- [ ] Crear 100+ socios activos
- [ ] Generar cuotas
- [ ] ✅ Tiempo < 10 segundos
- [ ] ✅ No hay timeout
- [ ] ✅ Todos los pagos se crean correctamente

**Test 10: Concurrencia**
- [ ] Usuario A inicia generación 2024
- [ ] Usuario B inicia generación 2024 simultáneamente
- [ ] ✅ Ambos terminan sin error
- [ ] ✅ No hay pagos duplicados
- [ ] ✅ Idempotencia se mantiene

#### Tests de UI/UX

**Test 11: Loading States**
- [ ] Click "Generar"
- [ ] ✅ Botón muestra spinner
- [ ] ✅ Botón se deshabilita
- [ ] ✅ Mensaje "Generando..." visible

**Test 12: Error Handling**
- [ ] Simular error de red (backend apagado)
- [ ] ✅ Muestra mensaje de error claro
- [ ] ✅ No se congela la UI
- [ ] ✅ Usuario puede reintentar

**Test 13: Responsividad**
- [ ] Abrir diálogo en móvil
- [ ] ✅ Formulario es usable
- [ ] ✅ Botones son clickables
- [ ] ✅ Textos legibles

---

## Reporte de Bugs

### Template para Reportar Bugs

```markdown
## Descripción
[Descripción clara del bug]

## Pasos para Reproducir
1. Navegar a...
2. Click en...
3. Ingresar...
4. Ver error...

## Comportamiento Esperado
[Qué debería pasar]

## Comportamiento Actual
[Qué está pasando]

## Screenshots
[Si aplica]

## Entorno
- Browser: [Chrome/Firefox/etc]
- OS: [Mac/Windows/Linux]
- Backend version: [commit hash]
- Frontend version: [commit hash]

## Logs
[Logs relevantes del backend o consola del browser]
```

---

## Métricas de Calidad

### Cobertura de Tests

**Objetivo**: > 80% de cobertura

```bash
# Backend
go test -cover ./internal/domain/services/...

# Frontend
npm run test:coverage
```

### Tests que Deben Pasar

- [ ] Todos los tests unitarios del servicio (8+ tests)
- [ ] Todos los tests del resolver (2+ tests)
- [ ] Todos los tests del hook (2+ tests)
- [ ] Todos los tests del componente (4+ tests)
- [ ] Test de integración E2E (1 test)
- [ ] Todos los tests manuales (13 tests)

---

**Próximo paso**: Leer [Despliegue](./deployment.md)

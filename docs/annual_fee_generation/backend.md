# Backend - Generación de Cuotas Anuales

## Índice

1. [Arquitectura](#arquitectura)
2. [Implementación Paso a Paso](#implementación-paso-a-paso)
3. [Código Completo](#código-completo)
4. [Testing](#testing)

---

## Arquitectura

### Componentes a Modificar/Crear

```
internal/
├── domain/
│   └── services/
│       └── payment_service.go          [MODIFICAR]
├── ports/
│   ├── input/
│   │   └── payment_service.go          [MODIFICAR]
│   └── output/
│       ├── member_repository.go        [VERIFICAR]
│       └── payment_repository.go       [VERIFICAR]
└── adapters/
    └── gql/
        ├── schema/
        │   └── payment.graphqls          [MODIFICAR]
        └── resolvers/
            └── schema.resolvers.go       [MODIFICAR]
```

---

## Implementación Paso a Paso

### PASO 1: Añadir Método en el Repositorio (si no existe)

**Archivo**: `internal/ports/output/member_repository.go`

**Verificar que existe** el método:

```go
// GetAllActive obtiene todos los miembros activos
GetAllActive(ctx context.Context) ([]*models.Member, error)
```

**Si NO existe, añadir en** `internal/adapters/db/member_repository.go`:

```go
func (r *memberRepository) GetAllActive(ctx context.Context) ([]*models.Member, error) {
    var members []*models.Member

    result := r.db.WithContext(ctx).
        Where("state = ?", models.EstadoActivo).
        Order("membership_number ASC").
        Find(&members)

    if result.Error != nil {
        return nil, appErrors.DB(result.Error, "error getting active members")
    }

    return members, nil
}
```

**Añadir firma en la interfaz** `internal/ports/output/member_repository.go`:

```go
type MemberRepository interface {
    // ... métodos existentes ...

    // GetAllActive obtiene todos los miembros activos
    GetAllActive(ctx context.Context) ([]*models.Member, error)
}
```

---

### PASO 2: Añadir DTOs de Request/Response

**Archivo**: `internal/ports/input/payment_service.go`

**Añadir al final del archivo**:

```go
// GenerateAnnualFeesRequest contiene los datos para generar cuotas anuales
type GenerateAnnualFeesRequest struct {
    Year              int     // Año para el cual generar cuotas
    BaseFeeAmount     float64 // Monto base (para socios individuales)
    FamilyFeeExtra    float64 // Monto adicional para socios familiares
}

// GenerateAnnualFeesResponse contiene el resultado de la generación
type GenerateAnnualFeesResponse struct {
    Year              int                // Año procesado
    MembershipFeeID   uint               // ID de la cuota anual creada/usada
    PaymentsGenerated int                // Número de pagos creados
    PaymentsExisting  int                // Número de pagos que ya existían
    TotalMembers      int                // Total de socios activos procesados
    Details           []PaymentGenDetail // Detalle por socio (opcional)
}

// PaymentGenDetail contiene el detalle de generación por socio
type PaymentGenDetail struct {
    MemberID     uint
    MemberNumber string
    MemberName   string
    Amount       float64
    WasCreated   bool // true si se creó, false si ya existía
    Error        string // Si hubo algún error específico
}
```

**Añadir método en la interfaz** `PaymentService`:

```go
type PaymentService interface {
    // ... métodos existentes ...

    // GenerateAnnualFees genera pagos pendientes para todos los socios activos de un año
    GenerateAnnualFees(ctx context.Context, req *GenerateAnnualFeesRequest) (*GenerateAnnualFeesResponse, error)
}
```

---

### PASO 3: Implementar Servicio de Generación

**Archivo**: `internal/domain/services/payment_service.go`

**Añadir al final del archivo, antes del último comentario**:

```go
// GenerateAnnualFees genera cuotas anuales para todos los socios activos
func (s *paymentService) GenerateAnnualFees(ctx context.Context, req *input.GenerateAnnualFeesRequest) (*input.GenerateAnnualFeesResponse, error) {
    // 1. Validar año (no puede ser futuro)
    currentYear := time.Now().Year()
    if req.Year > currentYear {
        return nil, errors.NewValidationError(
            "No se pueden generar cuotas para años futuros",
            map[string]string{"year": "Debe ser el año actual o anterior"},
        )
    }

    // Validar montos
    if req.BaseFeeAmount <= 0 {
        return nil, errors.NewValidationError(
            "El monto base debe ser positivo",
            map[string]string{"baseFeeAmount": "Debe ser mayor a 0"},
        )
    }

    if req.FamilyFeeExtra < 0 {
        return nil, errors.NewValidationError(
            "El monto extra familiar no puede ser negativo",
            map[string]string{"familyFeeExtra": "Debe ser 0 o mayor"},
        )
    }

    // 2. Buscar o crear MembershipFee para el año
    membershipFee, err := s.ensureMembershipFee(ctx, req.Year, req.BaseFeeAmount, req.FamilyFeeExtra)
    if err != nil {
        return nil, err
    }

    // 3. Obtener todos los socios activos
    activeMembers, err := s.memberRepo.GetAllActive(ctx)
    if err != nil {
        return nil, errors.DB(err, "error obteniendo socios activos")
    }

    if len(activeMembers) == 0 {
        return &input.GenerateAnnualFeesResponse{
            Year:              req.Year,
            MembershipFeeID:   membershipFee.ID,
            PaymentsGenerated: 0,
            PaymentsExisting:  0,
            TotalMembers:      0,
            Details:           []input.PaymentGenDetail{},
        }, nil
    }

    // 4. Generar pagos para cada socio (con idempotencia)
    response := &input.GenerateAnnualFeesResponse{
        Year:            req.Year,
        MembershipFeeID: membershipFee.ID,
        TotalMembers:    len(activeMembers),
        Details:         make([]input.PaymentGenDetail, 0, len(activeMembers)),
    }

    for _, member := range activeMembers {
        detail := s.generatePaymentForMember(ctx, member, membershipFee)
        response.Details = append(response.Details, detail)

        if detail.WasCreated {
            response.PaymentsGenerated++
        } else {
            response.PaymentsExisting++
        }
    }

    return response, nil
}

// ensureMembershipFee busca o crea la cuota anual para un año
func (s *paymentService) ensureMembershipFee(ctx context.Context, year int, baseAmount, familyExtra float64) (*models.MembershipFee, error) {
    // Buscar cuota existente
    fee, err := s.membershipFeeRepo.FindByYear(ctx, year)
    if err != nil {
        return nil, errors.DB(err, "error buscando cuota anual")
    }

    // Si existe, verificar que los montos coincidan
    if fee != nil {
        // Permitir actualizar montos si son diferentes
        needsUpdate := false
        if fee.BaseFeeAmount != baseAmount {
            fee.BaseFeeAmount = baseAmount
            needsUpdate = true
        }
        if fee.FamilyFeeExtra != familyExtra {
            fee.FamilyFeeExtra = familyExtra
            needsUpdate = true
        }

        if needsUpdate {
            if err := s.membershipFeeRepo.Update(ctx, fee); err != nil {
                return nil, errors.DB(err, "error actualizando cuota anual")
            }
        }

        return fee, nil
    }

    // Crear nueva cuota anual
    fee = &models.MembershipFee{
        Year:           year,
        BaseFeeAmount:  baseAmount,
        FamilyFeeExtra: familyExtra,
        DueDate:        time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC),
    }

    if err := s.membershipFeeRepo.Create(ctx, fee); err != nil {
        return nil, errors.DB(err, "error creando cuota anual")
    }

    return fee, nil
}

// generatePaymentForMember genera un pago para un socio específico (con idempotencia)
func (s *paymentService) generatePaymentForMember(ctx context.Context, member *models.Member, fee *models.MembershipFee) input.PaymentGenDetail {
    detail := input.PaymentGenDetail{
        MemberID:     member.ID,
        MemberNumber: member.MembershipNumber,
        MemberName:   member.NombreCompleto(),
        WasCreated:   false,
    }

    // Calcular monto según tipo de membresía
    amount := fee.Calculate(member.MembershipType == models.TipoMembresiaPFamiliar)
    detail.Amount = amount

    // Verificar si ya existe un pago para este socio y año
    existingPayments, err := s.paymentRepo.FindByMember(ctx, member.ID, time.Time{}, time.Now().AddDate(10, 0, 0))
    if err != nil {
        detail.Error = fmt.Sprintf("Error verificando pagos existentes: %v", err)
        return detail
    }

    // Buscar si ya existe pago para esta cuota
    for _, p := range existingPayments {
        if p.MembershipFeeID != nil && *p.MembershipFeeID == fee.ID {
            // Ya existe un pago para esta cuota
            detail.WasCreated = false
            return detail
        }
    }

    // Crear el pago pendiente
    payment := &models.Payment{
        MemberID:        member.ID,
        MembershipFeeID: &fee.ID,
        Amount:          amount,
        Status:          models.PaymentStatusPending,
        PaymentMethod:   "", // Se llenará cuando se confirme
        Notes:           fmt.Sprintf("Cuota anual %d generada automáticamente", fee.Year),
    }

    if err := s.paymentRepo.Create(ctx, payment); err != nil {
        detail.Error = fmt.Sprintf("Error creando pago: %v", err)
        return detail
    }

    detail.WasCreated = true
    return detail
}
```

---

### PASO 4: Actualizar Schema GraphQL

**Archivo**: `internal/adapters/gql/schema/payment.graphqls`

**Añadir al final del archivo**:

```graphql
# ============================================
# Inputs para generación de cuotas anuales
# ============================================

"""
Input para generar cuotas anuales de todos los socios activos
"""
input GenerateAnnualFeesInput {
  """
  Año para el cual generar las cuotas (no puede ser futuro)
  """
  year: Int!

  """
  Monto base de la cuota (para socios individuales)
  """
  baseFeeAmount: Float!

  """
  Monto adicional para socios familiares (0 si es el mismo)
  """
  familyFeeExtra: Float!
}

# ============================================
# Types para respuesta de generación
# ============================================

"""
Detalle de generación de pago por socio
"""
type PaymentGenerationDetail {
  memberId: ID!
  memberNumber: String!
  memberName: String!
  amount: Float!
  wasCreated: Boolean!
  error: String
}

"""
Respuesta de la generación masiva de cuotas anuales
"""
type GenerateAnnualFeesResponse {
  """
  Año para el cual se generaron las cuotas
  """
  year: Int!

  """
  ID de la cuota anual (MembershipFee) utilizada
  """
  membershipFeeId: ID!

  """
  Número de pagos nuevos generados
  """
  paymentsGenerated: Int!

  """
  Número de pagos que ya existían (idempotencia)
  """
  paymentsExisting: Int!

  """
  Total de socios activos procesados
  """
  totalMembers: Int!

  """
  Detalle de generación por socio (opcional, puede ser grande)
  """
  details: [PaymentGenerationDetail!]
}

# ============================================
# Mutation
# ============================================

extend type Mutation {
  """
  Genera cuotas anuales para todos los socios activos de un año específico.
  Solo accesible para administradores.
  Operación idempotente: si ya existen pagos para el año, no se crean duplicados.
  """
  generateAnnualFees(input: GenerateAnnualFeesInput!): GenerateAnnualFeesResponse!
}
```

---

### PASO 5: Implementar Resolver GraphQL

**Archivo**: `internal/adapters/gql/resolvers/schema.resolvers.go`

**Buscar la sección de mutations de Payment y añadir**:

```go
// GenerateAnnualFees is the resolver for the generateAnnualFees field.
func (r *mutationResolver) GenerateAnnualFees(ctx context.Context, input model.GenerateAnnualFeesInput) (*model.GenerateAnnualFeesResponse, error) {
    // Solo ADMIN puede generar cuotas
    if err := middleware.MustBeAdmin(ctx); err != nil {
        return nil, err
    }

    // Construir request
    req := &input.GenerateAnnualFeesRequest{
        Year:           input.Year,
        BaseFeeAmount:  input.BaseFeeAmount,
        FamilyFeeExtra: input.FamilyFeeExtra,
    }

    // Llamar al servicio
    result, err := r.PaymentService.GenerateAnnualFees(ctx, req)
    if err != nil {
        return nil, err
    }

    // Mapear respuesta
    response := &model.GenerateAnnualFeesResponse{
        Year:              result.Year,
        MembershipFeeID:   strconv.FormatUint(uint64(result.MembershipFeeID), 10),
        PaymentsGenerated: result.PaymentsGenerated,
        PaymentsExisting:  result.PaymentsExisting,
        TotalMembers:      result.TotalMembers,
        Details:           make([]*model.PaymentGenerationDetail, 0, len(result.Details)),
    }

    // Mapear detalles (opcional: solo si hay pocos socios)
    // Para producción con muchos socios, considerar omitir detalles
    if len(result.Details) <= 100 { // Límite razonable
        for _, detail := range result.Details {
            responseDetail := &model.PaymentGenerationDetail{
                MemberID:     strconv.FormatUint(uint64(detail.MemberID), 10),
                MemberNumber: detail.MemberNumber,
                MemberName:   detail.MemberName,
                Amount:       detail.Amount,
                WasCreated:   detail.WasCreated,
            }
            if detail.Error != "" {
                responseDetail.Error = &detail.Error
            }
            response.Details = append(response.Details, responseDetail)
        }
    }

    return response, nil
}
```

**Nota**: Asegúrate de importar `"strconv"` al inicio del archivo.

---

### PASO 6: Regenerar Código GraphQL

**Ejecutar desde el directorio raíz del backend**:

```bash
cd /Users/javierfernandezcabanas/repos/asam-backend

# Regenerar código GraphQL
go run github.com/99designs/gqlgen generate

# Verificar que no hay errores de compilación
go build ./...
```

---

## Código Completo de Referencia

### payment_service.go - Método Completo

```go
// GenerateAnnualFees genera cuotas anuales para todos los socios activos
func (s *paymentService) GenerateAnnualFees(ctx context.Context, req *input.GenerateAnnualFeesRequest) (*input.GenerateAnnualFeesResponse, error) {
    // 1. Validar año (no puede ser futuro)
    currentYear := time.Now().Year()
    if req.Year > currentYear {
        return nil, errors.NewValidationError(
            "No se pueden generar cuotas para años futuros",
            map[string]string{"year": "Debe ser el año actual o anterior"},
        )
    }

    // Validar montos
    if req.BaseFeeAmount <= 0 {
        return nil, errors.NewValidationError(
            "El monto base debe ser positivo",
            map[string]string{"baseFeeAmount": "Debe ser mayor a 0"},
        )
    }

    if req.FamilyFeeExtra < 0 {
        return nil, errors.NewValidationError(
            "El monto extra familiar no puede ser negativo",
            map[string]string{"familyFeeExtra": "Debe ser 0 o mayor"},
        )
    }

    // 2. Buscar o crear MembershipFee para el año
    membershipFee, err := s.ensureMembershipFee(ctx, req.Year, req.BaseFeeAmount, req.FamilyFeeExtra)
    if err != nil {
        return nil, err
    }

    // 3. Obtener todos los socios activos
    activeMembers, err := s.memberRepo.GetAllActive(ctx)
    if err != nil {
        return nil, errors.DB(err, "error obteniendo socios activos")
    }

    if len(activeMembers) == 0 {
        return &input.GenerateAnnualFeesResponse{
            Year:              req.Year,
            MembershipFeeID:   membershipFee.ID,
            PaymentsGenerated: 0,
            PaymentsExisting:  0,
            TotalMembers:      0,
            Details:           []input.PaymentGenDetail{},
        }, nil
    }

    // 4. Generar pagos para cada socio (con idempotencia)
    response := &input.GenerateAnnualFeesResponse{
        Year:            req.Year,
        MembershipFeeID: membershipFee.ID,
        TotalMembers:    len(activeMembers),
        Details:         make([]input.PaymentGenDetail, 0, len(activeMembers)),
    }

    for _, member := range activeMembers {
        detail := s.generatePaymentForMember(ctx, member, membershipFee)
        response.Details = append(response.Details, detail)

        if detail.WasCreated {
            response.PaymentsGenerated++
        } else {
            response.PaymentsExisting++
        }
    }

    return response, nil
}
```

---

## Testing

### Test Unitario

**Archivo**: `test/unit/services/payment_service_test.go`

```go
func TestGenerateAnnualFees_Success(t *testing.T) {
    ctx := context.Background()

    // Setup mocks
    memberRepo := &MockMemberRepository{
        GetAllActiveFunc: func(ctx context.Context) ([]*models.Member, error) {
            return []*models.Member{
                {ID: 1, MembershipNumber: "B00001", MembershipType: models.TipoMembresiaPIndividual, Name: "Juan", Surnames: "Pérez"},
                {ID: 2, MembershipNumber: "A00001", MembershipType: models.TipoMembresiaPFamiliar, Name: "María", Surnames: "García"},
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
            payment.ID = uint(rand.Intn(1000))
            return nil
        },
    }

    service := services.NewPaymentService(paymentRepo, membershipFeeRepo, memberRepo, nil, nil, nil)

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

    // Verificar montos calculados correctamente
    assert.Equal(t, 40.0, result.Details[0].Amount) // Individual
    assert.Equal(t, 50.0, result.Details[1].Amount) // Familiar (40 + 10)
}

func TestGenerateAnnualFees_FutureYear(t *testing.T) {
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
}

func TestGenerateAnnualFees_Idempotent(t *testing.T) {
    ctx := context.Background()

    // Mock: Pagos ya existen
    paymentRepo := &MockPaymentRepository{
        FindByMemberFunc: func(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
            // Simular que ya existe un pago para esta cuota
            return []models.Payment{
                {
                    ID:              1,
                    MemberID:        memberID,
                    MembershipFeeID: pointerUint(1),
                    Status:          models.PaymentStatusPending,
                },
            }, nil
        },
    }

    // ... resto de mocks ...

    service := services.NewPaymentService(paymentRepo, membershipFeeRepo, memberRepo, nil, nil, nil)

    req := &input.GenerateAnnualFeesRequest{
        Year:           2024,
        BaseFeeAmount:  40.0,
        FamilyFeeExtra: 10.0,
    }

    result, err := service.GenerateAnnualFees(ctx, req)

    assert.NoError(t, err)
    assert.Equal(t, 0, result.PaymentsGenerated) // No se crearon nuevos
    assert.Equal(t, 2, result.PaymentsExisting)   // Los 2 ya existían
}
```

---

## Checklist de Implementación

### Backend

- [ ] **Paso 1**: Verificar/añadir método `GetAllActive` en repositorio
- [ ] **Paso 2**: Añadir DTOs en `input/payment_service.go`
- [ ] **Paso 3**: Implementar `GenerateAnnualFees` en `payment_service.go`
- [ ] **Paso 4**: Actualizar schema GraphQL `payment.graphqls`
- [ ] **Paso 5**: Implementar resolver en `schema.resolvers.go`
- [ ] **Paso 6**: Regenerar código GraphQL con `gqlgen generate`
- [ ] **Paso 7**: Compilar y verificar: `go build ./...`
- [ ] **Paso 8**: Ejecutar tests: `go test ./...`

### Validaciones a Implementar

- [ ] ✅ Año no puede ser futuro
- [ ] ✅ Monto base debe ser positivo
- [ ] ✅ Extra familiar no puede ser negativo
- [ ] ✅ Solo admin puede ejecutar la operación
- [ ] ✅ Idempotencia (no duplicar pagos)
- [ ] ✅ Solo generar para socios activos

### Casos de Prueba

- [ ] Generación exitosa con socios individuales
- [ ] Generación exitosa con socios familiares
- [ ] Generación exitosa con mix de tipos
- [ ] Validación de año futuro
- [ ] Validación de montos negativos/cero
- [ ] Idempotencia (ejecutar dos veces)
- [ ] Generación sin socios activos
- [ ] Actualización de cuota existente con nuevos montos

---

## Troubleshooting

### Error: "GetAllActive" not found

**Solución**: Implementar el método en `member_repository.go` (ver Paso 1)

### Error: GraphQL schema mismatch

**Solución**: Ejecutar `go run github.com/99designs/gqlgen generate`

### Error: Pagos duplicados

**Verificar**: Lógica de idempotencia en `generatePaymentForMember`

### Performance: Lentitud con muchos socios

**Optimización**: Considerar batch inserts usando transacciones

```go
tx := s.paymentRepo.BeginTransaction(ctx)
defer tx.Rollback()

for _, member := range activeMembers {
    // Crear pagos con tx
    s.paymentRepo.CreateWithTx(ctx, tx, payment)
}

tx.Commit()
```

---

**Próximo paso**: Leer [Frontend - Instrucciones Detalladas](./frontend.md)

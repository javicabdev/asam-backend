# Conversión de Socio Individual a Familiar

## Overview

Este documento describe el plan de implementación para convertir un socio individual (prefijo B) a socio familiar (prefijo A), creando automáticamente la entidad `Family` asociada y ajustando los pagos pendientes.

## Decisiones de Negocio

### 1. Pagos Pendientes
**Decisión**: Ajustar el monto de pagos pendientes existentes (Opción B)

**Rationale**:
- Mantiene continuidad del registro de pago
- Más simple que cancelar + recrear
- Preserva el historial (fechas, notas, referencias)
- Menor impacto en relaciones con CashFlow

**Implementación**:
```go
// Para cada pago pendiente:
// 1. Calcular nuevo monto (cuota familiar)
// 2. Actualizar campo amount
// 3. Agregar nota explicativa del ajuste
```

### 2. Datos de Familia Obligatorios
**Decisión**: Solo esposo obligatorio (que es el socio original)

**Rationale**:
- Ya implementado en `ValidateConyugesFlexible`
- Permite familias monoparentales
- Esposa es opcional

### 3. Reversibilidad
**Decisión**: No reversible

**Rationale**:
- Simplifica lógica de implementación
- Caso de uso poco común
- Evita problemas con familiares dependientes
- Si es necesario revertir: dar de baja y crear nuevo socio individual

### 4. Número de Socio
**Decisión**: Asignar nuevo número correlativo con prefijo A

**Rationale**:
- Evita confusión (B00123 ≠ A00123)
- Auditoría clara del cambio
- No hay riesgo de colisión de números

**Ejemplo**:
```
Antes: B00123 (individual)
Después: A00456 (familiar) - nuevo número correlativo
```

---

## Arquitectura

### Modelos Involucrados

#### Member (`internal/domain/models/member.go`)
```go
type Member struct {
    ID               uint
    MembershipNumber string  // B→A (nuevo correlativo)
    MembershipType   string  // "individual" → "familiar"
    // ... resto de campos sin cambios
}
```

#### Family (`internal/domain/models/family.go`)
```go
type Family struct {
    ID              uint
    NumeroSocio     string  // Mismo que MembershipNumber del Member
    MiembroOrigenID *uint   // ID del Member convertido

    // Datos del esposo (obligatorios)
    EsposoNombre    string
    EsposoApellidos string

    // Datos de la esposa (opcionales)
    EsposaNombre    string
    EsposaApellidos string

    // Campos adicionales opcionales
    EsposoFechaNacimiento    *time.Time
    EsposoDocumentoIdentidad string
    EsposoCorreoElectronico  string
    EsposaFechaNacimiento    *time.Time
    EsposaDocumentoIdentidad string
    EsposaCorreoElectronico  string

    // Relaciones
    MiembroOrigen *Member
    Familiares    []Familiar
    Telefonos     []Telephone
}
```

#### Payment (`internal/domain/models/payment.go`)
```go
type Payment struct {
    MemberID uint    // Se mantiene igual (ID del socio)
    Amount   float64 // Se ajusta si Status == PENDING
    Status   PaymentStatus
    // ... resto de campos
}
```

---

## Plan de Implementación

### Fase 1: Backend - Repositorio

#### 1.1. Método en MemberRepository
**Archivo**: `internal/ports/output/member_repository.go`

```go
// GetLastByPrefix obtiene el último miembro creado con un prefijo específico
// Útil para generar números correlativos (A00001, A00002, etc.)
GetLastByPrefix(ctx context.Context, prefix string) (*models.Member, error)
```

**Implementación**: `internal/adapters/db/member_repository.go`

```go
func (r *memberRepository) GetLastByPrefix(ctx context.Context, prefix string) (*models.Member, error) {
    var member models.Member

    result := r.db.WithContext(ctx).
        Where("membership_number LIKE ?", prefix+"%").
        Order("membership_number DESC").
        First(&member)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, nil // No hay miembros con ese prefijo
        }
        return nil, appErrors.DB(result.Error, "error getting last member by prefix")
    }

    return &member, nil
}
```

#### 1.2. Método en PaymentRepository
**Archivo**: `internal/ports/output/payment_repository.go`

```go
// GetPendingByMember obtiene todos los pagos pendientes de un miembro
GetPendingByMember(ctx context.Context, memberID uint) ([]*models.Payment, error)

// UpdateAmountWithTx actualiza el monto de un pago dentro de una transacción
UpdateAmountWithTx(ctx context.Context, tx Transaction, paymentID uint, newAmount float64, notes string) error
```

**Implementación**: `internal/adapters/db/payment_repository.go`

```go
func (r *paymentRepository) GetPendingByMember(ctx context.Context, memberID uint) ([]*models.Payment, error) {
    var payments []*models.Payment

    result := r.db.WithContext(ctx).
        Where("member_id = ? AND status = ?", memberID, models.PaymentStatusPending).
        Order("created_at ASC").
        Find(&payments)

    if result.Error != nil {
        return nil, appErrors.DB(result.Error, "error getting pending payments")
    }

    return payments, nil
}

func (r *paymentRepository) UpdateAmountWithTx(ctx context.Context, tx Transaction, paymentID uint, newAmount float64, notes string) error {
    gormTx := tx.(*gorm.DB)

    result := gormTx.WithContext(ctx).
        Model(&models.Payment{}).
        Where("id = ?", paymentID).
        Updates(map[string]interface{}{
            "amount": newAmount,
            "notes":  notes,
        })

    if result.Error != nil {
        return appErrors.DB(result.Error, "error updating payment amount")
    }

    return nil
}
```

---

### Fase 2: Backend - Servicio

#### 2.1. Input DTO
**Archivo**: `internal/ports/input/member_types.go`

```go
// ConvertToFamilyRequest contiene los datos necesarios para convertir un socio
type ConvertToFamilyRequest struct {
    MemberID uint

    // Datos del esposo (obligatorios) - usualmente los del socio original
    EsposoNombre    string
    EsposoApellidos string

    // Datos de la esposa (opcionales)
    EsposaNombre    *string
    EsposaApellidos *string

    // Datos adicionales opcionales
    EsposoFechaNacimiento    *time.Time
    EsposoDocumentoIdentidad *string
    EsposoCorreoElectronico  *string
    EsposaFechaNacimiento    *time.Time
    EsposaDocumentoIdentidad *string
    EsposaCorreoElectronico  *string
}
```

#### 2.2. Interfaz del Servicio
**Archivo**: `internal/ports/input/member_service.go`

```go
// ConvertToFamily convierte un socio individual a familiar
ConvertToFamily(ctx context.Context, req *ConvertToFamilyRequest) (*ConvertToFamilyResponse, error)
```

**Response**:
```go
type ConvertToFamilyResponse struct {
    Member           *models.Member
    Family           *models.Family
    OldMemberNumber  string
    NewMemberNumber  string
    PaymentsAdjusted int
}
```

#### 2.3. Implementación del Servicio
**Archivo**: `internal/domain/services/member_service.go`

```go
func (s *memberService) ConvertToFamily(ctx context.Context, req *ConvertToFamilyRequest) (*ConvertToFamilyResponse, error) {
    // 1. Validar que el miembro existe y es individual
    member, err := s.memberRepository.GetByID(ctx, req.MemberID)
    if err != nil {
        return nil, errors.DB(err, "error obteniendo miembro")
    }
    if member == nil {
        return nil, errors.NotFound("Miembro no encontrado", nil)
    }
    if member.MembershipType != models.TipoMembresiaPIndividual {
        return nil, errors.New(errors.ErrInvalidOperation, "el miembro ya es de tipo familiar")
    }
    if !member.IsActive() {
        return nil, errors.New(errors.ErrInvalidOperation, "no se puede convertir un socio inactivo")
    }

    // 2. Validar datos de familia
    if req.EsposoNombre == "" || req.EsposoApellidos == "" {
        return nil, errors.NewValidationError(
            "Datos del esposo obligatorios",
            map[string]string{
                "esposoNombre": "El nombre del esposo es obligatorio",
                "esposoApellidos": "Los apellidos del esposo son obligatorios",
            },
        )
    }

    // 3. Iniciar transacción
    tx, err := s.memberRepository.BeginTransaction(ctx)
    if err != nil {
        return nil, errors.DB(err, "error iniciando transacción")
    }
    defer tx.Rollback()

    // 4. Generar nuevo número de socio (A)
    newMemberNumber, err := s.generateNextFamilyNumber(ctx)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrInternalError, "error generando número de socio")
    }

    // 5. Crear entrada en tabla families
    family := &models.Family{
        NumeroSocio:              newMemberNumber,
        MiembroOrigenID:          &req.MemberID,
        EsposoNombre:             req.EsposoNombre,
        EsposoApellidos:          req.EsposoApellidos,
        EsposaNombre:             ptrToString(req.EsposaNombre),
        EsposaApellidos:          ptrToString(req.EsposaApellidos),
        EsposoFechaNacimiento:    req.EsposoFechaNacimiento,
        EsposoDocumentoIdentidad: ptrToString(req.EsposoDocumentoIdentidad),
        EsposoCorreoElectronico:  ptrToString(req.EsposoCorreoElectronico),
        EsposaFechaNacimiento:    req.EsposaFechaNacimiento,
        EsposaDocumentoIdentidad: ptrToString(req.EsposaDocumentoIdentidad),
        EsposaCorreoElectronico:  ptrToString(req.EsposaCorreoElectronico),
    }

    // Validar datos de familia
    if err := family.Validate(); err != nil {
        return nil, err
    }

    // Crear familia con transacción
    if err := s.familyRepository.CreateWithTx(ctx, tx, family); err != nil {
        return nil, errors.DB(err, "error creando familia")
    }

    // 6. Actualizar el miembro
    oldMemberNumber := member.MembershipNumber
    member.MembershipNumber = newMemberNumber
    member.MembershipType = models.TipoMembresiaPFamiliar

    if err := s.memberRepository.UpdateWithTx(ctx, tx, member); err != nil {
        return nil, errors.DB(err, "error actualizando miembro")
    }

    // 7. Ajustar pagos pendientes
    paymentsAdjusted := 0
    pendingPayments, err := s.paymentRepository.GetPendingByMember(ctx, req.MemberID)
    if err != nil {
        return nil, errors.DB(err, "error obteniendo pagos pendientes")
    }

    for _, payment := range pendingPayments {
        // Obtener cuota familiar del año del pago
        year := time.Now().Year()
        if payment.CreatedAt.Year() > 0 {
            year = payment.CreatedAt.Year()
        }

        oldAmount := payment.Amount
        newAmount := s.feeCalculator.CalculateFamilyFee(year, 1) // mes no importa por ahora

        if oldAmount != newAmount {
            notes := payment.Notes
            if notes != "" {
                notes += " | "
            }
            notes += fmt.Sprintf("Ajustado de %.2f€ a %.2f€ por conversión a socio familiar", oldAmount, newAmount)

            if err := s.paymentRepository.UpdateAmountWithTx(ctx, tx, payment.ID, newAmount, notes); err != nil {
                return nil, errors.DB(err, "error ajustando pago")
            }
            paymentsAdjusted++
        }
    }

    // 8. Registrar en audit log
    s.auditLogger.LogAction(
        ctx,
        audit.ActionUpdate,
        audit.EntityMember,
        strconv.FormatUint(uint64(req.MemberID), 10),
        fmt.Sprintf("Convertido de individual (%s) a familiar (%s)", oldMemberNumber, newMemberNumber),
        map[string]interface{}{
            "old_number": oldMemberNumber,
            "new_number": newMemberNumber,
            "family_id":  family.ID,
            "payments_adjusted": paymentsAdjusted,
        },
    )

    // 9. Commit transacción
    if err := tx.Commit(); err != nil {
        return nil, errors.DB(err, "error confirmando transacción")
    }

    // 10. Retornar respuesta
    return &ConvertToFamilyResponse{
        Member:           member,
        Family:           family,
        OldMemberNumber:  oldMemberNumber,
        NewMemberNumber:  newMemberNumber,
        PaymentsAdjusted: paymentsAdjusted,
    }, nil
}

// generateNextFamilyNumber genera el siguiente número de socio familiar disponible
func (s *memberService) generateNextFamilyNumber(ctx context.Context) (string, error) {
    lastMember, err := s.memberRepository.GetLastByPrefix(ctx, "A")
    if err != nil {
        return "", err
    }

    nextNumber := 1
    if lastMember != nil {
        // Extraer número: "A00123" -> 123
        numStr := strings.TrimPrefix(lastMember.MembershipNumber, "A")
        num, parseErr := strconv.Atoi(numStr)
        if parseErr != nil {
            return "", errors.New(errors.ErrInternalError, "formato de número de socio inválido")
        }
        nextNumber = num + 1
    }

    // Formato: A00001, A00002, etc. (5 dígitos)
    return fmt.Sprintf("A%05d", nextNumber), nil
}

// Helper para convertir *string a string
func ptrToString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

---

### Fase 3: Backend - GraphQL

#### 3.1. Schema GraphQL
**Archivo**: `internal/adapters/gql/schema/member.graphqls`

```graphql
# Input para convertir socio a familiar
input ConvertToFamilyInput {
  memberId: ID!

  # Datos del esposo (obligatorios)
  esposoNombre: String!
  esposoApellidos: String!

  # Datos de la esposa (opcionales)
  esposaNombre: String
  esposaApellidos: String

  # Datos adicionales opcionales
  esposoFechaNacimiento: Time
  esposoDocumentoIdentidad: String
  esposoCorreoElectronico: String
  esposaFechaNacimiento: Time
  esposaDocumentoIdentidad: String
  esposaCorreoElectronico: String
}

# Respuesta de conversión
type ConvertToFamilyResponse {
  member: Member!
  family: Family!
  oldMemberNumber: String!
  newMemberNumber: String!
  paymentsAdjusted: Int!
}

extend type Mutation {
  convertMemberToFamily(input: ConvertToFamilyInput!): ConvertToFamilyResponse!
}
```

#### 3.2. Resolver
**Archivo**: `internal/adapters/gql/resolvers/schema.resolvers.go`

```go
// ConvertMemberToFamily is the resolver for the convertMemberToFamily field.
func (r *mutationResolver) ConvertMemberToFamily(ctx context.Context, input model.ConvertToFamilyInput) (*model.ConvertToFamilyResponse, error) {
    // Solo ADMIN puede convertir socios
    if err := middleware.MustBeAdmin(ctx); err != nil {
        return nil, err
    }

    // Parsear ID
    memberID, err := parseID(input.MemberID)
    if err != nil {
        return nil, err
    }

    // Construir request
    req := &input.ConvertToFamilyRequest{
        MemberID:        memberID,
        EsposoNombre:    input.EsposoNombre,
        EsposoApellidos: input.EsposoApellidos,
    }

    // Campos opcionales
    if input.EsposaNombre != nil {
        req.EsposaNombre = input.EsposaNombre
    }
    if input.EsposaApellidos != nil {
        req.EsposaApellidos = input.EsposaApellidos
    }
    if input.EsposoFechaNacimiento != nil {
        req.EsposoFechaNacimiento = input.EsposoFechaNacimiento
    }
    if input.EsposoDocumentoIdentidad != nil {
        req.EsposoDocumentoIdentidad = input.EsposoDocumentoIdentidad
    }
    if input.EsposoCorreoElectronico != nil {
        req.EsposoCorreoElectronico = input.EsposoCorreoElectronico
    }
    if input.EsposaFechaNacimiento != nil {
        req.EsposaFechaNacimiento = input.EsposaFechaNacimiento
    }
    if input.EsposaDocumentoIdentidad != nil {
        req.EsposaDocumentoIdentidad = input.EsposaDocumentoIdentidad
    }
    if input.EsposaCorreoElectronico != nil {
        req.EsposaCorreoElectronico = input.EsposaCorreoElectronico
    }

    // Llamar al servicio
    result, err := r.MemberService.ConvertToFamily(ctx, req)
    if err != nil {
        return nil, err
    }

    // Mapear respuesta
    return &model.ConvertToFamilyResponse{
        Member:           result.Member,
        Family:           result.Family,
        OldMemberNumber:  result.OldMemberNumber,
        NewMemberNumber:  result.NewMemberNumber,
        PaymentsAdjusted: result.PaymentsAdjusted,
    }, nil
}
```

---

### Fase 4: Frontend

#### 4.1. GraphQL Mutation
**Archivo**: `src/features/members/api/mutations.ts`

```typescript
import { gql } from '@apollo/client'

export const CONVERT_MEMBER_TO_FAMILY = gql`
  mutation ConvertMemberToFamily($input: ConvertToFamilyInput!) {
    convertMemberToFamily(input: $input) {
      member {
        miembro_id
        numero_socio
        tipo_membresia
        nombre
        apellidos
        estado
      }
      family {
        id
        numero_socio
        esposo_nombre
        esposo_apellidos
        esposa_nombre
        esposa_apellidos
      }
      oldMemberNumber
      newMemberNumber
      paymentsAdjusted
    }
  }
`
```

#### 4.2. Hook
**Archivo**: `src/features/members/hooks/useConvertToFamily.ts`

```typescript
import { useMutation } from '@apollo/client'
import { CONVERT_MEMBER_TO_FAMILY } from '../api/mutations'
import { GET_MEMBERS_QUERY } from '../api/queries'
import type { ConvertToFamilyInput } from '../types'

export const useConvertToFamily = () => {
  const [mutate, { loading, error }] = useMutation(CONVERT_MEMBER_TO_FAMILY, {
    refetchQueries: [{ query: GET_MEMBERS_QUERY }],
  })

  const convertToFamily = async (input: ConvertToFamilyInput) => {
    const result = await mutate({
      variables: { input },
    })
    return result.data?.convertMemberToFamily
  }

  return {
    convertToFamily,
    loading,
    error,
  }
}
```

#### 4.3. Componente de Diálogo
**Archivo**: `src/features/members/components/ConvertToFamilyDialog.tsx`

```typescript
import React, { useState } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Alert,
  Box,
  Typography,
  Grid,
} from '@mui/material'
import { useConvertToFamily } from '../hooks/useConvertToFamily'
import type { Member } from '../types'

interface Props {
  open: boolean
  member: Member | null
  onClose: () => void
  onSuccess: () => void
}

export const ConvertToFamilyDialog: React.FC<Props> = ({
  open,
  member,
  onClose,
  onSuccess,
}) => {
  const { convertToFamily, loading, error } = useConvertToFamily()

  const [formData, setFormData] = useState({
    esposoNombre: member?.nombre || '',
    esposoApellidos: member?.apellidos || '',
    esposaNombre: '',
    esposaApellidos: '',
    esposoDocumentoIdentidad: member?.documento_identidad || '',
    esposoCorreoElectronico: member?.correo_electronico || '',
    esposaDocumentoIdentidad: '',
    esposaCorreoElectronico: '',
  })

  const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, [field]: e.target.value }))
  }

  const handleSubmit = async () => {
    if (!member) return

    try {
      const result = await convertToFamily({
        memberId: member.miembro_id,
        esposoNombre: formData.esposoNombre,
        esposoApellidos: formData.esposoApellidos,
        esposaNombre: formData.esposaNombre || undefined,
        esposaApellidos: formData.esposaApellidos || undefined,
        esposoDocumentoIdentidad: formData.esposoDocumentoIdentidad || undefined,
        esposoCorreoElectronico: formData.esposoCorreoElectronico || undefined,
        esposaDocumentoIdentidad: formData.esposaDocumentoIdentidad || undefined,
        esposaCorreoElectronico: formData.esposaCorreoElectronico || undefined,
      })

      if (result) {
        onSuccess()
        onClose()
      }
    } catch (err) {
      console.error('Error converting to family:', err)
    }
  }

  if (!member) return null

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        Convertir a Socio Familiar
      </DialogTitle>

      <DialogContent>
        <Alert severity="info" sx={{ mb: 3 }}>
          <Typography variant="body2" fontWeight="bold">
            Cambios que se realizarán:
          </Typography>
          <Typography variant="body2">
            • Número de socio: {member.numero_socio} → Nuevo número con prefijo A
          </Typography>
          <Typography variant="body2">
            • Tipo: Individual → Familiar
          </Typography>
          <Typography variant="body2">
            • Los pagos pendientes se ajustarán al monto de cuota familiar
          </Typography>
        </Alert>

        <Alert severity="warning" sx={{ mb: 3 }}>
          <Typography variant="body2" fontWeight="bold">
            ⚠️ Este cambio no es reversible
          </Typography>
        </Alert>

        <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>
          Datos del Esposo (Obligatorios)
        </Typography>

        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Nombre del Esposo *"
              value={formData.esposoNombre}
              onChange={handleChange('esposoNombre')}
              required
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Apellidos del Esposo *"
              value={formData.esposoApellidos}
              onChange={handleChange('esposoApellidos')}
              required
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="DNI/NIE del Esposo"
              value={formData.esposoDocumentoIdentidad}
              onChange={handleChange('esposoDocumentoIdentidad')}
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Email del Esposo"
              type="email"
              value={formData.esposoCorreoElectronico}
              onChange={handleChange('esposoCorreoElectronico')}
            />
          </Grid>
        </Grid>

        <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
          Datos de la Esposa (Opcionales)
        </Typography>

        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Nombre de la Esposa"
              value={formData.esposaNombre}
              onChange={handleChange('esposaNombre')}
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Apellidos de la Esposa"
              value={formData.esposaApellidos}
              onChange={handleChange('esposaApellidos')}
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="DNI/NIE de la Esposa"
              value={formData.esposaDocumentoIdentidad}
              onChange={handleChange('esposaDocumentoIdentidad')}
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Email de la Esposa"
              type="email"
              value={formData.esposaCorreoElectronico}
              onChange={handleChange('esposaCorreoElectronico')}
            />
          </Grid>
        </Grid>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error.message}
          </Alert>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          Cancelar
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          color="primary"
          disabled={loading || !formData.esposoNombre || !formData.esposoApellidos}
        >
          {loading ? 'Convirtiendo...' : 'Convertir a Familiar'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
```

#### 4.4. Integración en MemberList
**Archivo**: `src/features/members/components/MemberList.tsx`

```typescript
// Agregar botón en el menú de acciones de cada socio individual
<MenuItem
  onClick={() => {
    handleCloseMenu()
    setConvertDialogOpen(true)
    setSelectedMember(member)
  }}
  disabled={member.tipo_membresia !== 'INDIVIDUAL' || member.estado !== 'ACTIVE'}
>
  <GroupAdd fontSize="small" sx={{ mr: 1 }} />
  Convertir a Familiar
</MenuItem>

// Agregar el diálogo
<ConvertToFamilyDialog
  open={convertDialogOpen}
  member={selectedMember}
  onClose={() => {
    setConvertDialogOpen(false)
    setSelectedMember(null)
  }}
  onSuccess={() => {
    // Mostrar notificación de éxito
    enqueueSnackbar('Socio convertido a familiar exitosamente', {
      variant: 'success',
    })
  }}
/>
```

---

### Fase 5: Testing

#### 5.1. Tests Unitarios - Servicio
**Archivo**: `test/unit/services/member_service_convert_test.go`

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
)

func TestConvertToFamily(t *testing.T) {
    t.Run("Success - Convert individual to family", func(t *testing.T) {
        // Setup
        ctx := context.Background()

        // Mock repositories
        memberRepo := &MockMemberRepository{
            GetByIDFunc: func(ctx context.Context, id uint) (*models.Member, error) {
                return &models.Member{
                    ID:               1,
                    MembershipNumber: "B00123",
                    MembershipType:   models.TipoMembresiaPIndividual,
                    State:            models.EstadoActivo,
                    Name:             "Juan",
                    Surnames:         "Pérez",
                }, nil
            },
            GetLastByPrefixFunc: func(ctx context.Context, prefix string) (*models.Member, error) {
                if prefix == "A" {
                    return &models.Member{MembershipNumber: "A00099"}, nil
                }
                return nil, nil
            },
            UpdateWithTxFunc: func(ctx context.Context, tx output.Transaction, member *models.Member) error {
                return nil
            },
            BeginTransactionFunc: func(ctx context.Context) (output.Transaction, error) {
                return &MockTransaction{}, nil
            },
        }

        familyRepo := &MockFamilyRepository{
            CreateWithTxFunc: func(ctx context.Context, tx output.Transaction, family *models.Family) error {
                family.ID = 1
                return nil
            },
        }

        paymentRepo := &MockPaymentRepository{
            GetPendingByMemberFunc: func(ctx context.Context, memberID uint) ([]*models.Payment, error) {
                return []*models.Payment{
                    {
                        ID:        1,
                        MemberID:  1,
                        Amount:    10.0,
                        Status:    models.PaymentStatusPending,
                        CreatedAt: time.Now(),
                    },
                }, nil
            },
            UpdateAmountWithTxFunc: func(ctx context.Context, tx output.Transaction, paymentID uint, newAmount float64, notes string) error {
                return nil
            },
        }

        feeCalculator := services.NewFeeCalculator(10.0, 5.0, 0.1, 10.0)

        service := services.NewMemberService(memberRepo, familyRepo, paymentRepo, nil, feeCalculator)

        // Execute
        req := &input.ConvertToFamilyRequest{
            MemberID:        1,
            EsposoNombre:    "Juan",
            EsposoApellidos: "Pérez",
        }

        result, err := service.ConvertToFamily(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, "B00123", result.OldMemberNumber)
        assert.Equal(t, "A00100", result.NewMemberNumber)
        assert.Equal(t, 1, result.PaymentsAdjusted)
        assert.Equal(t, models.TipoMembresiaPFamiliar, result.Member.MembershipType)
    })

    t.Run("Error - Member not found", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetByIDFunc: func(ctx context.Context, id uint) (*models.Member, error) {
                return nil, nil
            },
        }

        service := services.NewMemberService(memberRepo, nil, nil, nil, nil)

        req := &input.ConvertToFamilyRequest{
            MemberID:        999,
            EsposoNombre:    "Juan",
            EsposoApellidos: "Pérez",
        }

        result, err := service.ConvertToFamily(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "no encontrado")
    })

    t.Run("Error - Member already family type", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetByIDFunc: func(ctx context.Context, id uint) (*models.Member, error) {
                return &models.Member{
                    ID:               1,
                    MembershipNumber: "A00123",
                    MembershipType:   models.TipoMembresiaPFamiliar,
                    State:            models.EstadoActivo,
                }, nil
            },
        }

        service := services.NewMemberService(memberRepo, nil, nil, nil, nil)

        req := &input.ConvertToFamilyRequest{
            MemberID:        1,
            EsposoNombre:    "Juan",
            EsposoApellidos: "Pérez",
        }

        result, err := service.ConvertToFamily(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "ya es de tipo familiar")
    })

    t.Run("Error - Member inactive", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetByIDFunc: func(ctx context.Context, id uint) (*models.Member, error) {
                return &models.Member{
                    ID:               1,
                    MembershipNumber: "B00123",
                    MembershipType:   models.TipoMembresiaPIndividual,
                    State:            models.EstadoInactivo,
                }, nil
            },
        }

        service := services.NewMemberService(memberRepo, nil, nil, nil, nil)

        req := &input.ConvertToFamilyRequest{
            MemberID:        1,
            EsposoNombre:    "Juan",
            EsposoApellidos: "Pérez",
        }

        result, err := service.ConvertToFamily(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "inactivo")
    })

    t.Run("Error - Missing required spouse data", func(t *testing.T) {
        ctx := context.Background()

        memberRepo := &MockMemberRepository{
            GetByIDFunc: func(ctx context.Context, id uint) (*models.Member, error) {
                return &models.Member{
                    ID:               1,
                    MembershipNumber: "B00123",
                    MembershipType:   models.TipoMembresiaPIndividual,
                    State:            models.EstadoActivo,
                }, nil
            },
        }

        service := services.NewMemberService(memberRepo, nil, nil, nil, nil)

        req := &input.ConvertToFamilyRequest{
            MemberID:        1,
            EsposoNombre:    "", // Missing
            EsposoApellidos: "Pérez",
        }

        result, err := service.ConvertToFamily(ctx, req)

        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "obligatorio")
    })
}
```

#### 5.2. Tests de Integración
**Archivo**: `test/integration/convert_member_integration_test.go`

```go
// Test completo end-to-end con base de datos real
func TestConvertMemberToFamily_Integration(t *testing.T) {
    // Setup database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create test member
    member := createTestMember(t, db, "B00123", models.TipoMembresiaPIndividual)

    // Create pending payment
    payment := createTestPayment(t, db, member.ID, 10.0, models.PaymentStatusPending)

    // Execute conversion
    service := setupMemberService(t, db)
    result, err := service.ConvertToFamily(context.Background(), &input.ConvertToFamilyRequest{
        MemberID:        member.ID,
        EsposoNombre:    "Juan",
        EsposoApellidos: "Pérez",
    })

    // Verify
    assert.NoError(t, err)
    assert.NotNil(t, result)

    // Verify member updated
    updatedMember := getMemberFromDB(t, db, member.ID)
    assert.Equal(t, models.TipoMembresiaPFamiliar, updatedMember.MembershipType)
    assert.True(t, strings.HasPrefix(updatedMember.MembershipNumber, "A"))

    // Verify family created
    family := getFamilyByMemberID(t, db, member.ID)
    assert.NotNil(t, family)
    assert.Equal(t, "Juan", family.EsposoNombre)
    assert.Equal(t, "Pérez", family.EsposoApellidos)

    // Verify payment adjusted
    updatedPayment := getPaymentFromDB(t, db, payment.ID)
    assert.Equal(t, 15.0, updatedPayment.Amount) // 10 + 5 (family extra)
    assert.Contains(t, updatedPayment.Notes, "Ajustado")
}
```

---

## Casos de Prueba Manual

### Caso 1: Conversión Exitosa
1. Crear socio individual B00123
2. Crear pago pendiente de 10€
3. Convertir a familiar con datos del esposo
4. Verificar:
   - ✅ Nuevo número A00XXX
   - ✅ Tipo cambiado a familiar
   - ✅ Familia creada
   - ✅ Pago ajustado a 15€
   - ✅ Nota en pago explicando el ajuste

### Caso 2: Validación - Socio Inactivo
1. Crear socio individual inactivo
2. Intentar convertir
3. Verificar: ❌ Error "no se puede convertir un socio inactivo"

### Caso 3: Validación - Ya es Familiar
1. Crear socio familiar A00123
2. Intentar convertir
3. Verificar: ❌ Error "el miembro ya es de tipo familiar"

### Caso 4: Validación - Datos Incompletos
1. Crear socio individual
2. Intentar convertir sin nombre de esposo
3. Verificar: ❌ Error de validación

### Caso 5: Múltiples Pagos Pendientes
1. Crear socio individual
2. Crear 3 pagos pendientes (2024, 2025, 2026)
3. Convertir a familiar
4. Verificar: ✅ Los 3 pagos ajustados

### Caso 6: Sin Pagos Pendientes
1. Crear socio individual
2. Sin pagos o todos pagados
3. Convertir a familiar
4. Verificar: ✅ Conversión exitosa, 0 pagos ajustados

---

## Rollout Plan

### Fase 1: Desarrollo (Semana 1)
- [ ] Implementar métodos en repositorios
- [ ] Implementar servicio `ConvertToFamily`
- [ ] Tests unitarios del servicio
- [ ] Tests de repositorio

### Fase 2: GraphQL + Frontend (Semana 1-2)
- [ ] Schema GraphQL
- [ ] Resolver GraphQL
- [ ] Componente de diálogo frontend
- [ ] Integración en lista de socios
- [ ] Traducciones

### Fase 3: Testing (Semana 2)
- [ ] Tests de integración
- [ ] Tests manuales E2E
- [ ] Pruebas de regresión (verificar que no se rompió nada)

### Fase 4: Deploy (Semana 2-3)
- [ ] Code review
- [ ] Merge a main
- [ ] Deploy a staging
- [ ] Testing en staging
- [ ] Deploy a producción
- [ ] Monitoreo post-deploy

---

## Consideraciones Adicionales

### Seguridad
- ✅ Solo ADMIN puede ejecutar la conversión
- ✅ Validación de datos en backend (no confiar en frontend)
- ✅ Transacción atómica (todo o nada)
- ✅ Audit log de la operación

### Performance
- ✅ Conversión rápida (< 1 segundo típicamente)
- ✅ Transacción única para evitar locks prolongados
- ✅ Solo se ajustan pagos PENDING (no todos los pagos históricos)

### UX
- ⚠️ Advertencia clara de irreversibilidad
- ✅ Mostrar preview de cambios antes de confirmar
- ✅ Feedback inmediato del resultado
- ✅ Recarga automática de lista de socios

### Monitoreo
- Log cuando se realiza una conversión
- Métrica: número de conversiones por mes
- Alerta si hay muchos rollbacks de transacciones

---

## Estimación Final

| Fase | Tiempo Estimado | Complejidad |
|------|----------------|-------------|
| Backend (Repo + Service) | 1-2 días | Media |
| GraphQL | 0.5 días | Baja |
| Frontend | 1 día | Media |
| Testing | 1 día | Media |
| **TOTAL** | **3-4 días** | **Media** |

---

## Preguntas Pendientes

1. ¿Se debe enviar notificación por email al socio cuando se convierte?
2. ¿Se debe permitir conversión si hay pagos vencidos (PENDING antiguos)?
3. ¿Qué hacer si el número de socio A correlativo ya existe (caso edge)?
4. ¿Se necesita algún reporte de conversiones realizadas?

---

**Fecha de creación**: 2025-11-07
**Autor**: Claude Code
**Estado**: Pendiente de implementación

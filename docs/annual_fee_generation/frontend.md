# Frontend - Generación de Cuotas Anuales

## Índice

1. [Arquitectura](#arquitectura)
2. [Implementación Paso a Paso](#implementación-paso-a-paso)
3. [Código Completo](#código-completo)
4. [UI/UX](#uiux)

---

## Arquitectura

### Estructura de Archivos

```
src/features/payments/
├── api/
│   ├── mutations.ts                  [MODIFICAR]
│   └── queries.ts                    [VERIFICAR]
├── components/
│   ├── GenerateFeesDialog.tsx        [CREAR]
│   └── GenerateFeesButton.tsx        [CREAR]
├── hooks/
│   ├── useGenerateAnnualFees.ts      [CREAR]
│   └── index.ts                      [MODIFICAR]
├── types.ts                          [MODIFICAR]
└── index.ts                          [MODIFICAR]
```

### Flujo de Usuario

```
┌─────────────────────────┐
│  PaymentsPage           │
│  - Ver lista de pagos   │
│  - Botón "Generar       │
│    Cuotas Anuales"      │
└────────┬────────────────┘
         │ Click
         ▼
┌─────────────────────────┐
│  GenerateFeesDialog     │
│  - Input: Año           │
│  - Input: Monto base    │
│  - Input: Extra familiar│
│  - Botón: Generar       │
└────────┬────────────────┘
         │ Submit
         ▼
┌─────────────────────────┐
│  useGenerateAnnualFees  │
│  - Mutation GraphQL     │
│  - Loading state        │
│  - Error handling       │
└────────┬────────────────┘
         │ Success
         ▼
┌─────────────────────────┐
│  ResultDialog           │
│  - N pagos generados    │
│  - N pagos existentes   │
│  - Botón: Ver pagos     │
└─────────────────────────┘
```

---

## Implementación Paso a Paso

### PASO 1: Añadir Types

**Archivo**: `src/features/payments/types.ts`

**Añadir al final del archivo**:

```typescript
/**
 * Input para generar cuotas anuales
 */
export interface GenerateAnnualFeesInput {
  year: number
  baseFeeAmount: number
  familyFeeExtra: number
}

/**
 * Detalle de generación por socio
 */
export interface PaymentGenerationDetail {
  memberId: string
  memberNumber: string
  memberName: string
  amount: number
  wasCreated: boolean
  error?: string | null
}

/**
 * Respuesta de generación de cuotas
 */
export interface GenerateAnnualFeesResponse {
  year: number
  membershipFeeId: string
  paymentsGenerated: number
  paymentsExisting: number
  totalMembers: number
  details: PaymentGenerationDetail[]
}
```

---

### PASO 2: Añadir Mutation GraphQL

**Archivo**: `src/features/payments/api/mutations.ts`

**Añadir al final del archivo**:

```typescript
import { gql } from '@apollo/client'

// ... mutations existentes ...

export const GENERATE_ANNUAL_FEES = gql`
  mutation GenerateAnnualFees($input: GenerateAnnualFeesInput!) {
    generateAnnualFees(input: $input) {
      year
      membershipFeeId
      paymentsGenerated
      paymentsExisting
      totalMembers
      details {
        memberId
        memberNumber
        memberName
        amount
        wasCreated
        error
      }
    }
  }
`
```

---

### PASO 3: Crear Hook

**Archivo**: `src/features/payments/hooks/useGenerateAnnualFees.ts`

```typescript
import { useMutation } from '@apollo/client'
import { GENERATE_ANNUAL_FEES } from '../api/mutations'
import { GET_PAYMENTS_QUERY } from '../api/queries'
import type {
  GenerateAnnualFeesInput,
  GenerateAnnualFeesResponse,
} from '../types'

export const useGenerateAnnualFees = () => {
  const [mutate, { loading, error }] = useMutation(GENERATE_ANNUAL_FEES, {
    // Refetch payments after generation
    refetchQueries: [{ query: GET_PAYMENTS_QUERY }],
  })

  const generateFees = async (
    input: GenerateAnnualFeesInput
  ): Promise<GenerateAnnualFeesResponse> => {
    const result = await mutate({
      variables: { input },
    })

    if (!result.data?.generateAnnualFees) {
      throw new Error('No se recibió respuesta del servidor')
    }

    return result.data.generateAnnualFees
  }

  return {
    generateFees,
    loading,
    error,
  }
}
```

**Exportar en** `src/features/payments/hooks/index.ts`:

```typescript
export * from './useGenerateAnnualFees'
```

---

### PASO 4: Crear Componente de Diálogo

**Archivo**: `src/features/payments/components/GenerateFeesDialog.tsx`

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
  Divider,
  CircularProgress,
} from '@mui/material'
import { useGenerateAnnualFees } from '../hooks'
import type { GenerateAnnualFeesResponse } from '../types'

interface Props {
  open: boolean
  onClose: () => void
  onSuccess?: (result: GenerateAnnualFeesResponse) => void
}

export const GenerateFeesDialog: React.FC<Props> = ({
  open,
  onClose,
  onSuccess,
}) => {
  const currentYear = new Date().getFullYear()

  const [formData, setFormData] = useState({
    year: currentYear,
    baseFeeAmount: 40,
    familyFeeExtra: 10,
  })

  const [result, setResult] = useState<GenerateAnnualFeesResponse | null>(null)
  const [showResult, setShowResult] = useState(false)

  const { generateFees, loading, error } = useGenerateAnnualFees()

  const handleChange = (field: keyof typeof formData) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = parseFloat(e.target.value)
    setFormData((prev) => ({ ...prev, [field]: value }))
  }

  const handleSubmit = async () => {
    try {
      const response = await generateFees(formData)
      setResult(response)
      setShowResult(true)

      if (onSuccess) {
        onSuccess(response)
      }
    } catch (err) {
      console.error('Error generating fees:', err)
    }
  }

  const handleClose = () => {
    setShowResult(false)
    setResult(null)
    onClose()
  }

  const handleBackToForm = () => {
    setShowResult(false)
    setResult(null)
  }

  // Validaciones
  const yearError = formData.year > currentYear
  const amountError = formData.baseFeeAmount <= 0
  const extraError = formData.familyFeeExtra < 0
  const hasErrors = yearError || amountError || extraError

  // Mostrar resultado
  if (showResult && result) {
    return (
      <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
        <DialogTitle>Cuotas Generadas Exitosamente</DialogTitle>

        <DialogContent>
          <Alert severity="success" sx={{ mb: 3 }}>
            <Typography variant="body1" fontWeight="bold">
              ✓ Generación completada
            </Typography>
          </Alert>

          <Grid container spacing={2}>
            <Grid item xs={6}>
              <Box
                sx={{
                  p: 2,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  textAlign: 'center',
                }}
              >
                <Typography variant="h3" color="success.main">
                  {result.paymentsGenerated}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Pagos nuevos creados
                </Typography>
              </Box>
            </Grid>

            <Grid item xs={6}>
              <Box
                sx={{
                  p: 2,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  textAlign: 'center',
                }}
              >
                <Typography variant="h3" color="info.main">
                  {result.paymentsExisting}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Pagos ya existentes
                </Typography>
              </Box>
            </Grid>

            <Grid item xs={12}>
              <Divider sx={{ my: 2 }} />
            </Grid>

            <Grid item xs={12}>
              <Typography variant="body2" color="text.secondary">
                <strong>Año:</strong> {result.year}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                <strong>Total de socios procesados:</strong>{' '}
                {result.totalMembers}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                <strong>Monto base:</strong> {formData.baseFeeAmount}€
              </Typography>
              <Typography variant="body2" color="text.secondary">
                <strong>Extra familiar:</strong> {formData.familyFeeExtra}€
              </Typography>
            </Grid>

            {result.details && result.details.length > 0 && (
              <Grid item xs={12}>
                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                  Detalle por Socio
                </Typography>
                <Box
                  sx={{
                    maxHeight: 200,
                    overflow: 'auto',
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: 1,
                    p: 1,
                  }}
                >
                  {result.details.map((detail) => (
                    <Box
                      key={detail.memberId}
                      sx={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        py: 0.5,
                        borderBottom: '1px solid',
                        borderColor: 'divider',
                      }}
                    >
                      <Typography variant="body2">
                        {detail.memberNumber} - {detail.memberName}
                      </Typography>
                      <Typography
                        variant="body2"
                        color={detail.wasCreated ? 'success.main' : 'info.main'}
                      >
                        {detail.amount}€{' '}
                        {detail.wasCreated ? '(nuevo)' : '(ya existía)'}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              </Grid>
            )}
          </Grid>
        </DialogContent>

        <DialogActions>
          <Button onClick={handleBackToForm} variant="outlined">
            Generar Otro Año
          </Button>
          <Button onClick={handleClose} variant="contained" color="primary">
            Cerrar
          </Button>
        </DialogActions>
      </Dialog>
    )
  }

  // Mostrar formulario
  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Generar Cuotas Anuales</DialogTitle>

      <DialogContent>
        <Alert severity="info" sx={{ mb: 3 }}>
          Esta operación creará pagos pendientes de cuota anual para todos los
          socios activos del año seleccionado.
        </Alert>

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label="Año"
              type="number"
              value={formData.year}
              onChange={handleChange('year')}
              error={yearError}
              helperText={
                yearError
                  ? 'No se pueden generar cuotas de años futuros'
                  : `Máximo: ${currentYear}`
              }
              inputProps={{
                min: 2000,
                max: currentYear,
              }}
            />
          </Grid>

          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Monto Base (€)"
              type="number"
              value={formData.baseFeeAmount}
              onChange={handleChange('baseFeeAmount')}
              error={amountError}
              helperText={
                amountError
                  ? 'El monto debe ser mayor a 0'
                  : 'Para socios individuales'
              }
              inputProps={{
                min: 0,
                step: 0.5,
              }}
            />
          </Grid>

          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Extra Familiar (€)"
              type="number"
              value={formData.familyFeeExtra}
              onChange={handleChange('familyFeeExtra')}
              error={extraError}
              helperText={
                extraError
                  ? 'No puede ser negativo'
                  : 'Adicional para familias'
              }
              inputProps={{
                min: 0,
                step: 0.5,
              }}
            />
          </Grid>

          <Grid item xs={12}>
            <Box
              sx={{
                p: 2,
                bgcolor: 'background.default',
                borderRadius: 1,
              }}
            >
              <Typography variant="body2" fontWeight="bold" gutterBottom>
                Vista previa:
              </Typography>
              <Typography variant="body2">
                • Socio individual: {formData.baseFeeAmount.toFixed(2)}€
              </Typography>
              <Typography variant="body2">
                • Socio familiar:{' '}
                {(formData.baseFeeAmount + formData.familyFeeExtra).toFixed(2)}€
              </Typography>
            </Box>
          </Grid>
        </Grid>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error.message}
          </Alert>
        )}

        <Alert severity="warning" sx={{ mt: 2 }}>
          <Typography variant="body2" fontWeight="bold">
            ⚠️ Importante
          </Typography>
          <Typography variant="body2">
            • Solo se generarán cuotas para socios activos
          </Typography>
          <Typography variant="body2">
            • Si ya existen cuotas para este año, no se crearán duplicados
          </Typography>
        </Alert>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          Cancelar
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          color="primary"
          disabled={loading || hasErrors}
          startIcon={loading && <CircularProgress size={20} />}
        >
          {loading ? 'Generando...' : 'Generar Cuotas'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
```

---

### PASO 5: Crear Botón de Acción

**Archivo**: `src/features/payments/components/GenerateFeesButton.tsx`

```typescript
import React, { useState } from 'react'
import { Button } from '@mui/material'
import { Add as AddIcon } from '@mui/icons-material'
import { GenerateFeesDialog } from './GenerateFeesDialog'
import { useSnackbar } from 'notistack'
import type { GenerateAnnualFeesResponse } from '../types'

export const GenerateFeesButton: React.FC = () => {
  const [dialogOpen, setDialogOpen] = useState(false)
  const { enqueueSnackbar } = useSnackbar()

  const handleSuccess = (result: GenerateAnnualFeesResponse) => {
    enqueueSnackbar(
      `${result.paymentsGenerated} cuotas generadas exitosamente para ${result.year}`,
      { variant: 'success' }
    )
  }

  return (
    <>
      <Button
        variant="contained"
        color="primary"
        startIcon={<AddIcon />}
        onClick={() => setDialogOpen(true)}
      >
        Generar Cuotas Anuales
      </Button>

      <GenerateFeesDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        onSuccess={handleSuccess}
      />
    </>
  )
}
```

**Exportar en** `src/features/payments/components/index.ts`:

```typescript
export * from './GenerateFeesDialog'
export * from './GenerateFeesButton'
```

---

### PASO 6: Integrar en Página de Pagos

**Archivo**: Buscar donde se muestra la lista de pagos (probablemente `src/pages/payments/index.tsx` o similar)

**Añadir el botón en la barra de acciones**:

```typescript
import { GenerateFeesButton } from '@/features/payments'

// Dentro del componente, en la sección de acciones/toolbar:
<Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
  {/* Otros botones/filtros existentes */}

  <GenerateFeesButton />
</Box>
```

---

## Código Completo de Referencia

### GenerateFeesDialog.tsx - Versión Simplificada

Si prefieres una versión más compacta sin mostrar detalles:

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
  Grid,
  CircularProgress,
} from '@mui/material'
import { useGenerateAnnualFees } from '../hooks'
import { useSnackbar } from 'notistack'

interface Props {
  open: boolean
  onClose: () => void
}

export const GenerateFeesDialog: React.FC<Props> = ({ open, onClose }) => {
  const currentYear = new Date().getFullYear()
  const { enqueueSnackbar } = useSnackbar()

  const [formData, setFormData] = useState({
    year: currentYear,
    baseFeeAmount: 40,
    familyFeeExtra: 10,
  })

  const { generateFees, loading, error } = useGenerateAnnualFees()

  const handleSubmit = async () => {
    try {
      const result = await generateFees(formData)

      enqueueSnackbar(
        `${result.paymentsGenerated} cuotas generadas para ${result.year}`,
        { variant: 'success' }
      )

      onClose()
    } catch (err) {
      console.error('Error generating fees:', err)
    }
  }

  // Validaciones
  const yearError = formData.year > currentYear
  const amountError = formData.baseFeeAmount <= 0
  const hasErrors = yearError || amountError

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Generar Cuotas Anuales</DialogTitle>

      <DialogContent>
        <Alert severity="info" sx={{ mb: 3 }}>
          Se crearán pagos pendientes para todos los socios activos
        </Alert>

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label="Año"
              type="number"
              value={formData.year}
              onChange={(e) =>
                setFormData({ ...formData, year: parseInt(e.target.value) })
              }
              error={yearError}
              helperText={yearError && 'No puede ser año futuro'}
            />
          </Grid>

          <Grid item xs={6}>
            <TextField
              fullWidth
              label="Monto Base (€)"
              type="number"
              value={formData.baseFeeAmount}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  baseFeeAmount: parseFloat(e.target.value),
                })
              }
              error={amountError}
            />
          </Grid>

          <Grid item xs={6}>
            <TextField
              fullWidth
              label="Extra Familiar (€)"
              type="number"
              value={formData.familyFeeExtra}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  familyFeeExtra: parseFloat(e.target.value),
                })
              }
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
          disabled={loading || hasErrors}
          startIcon={loading && <CircularProgress size={20} />}
        >
          Generar
        </Button>
      </DialogActions>
    </Dialog>
  )
}
```

---

## UI/UX

### Ubicación del Botón

**Opción 1 (Recomendada)**: En la página de Pagos, junto a los filtros

```
┌─────────────────────────────────────────────────────┐
│  Pagos                                              │
├─────────────────────────────────────────────────────┤
│  [Filtros...] [Buscar...] [+ Generar Cuotas]       │
│                                                      │
│  Tabla de Pagos...                                  │
└─────────────────────────────────────────────────────┘
```

**Opción 2**: En un menú de administración

```
┌─────────────────────────────────────────────────────┐
│  Administración                                     │
├─────────────────────────────────────────────────────┤
│  □ Generar Cuotas Anuales                          │
│  □ Importar Datos desde Excel                      │
│  □ Exportar Reportes                               │
└─────────────────────────────────────────────────────┘
```

### Estados del Diálogo

#### 1. Estado Inicial (Formulario)
```
┌────────────────────────────────┐
│ Generar Cuotas Anuales         │
├────────────────────────────────┤
│ ℹ️ Se crearán pagos pendientes│
│                                 │
│ Año: [2024 ▼]                  │
│ Monto Base: [40.00€]           │
│ Extra Familiar: [10.00€]       │
│                                 │
│ Vista previa:                  │
│ • Individual: 40.00€           │
│ • Familiar: 50.00€             │
│                                 │
│ ⚠️ Solo socios activos         │
│                                 │
│ [Cancelar] [Generar]           │
└────────────────────────────────┘
```

#### 2. Estado Loading
```
┌────────────────────────────────┐
│ Generando cuotas...            │
├────────────────────────────────┤
│                                 │
│        [⏳ Spinner]             │
│                                 │
│  Procesando 50 socios...       │
│                                 │
└────────────────────────────────┘
```

#### 3. Estado Éxito
```
┌────────────────────────────────┐
│ Cuotas Generadas ✓             │
├────────────────────────────────┤
│ ✓ Generación completada        │
│                                 │
│   45                 5          │
│ Nuevos           Existentes    │
│                                 │
│ Año: 2024                      │
│ Total procesados: 50           │
│                                 │
│ [Generar Otro] [Cerrar]        │
└────────────────────────────────┘
```

#### 4. Estado Error
```
┌────────────────────────────────┐
│ Generar Cuotas Anuales         │
├────────────────────────────────┤
│ ❌ Error al generar cuotas     │
│                                 │
│ No se pueden generar cuotas    │
│ para años futuros              │
│                                 │
│ [Cerrar]                       │
└────────────────────────────────┘
```

---

## Checklist de Implementación

### Frontend

- [ ] **Paso 1**: Añadir types en `types.ts`
- [ ] **Paso 2**: Añadir mutation en `api/mutations.ts`
- [ ] **Paso 3**: Crear hook `useGenerateAnnualFees.ts`
- [ ] **Paso 4**: Crear componente `GenerateFeesDialog.tsx`
- [ ] **Paso 5**: Crear componente `GenerateFeesButton.tsx`
- [ ] **Paso 6**: Integrar botón en página de pagos
- [ ] **Paso 7**: Probar flujo completo

### Validaciones del Formulario

- [ ] ✅ Año no puede ser futuro
- [ ] ✅ Año no puede ser anterior a 2000 (límite razonable)
- [ ] ✅ Monto base debe ser mayor a 0
- [ ] ✅ Extra familiar no puede ser negativo
- [ ] ✅ Deshabilitar botón si hay errores
- [ ] ✅ Mostrar errores en tiempo real

### UX

- [ ] ✅ Loading state durante generación
- [ ] ✅ Mensaje de éxito con resumen
- [ ] ✅ Manejo de errores con mensajes claros
- [ ] ✅ Refetch de lista de pagos después de generar
- [ ] ✅ Vista previa de montos antes de confirmar
- [ ] ✅ Opción de generar otro año sin cerrar

---

## Testing

### Casos de Prueba Manual

#### Caso 1: Generación Exitosa
1. Navegar a Pagos
2. Click en "Generar Cuotas Anuales"
3. Seleccionar año actual
4. Ingresar monto base: 40€
5. Ingresar extra familiar: 10€
6. Click "Generar"
7. **Verificar**: Mensaje de éxito
8. **Verificar**: Lista de pagos se actualiza
9. **Verificar**: Aparecen N nuevos pagos PENDING

#### Caso 2: Validación de Año Futuro
1. Abrir diálogo
2. Ingresar año 2030
3. **Verificar**: Campo muestra error
4. **Verificar**: Botón "Generar" deshabilitado

#### Caso 3: Validación de Monto Inválido
1. Abrir diálogo
2. Ingresar monto base: 0
3. **Verificar**: Campo muestra error
4. **Verificar**: Botón "Generar" deshabilitado

#### Caso 4: Idempotencia
1. Generar cuotas de 2024
2. **Resultado**: "45 nuevos, 0 existentes"
3. Generar cuotas de 2024 nuevamente
4. **Resultado**: "0 nuevos, 45 existentes"

#### Caso 5: Generación de Años Pasados (Migración)
1. Generar cuotas de 2020
2. Generar cuotas de 2021
3. Generar cuotas de 2022
4. **Verificar**: Cada socio tiene 3 cuotas pendientes

---

## Troubleshooting

### Error: "Network error"

**Causa**: Backend no está corriendo o URL incorrecta

**Solución**: Verificar que el backend esté levantado en el puerto correcto

### Error: "Unauthorized"

**Causa**: Usuario no es administrador

**Solución**: Login como admin o verificar permisos en backend

### Los pagos no se actualizan en la lista

**Causa**: Falta refetchQueries en la mutation

**Solución**: Verificar que `refetchQueries: [{ query: GET_PAYMENTS_QUERY }]` esté en el hook

### El diálogo no muestra errores del backend

**Causa**: El componente no está mostrando `error` del hook

**Solución**: Añadir `{error && <Alert severity="error">{error.message}</Alert>}`

---

## Mejoras Futuras (Opcional)

### 1. Batch Progress
Mostrar progreso en tiempo real mientras se generan cuotas:

```typescript
// Backend: Usar websockets o polling
// Frontend: Mostrar barra de progreso

Procesando: 25/50 socios (50%)
[████████░░░░░░░░] 50%
```

### 2. Vista Previa de Socios
Mostrar lista de socios antes de generar:

```typescript
Se generarán cuotas para:
- B00001 Juan Pérez (40€)
- A00001 María García (50€)
- ...
Total: 50 socios
```

### 3. Filtros Avanzados
Permitir filtrar qué socios incluir:

```typescript
□ Solo socios individuales
□ Solo socios familiares
□ Excluir socios con deudas
```

### 4. Programación
Permitir programar generación automática:

```typescript
Generar automáticamente cada 1 de enero
□ Activar generación automática
```

---

**Próximo paso**: Leer [Testing](./testing.md)

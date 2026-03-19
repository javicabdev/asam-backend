# Guía de Integración - Módulo de Cobros (Payments)

## 📋 Índice
1. [Paginación](#paginación)
2. [Filtros Disponibles](#filtros-disponibles)
3. [Estados del Pago](#estados-del-pago)
4. [Flujo de Trabajo](#flujo-de-trabajo)
5. [Queries Disponibles](#queries-disponibles)
6. [Mutations Disponibles](#mutations-disponibles)
7. [Comportamientos Importantes](#comportamientos-importantes)
8. [Ejemplos de Uso](#ejemplos-de-uso)

---

## 🔢 Paginación

### Estructura de Respuesta
Todos los listados de pagos usan `PaymentConnection` con la siguiente estructura:

```graphql
type PaymentConnection {
    nodes: [Payment!]!      # Lista de pagos
    pageInfo: PageInfo!      # Información de paginación
}

type PageInfo {
    hasNextPage: Boolean!     # true si hay más páginas
    hasPreviousPage: Boolean! # true si no es la primera página
    totalCount: Int!          # Total de pagos en la BD
}
```

### Query de Ejemplo
```graphql
query ListPayments {
  listPayments(
    filter: {
      pagination: { page: 1, pageSize: 10 }
      status: PENDING
    }
  ) {
    nodes {
      id
      amount
      status
      payment_date
      payment_method
      member {
        miembro_id
        nombre
        apellidos
      }
    }
    pageInfo {
      hasNextPage       # Deshabilitar botón "Siguiente" si false
      hasPreviousPage   # Deshabilitar botón "Anterior" si false
      totalCount        # Mostrar "Mostrando 1-10 de 150"
    }
  }
}
```

### Cálculo de Páginas Totales
```typescript
const totalPages = Math.ceil(pageInfo.totalCount / pageSize);
const currentPage = 1; // o el que corresponda

// UI hints
const showingFrom = (currentPage - 1) * pageSize + 1;
const showingTo = Math.min(currentPage * pageSize, pageInfo.totalCount);
// "Mostrando 1-10 de 150 pagos"
```

---

## 🔍 Filtros Disponibles

### PaymentFilter
Todos los filtros son **opcionales**:

```graphql
input PaymentFilter {
    # Estado del pago
    status: PaymentStatus              # PENDING, PAID, CANCELLED

    # Método de pago (texto libre)
    payment_method: String             # "Efectivo", "Transferencia", etc.

    # Rango de fechas
    start_date: Time                   # Fecha inicio (ISO 8601)
    end_date: Time                     # Fecha fin (ISO 8601)

    # Rango de montos
    min_amount: Float                  # Monto mínimo
    max_amount: Float                  # Monto máximo

    # Filtrar por miembro específico
    member_id: ID                      # ID del miembro

    # Paginación
    pagination: PaginationInput        # { page: 1, pageSize: 10 }

    # Ordenamiento
    sort: SortInput                    # { field: "payment_date", direction: DESC }
}
```

### Campos de Ordenamiento Comunes
- `payment_date` - Fecha de pago (más común)
- `amount` - Monto del pago
- `created_at` - Fecha de creación del registro
- `status` - Estado del pago

---

## 📊 Estados del Pago

```graphql
enum PaymentStatus {
    PENDING    # Pendiente de pago
    PAID       # Pagado
    CANCELLED  # Cancelado
}
```

### Transiciones de Estado Válidas

```
┌──────────┐
│ PENDING  │ ──────confirmPayment()──────> ┌──────┐
│          │                                │ PAID │
└──────────┘                                └──────┘
     │
     │
     └──────────cancelPayment()──────> ┌────────────┐
                                        │ CANCELLED  │
                                        └────────────┘
```

### ⚠️ Restricciones Importantes
- ✅ Solo se pueden confirmar pagos en estado `PENDING`
- ✅ Solo se pueden cancelar pagos en estado `PENDING`
- ❌ **NO se puede** confirmar un pago `CANCELLED`
- ❌ **NO se puede** cancelar un pago `PAID`

---

## 🔄 Flujo de Trabajo

### 1. Registro Inicial de Pago
```graphql
mutation RegisterPayment {
  registerPayment(input: {
    member_id: "123"
    amount: 50.00
    payment_method: "Pendiente"    # Valor temporal
    notes: "Cuota anual 2025"
  }) {
    id
    status    # PENDING
    amount
    member {
      nombre
      apellidos
    }
  }
}
```

**Comportamiento del Backend:**
- ✅ Se crea el pago con estado `PENDING`
- ✅ Se asocia automáticamente con la cuota anual del año actual
- ✅ Si no existe cuota anual, se usa el importe como cuota base
- ❌ **NO se crea entrada en CashFlow** (solo al confirmar)

---

### 2. Confirmación de Pago
```graphql
mutation ConfirmPayment {
  confirmPayment(
    id: "123"
    paymentMethod: "Transferencia bancaria"
    paymentDate: "2025-11-08T10:30:00Z"
    notes: "Ref: TRANS-2025-001"
  ) {
    id
    status              # PAID
    payment_method      # "Transferencia bancaria"
    payment_date        # "2025-11-08T10:30:00Z"
    notes
  }
}
```

**Comportamiento del Backend:**
- ✅ Cambia estado a `PAID`
- ✅ Actualiza `payment_method`, `payment_date`, y `notes`
- ✅ **Crea automáticamente entrada en CashFlow** tipo `INGRESO_CUOTA`
- ⚠️ Si falla la creación del CashFlow, el pago **SÍ se confirma** (error solo se loguea)

---

### 3. Cancelación de Pago
```graphql
mutation CancelPayment {
  cancelPayment(
    id: "123"
    reason: "Socio se dio de baja antes de pagar"
  ) {
    success
    message
  }
}
```

**Comportamiento del Backend:**
- ✅ Cambia estado a `CANCELLED`
- ✅ Añade la razón a las notas del pago
- ❌ **NO elimina** el registro, solo cambia el estado

---

## 📖 Queries Disponibles

### 1. Listar Todos los Pagos (Admin)
```graphql
query {
  listPayments(filter: {
    status: PENDING
    start_date: "2025-01-01T00:00:00Z"
    end_date: "2025-12-31T23:59:59Z"
    pagination: { page: 1, pageSize: 20 }
    sort: { field: "payment_date", direction: DESC }
  }) {
    nodes { id, amount, status, payment_date }
    pageInfo { totalCount, hasNextPage }
  }
}
```

### 2. Pagos de un Miembro Específico
```graphql
query {
  getMemberPayments(memberId: "123") {
    id
    amount
    status
    payment_date
    payment_method
    membership_fee {
      year
      base_fee_amount
    }
  }
}
```

### 3. Pagos de una Familia
```graphql
query {
  getFamilyPayments(familyId: "456") {
    id
    amount
    status
    member {
      nombre
      apellidos
    }
  }
}
```

### 4. Consultar Estado de un Pago
```graphql
query {
  getPaymentStatus(id: "123")  # Retorna: PENDING | PAID | CANCELLED
}
```

---

## ✏️ Mutations Disponibles

### 1. registerPayment
Crea un nuevo pago en estado `PENDING`.

**Input:**
```graphql
input PaymentInput {
    member_id: ID!          # Requerido
    amount: Float!          # Requerido
    payment_method: String! # Requerido (aunque sea temporal)
    notes: String           # Opcional
}
```

### 2. updatePayment
Actualiza un pago **solo si está en estado PENDING**.

```graphql
mutation {
  updatePayment(
    id: "123"
    input: {
      amount: 55.00
      payment_method: "Efectivo"
      notes: "Monto actualizado"
    }
  ) {
    id
    amount
  }
}
```

### 3. confirmPayment
Confirma un pago y crea entrada en CashFlow.

**Parámetros:**
- `id`: ID del pago (requerido)
- `paymentMethod`: Método de pago real (requerido)
- `paymentDate`: Fecha de pago (opcional, usa fecha actual si no se proporciona)
- `notes`: Notas adicionales (opcional)

### 4. cancelPayment
Cancela un pago pendiente.

**Parámetros:**
- `id`: ID del pago (requerido)
- `reason`: Razón de cancelación (requerido)

### 5. generateAnnualFees
Genera cuotas anuales para todos los miembros activos.

```graphql
mutation {
  generateAnnualFees(input: {
    year: 2025
    base_fee_amount: 50.00
    family_fee_extra: 20.00
  }) {
    year
    membership_fee_id
    payments_generated    # Cuántos pagos se crearon
    payments_existing     # Cuántos ya existían
    total_members         # Total de miembros procesados
    total_expected_amount # Suma total esperada
    details {
      member_id
      member_number
      member_name
      amount
      was_created         # true si se creó, false si ya existía
      error               # mensaje de error si falló
    }
  }
}
```

---

## ⚠️ Comportamientos Importantes

### 1. Integración con CashFlow
- ✅ Al confirmar un pago (`confirmPayment`), se crea automáticamente un registro en `cash_flows`
- ✅ Tipo de operación: `INGRESO_CUOTA`
- ✅ Monto: positivo (ingreso)
- ✅ Vinculación: `payment_id` apunta al pago
- ⚠️ **Si falla la creación del CashFlow, el pago SÍ se confirma** (error solo se loguea)

### 2. Asociación con Cuotas Anuales
- ✅ Al registrar un pago, se asocia automáticamente con la cuota anual del año actual
- ✅ Si no existe cuota anual para el año, se usa el monto del pago como cuota base
- ✅ Familias pagan cuota base + extra familiar

### 3. Validaciones
- ✅ No se pueden crear pagos duplicados de "pago inicial" para el mismo miembro
- ✅ Solo se confirman pagos en estado `PENDING`
- ✅ `payment_method` es requerido al confirmar (no puede estar vacío)
- ✅ Montos deben ser mayores a 0

### 4. Permisos
- 🔒 **ADMIN only**: Todos los queries y mutations de pagos requieren rol `admin`
- 🔒 **USER**: No tiene acceso directo a gestión de pagos

---

## 💡 Ejemplos de Uso en Frontend

### Tabla de Pagos con Paginación
```typescript
const PaymentsTable = () => {
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const { data, loading } = useQuery(LIST_PAYMENTS, {
    variables: {
      filter: {
        pagination: { page, pageSize },
        sort: { field: "payment_date", direction: "DESC" }
      }
    }
  });

  const { nodes: payments, pageInfo } = data?.listPayments || { nodes: [], pageInfo: {} };

  return (
    <div>
      <Table data={payments} />

      <Pagination
        currentPage={page}
        totalCount={pageInfo.totalCount}
        pageSize={pageSize}
        onPageChange={setPage}
        hasNextPage={pageInfo.hasNextPage}
        hasPreviousPage={pageInfo.hasPreviousPage}
      />

      <p>
        Mostrando {(page - 1) * pageSize + 1} -
        {Math.min(page * pageSize, pageInfo.totalCount)} de {pageInfo.totalCount} pagos
      </p>
    </div>
  );
};
```

### Confirmación de Pago
```typescript
const ConfirmPaymentButton = ({ paymentId }) => {
  const [confirmPayment] = useMutation(CONFIRM_PAYMENT);

  const handleConfirm = async () => {
    try {
      const result = await confirmPayment({
        variables: {
          id: paymentId,
          paymentMethod: "Transferencia",
          paymentDate: new Date().toISOString(),
          notes: "Confirmado desde el sistema"
        }
      });

      if (result.data.confirmPayment.status === "PAID") {
        toast.success("Pago confirmado correctamente");
        // Nota: El CashFlow se crea automáticamente en el backend
      }
    } catch (error) {
      toast.error(error.message);
    }
  };

  return <button onClick={handleConfirm}>Confirmar Pago</button>;
};
```

### Filtrado Avanzado
```typescript
const FilteredPayments = () => {
  const [filters, setFilters] = useState({
    status: null,
    startDate: null,
    endDate: null,
    minAmount: null,
    maxAmount: null
  });

  const { data } = useQuery(LIST_PAYMENTS, {
    variables: {
      filter: {
        ...(filters.status && { status: filters.status }),
        ...(filters.startDate && { start_date: filters.startDate }),
        ...(filters.endDate && { end_date: filters.endDate }),
        ...(filters.minAmount && { min_amount: filters.minAmount }),
        ...(filters.maxAmount && { max_amount: filters.maxAmount }),
        pagination: { page: 1, pageSize: 20 }
      }
    }
  });

  return (
    <div>
      <FilterForm onFilterChange={setFilters} />
      <PaymentsList payments={data?.listPayments?.nodes} />
    </div>
  );
};
```

---

## 📝 Notas Finales

1. **Fechas**: Todas las fechas están en formato ISO 8601 (UTC)
2. **IDs**: Los IDs son strings en GraphQL pero uint internamente
3. **Montos**: Los montos son Float (usar 2 decimales en el frontend)
4. **Estados**: Siempre validar transiciones de estado antes de mostrar botones
5. **Errores**: El backend devuelve errores descriptivos con códigos específicos

---

**Última actualización:** 2025-11-08
**Versión del API:** v1

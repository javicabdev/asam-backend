# Guía Frontend - Módulo de Informes (Reports)

## Cambios Importantes - Enero 2025

### Paginación Actualizada en Informe de Morosos

El informe de morosos ahora sigue el mismo patrón de paginación que las demás tablas del sistema (Usuarios, Miembros, Cobros y Flujo de Caja).

**ANTES:**
```graphql
type DelinquentReportResponse {
  debtors: [Debtor!]!
  summary: DelinquentSummary!
  generatedAt: Time!
}
```

**AHORA:**
```graphql
type DelinquentReportResponse {
  debtors: [Debtor!]!          # Lista paginada
  pageInfo: PageInfo!           # NUEVO: Información de paginación
  summary: DelinquentSummary!   # Estadísticas de TODOS los deudores
  generatedAt: Time!
}
```

## 1. Estructura de Paginación

### PageInfo
```typescript
interface PageInfo {
  hasNextPage: boolean      // true si hay más páginas después
  hasPreviousPage: boolean  // true si hay páginas anteriores
  totalCount: number        // Total de deudores (filtrados)
}
```

### PaginationInput
```typescript
interface PaginationInput {
  page: number      // Página actual (default: 1)
  pageSize: number  // Elementos por página (default: 10)
}
```

## 2. Query: getDelinquentReport

### Parámetros de Entrada

```graphql
input DelinquentReportInput {
  cutoffDate: Time           # Fecha de corte (default: hoy)
  minAmount: Float           # Deuda mínima a incluir
  debtorType: String         # "INDIVIDUAL" | "FAMILY" | null (ambos)
  sortBy: String             # Ver opciones de ordenamiento abajo
  pagination: PaginationInput # NUEVO: Paginación
}
```

### Opciones de Ordenamiento (sortBy)
- `"DAYS_DESC"` (default): Más días de atraso primero
- `"DAYS_ASC"`: Menos días de atraso primero
- `"AMOUNT_DESC"`: Mayor deuda primero
- `"AMOUNT_ASC"`: Menor deuda primero
- `"NAME_ASC"`: Orden alfabético

### Ejemplo de Query Completa

```graphql
query GetDelinquentReport(
  $cutoffDate: Time
  $minAmount: Float
  $debtorType: String
  $sortBy: String
  $page: Int!
  $pageSize: Int!
) {
  getDelinquentReport(
    input: {
      cutoffDate: $cutoffDate
      minAmount: $minAmount
      debtorType: $debtorType
      sortBy: $sortBy
      pagination: {
        page: $page
        pageSize: $pageSize
      }
    }
  ) {
    debtors {
      memberId
      familyId
      type
      member {
        id
        memberNumber
        firstName
        lastName
        email
        phone
        status
      }
      family {
        id
        familyName
        primaryMember {
          id
          memberNumber
          firstName
          lastName
          email
          status
        }
        totalMembers
      }
      pendingPayments {
        id
        amount
        createdAt
        daysOverdue
        notes
      }
      totalDebt
      oldestDebtDays
      oldestDebtDate
      lastPaymentDate
      lastPaymentAmount
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
    summary {
      totalDebtors
      individualDebtors
      familyDebtors
      totalDebtAmount
      averageDaysOverdue
      averageDebtPerDebtor
    }
    generatedAt
  }
}
```

## 3. Tipos de Deudor

### INDIVIDUAL
- Deuda asociada a un socio individual
- `member` contendrá los datos del socio
- `family` será `null`

### FAMILY
- Deuda asociada a una familia completa
- `family` contendrá los datos de la familia
- `member` será `null`

## 4. Implementación en React/TypeScript

### Interfaces TypeScript

```typescript
interface DelinquentReportResponse {
  debtors: Debtor[]
  pageInfo: PageInfo
  summary: DelinquentSummary
  generatedAt: string
}

interface Debtor {
  memberId?: string
  familyId?: string
  type: 'INDIVIDUAL' | 'FAMILY'
  member?: DebtorMemberInfo
  family?: DebtorFamilyInfo
  pendingPayments: PendingPayment[]
  totalDebt: number
  oldestDebtDays: number
  oldestDebtDate: string
  lastPaymentDate?: string
  lastPaymentAmount?: number
}

interface DebtorMemberInfo {
  id: string
  memberNumber: string
  firstName: string
  lastName: string
  email?: string
  phone?: string
  status: string
}

interface DebtorFamilyInfo {
  id: string
  familyName: string
  primaryMember: DebtorMemberInfo
  totalMembers: number
}

interface PendingPayment {
  id: string
  amount: number
  createdAt: string
  daysOverdue: number
  notes?: string
}

interface DelinquentSummary {
  totalDebtors: number
  individualDebtors: number
  familyDebtors: number
  totalDebtAmount: number
  averageDaysOverdue: number
  averageDebtPerDebtor: number
}
```

### Hook de Ejemplo con Paginación

```typescript
import { useQuery } from '@apollo/client'
import { useState } from 'react'

const GET_DELINQUENT_REPORT = gql`
  query GetDelinquentReport(
    $cutoffDate: Time
    $minAmount: Float
    $debtorType: String
    $sortBy: String
    $page: Int!
    $pageSize: Int!
  ) {
    getDelinquentReport(
      input: {
        cutoffDate: $cutoffDate
        minAmount: $minAmount
        debtorType: $debtorType
        sortBy: $sortBy
        pagination: { page: $page, pageSize: $pageSize }
      }
    ) {
      debtors {
        # ... campos
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        totalCount
      }
      summary {
        # ... campos
      }
      generatedAt
    }
  }
`

function useDelinquentReport(filters: DelinquentReportFilters) {
  const [page, setPage] = useState(1)
  const pageSize = 10

  const { data, loading, error, refetch } = useQuery(
    GET_DELINQUENT_REPORT,
    {
      variables: {
        ...filters,
        page,
        pageSize,
      },
    }
  )

  const handleNextPage = () => {
    if (data?.getDelinquentReport.pageInfo.hasNextPage) {
      setPage(p => p + 1)
    }
  }

  const handlePreviousPage = () => {
    if (data?.getDelinquentReport.pageInfo.hasPreviousPage) {
      setPage(p => p - 1)
    }
  }

  return {
    debtors: data?.getDelinquentReport.debtors || [],
    pageInfo: data?.getDelinquentReport.pageInfo,
    summary: data?.getDelinquentReport.summary,
    generatedAt: data?.getDelinquentReport.generatedAt,
    loading,
    error,
    page,
    pageSize,
    handleNextPage,
    handlePreviousPage,
    refetch,
  }
}
```

### Componente de Tabla con Paginación

```typescript
function DelinquentReportTable() {
  const [filters, setFilters] = useState({
    minAmount: null,
    debtorType: null,
    sortBy: 'DAYS_DESC',
  })

  const {
    debtors,
    pageInfo,
    summary,
    loading,
    page,
    pageSize,
    handleNextPage,
    handlePreviousPage,
  } = useDelinquentReport(filters)

  if (loading) return <Spinner />

  return (
    <div>
      {/* Resumen Estadístico - Basado en TODOS los deudores */}
      <ReportSummary summary={summary} />

      {/* Filtros */}
      <ReportFilters filters={filters} onChange={setFilters} />

      {/* Tabla de Deudores - Página Actual */}
      <table>
        <thead>
          <tr>
            <th>Tipo</th>
            <th>Nombre</th>
            <th>Deuda Total</th>
            <th>Días Atraso</th>
            <th>Pagos Pendientes</th>
          </tr>
        </thead>
        <tbody>
          {debtors.map(debtor => (
            <tr key={debtor.memberId || debtor.familyId}>
              <td>{debtor.type === 'INDIVIDUAL' ? 'Socio' : 'Familia'}</td>
              <td>
                {debtor.type === 'INDIVIDUAL'
                  ? `${debtor.member?.firstName} ${debtor.member?.lastName}`
                  : debtor.family?.familyName}
              </td>
              <td>{debtor.totalDebt.toFixed(2)} €</td>
              <td>{debtor.oldestDebtDays} días</td>
              <td>{debtor.pendingPayments.length}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Controles de Paginación */}
      {pageInfo && (
        <Pagination
          currentPage={page}
          totalCount={pageInfo.totalCount}
          pageSize={pageSize}
          onNext={handleNextPage}
          onPrevious={handlePreviousPage}
          hasNext={pageInfo.hasNextPage}
          hasPrevious={pageInfo.hasPreviousPage}
        />
      )}
    </div>
  )
}
```

## 5. Características Importantes

### Summary (Resumen)
El campo `summary` contiene estadísticas calculadas sobre **TODOS** los deudores que cumplen los filtros, no solo los de la página actual. Esto permite:
- Mostrar totales globales en el encabezado del reporte
- Calcular promedios correctamente
- Tener una visión completa de la situación financiera

### Filtros Combinables
Puedes combinar múltiples filtros:
```typescript
const filters = {
  minAmount: 100,           // Solo deudas > 100€
  debtorType: 'FAMILY',     // Solo familias
  sortBy: 'AMOUNT_DESC',    // Mayor deuda primero
}
```

### Información de Contacto
- Para socios individuales: `debtor.member` contiene email y teléfono
- Para familias: `debtor.family.primaryMember` contiene los datos del socio principal

### Último Pago
- `lastPaymentDate`: Fecha del último pago exitoso (status PAID)
- `lastPaymentAmount`: Importe del último pago
- Ambos pueden ser `null` si nunca pagó

## 6. Permisos

**IMPORTANTE:** Solo usuarios con rol `ADMIN` pueden acceder a este informe.

```typescript
// El backend retornará error 403 si el usuario no es admin
{
  "errors": [{
    "message": "forbidden: only admin can access delinquent report"
  }]
}
```

## 7. Casos de Uso Comunes

### Reporte Básico (Primera Página)
```typescript
const { data } = useQuery(GET_DELINQUENT_REPORT, {
  variables: {
    page: 1,
    pageSize: 10,
  }
})
```

### Filtrar Deudas Significativas
```typescript
const { data } = useQuery(GET_DELINQUENT_REPORT, {
  variables: {
    minAmount: 500,  // Solo deudas > 500€
    page: 1,
    pageSize: 20,
  }
})
```

### Solo Familias con Mayor Deuda
```typescript
const { data } = useQuery(GET_DELINQUENT_REPORT, {
  variables: {
    debtorType: 'FAMILY',
    sortBy: 'AMOUNT_DESC',
    page: 1,
    pageSize: 10,
  }
})
```

### Exportar Datos Completos
Para exportar el reporte completo a CSV/Excel, solicita todas las páginas:
```typescript
async function exportFullReport() {
  const allDebtors = []
  let page = 1
  let hasMore = true

  while (hasMore) {
    const { data } = await client.query({
      query: GET_DELINQUENT_REPORT,
      variables: { page, pageSize: 100 }
    })

    allDebtors.push(...data.getDelinquentReport.debtors)
    hasMore = data.getDelinquentReport.pageInfo.hasNextPage
    page++
  }

  return exportToCSV(allDebtors)
}
```

## 8. Validaciones y Consideraciones

1. **Paginación por defecto:** Si no se proporciona, usa page=1, pageSize=10
2. **Summary global:** El resumen estadístico se calcula ANTES de paginar, refleja todos los deudores filtrados
3. **Ordenamiento:** Se aplica ANTES de paginar para mantener consistencia
4. **Fechas:** Todos los timestamps están en formato ISO 8601
5. **Importes:** Todos los montos están en euros (€)

## 9. Migración desde Versión Anterior

Si ya tenías implementado el informe de morosos:

### Cambios Necesarios

**1. Actualizar la Query:**
```diff
  query GetDelinquentReport {
    getDelinquentReport(input: {
      sortBy: "DAYS_DESC"
+     pagination: { page: 1, pageSize: 10 }
    }) {
      debtors { ... }
+     pageInfo {
+       hasNextPage
+       hasPreviousPage
+       totalCount
+     }
      summary { ... }
      generatedAt
    }
  }
```

**2. Actualizar el Tipo de Respuesta:**
```diff
  interface DelinquentReportResponse {
    debtors: Debtor[]
+   pageInfo: PageInfo
    summary: DelinquentSummary
    generatedAt: string
  }
```

**3. Añadir Controles de Paginación:**
```typescript
// Antes: Solo mostrabas todos los debtors
{debtors.map(...)}

// Ahora: Añade controles de paginación
<Table data={debtors} />
<Pagination pageInfo={pageInfo} ... />
```

## 10. Testing

### Test de Query Básica
```typescript
it('should fetch first page of delinquent report', async () => {
  const { data } = await client.query({
    query: GET_DELINQUENT_REPORT,
    variables: { page: 1, pageSize: 10 }
  })

  expect(data.getDelinquentReport).toBeDefined()
  expect(data.getDelinquentReport.pageInfo.totalCount).toBeGreaterThanOrEqual(0)
  expect(data.getDelinquentReport.debtors.length).toBeLessThanOrEqual(10)
})
```

### Test de Filtros
```typescript
it('should filter by minimum amount', async () => {
  const { data } = await client.query({
    query: GET_DELINQUENT_REPORT,
    variables: {
      minAmount: 1000,
      page: 1,
      pageSize: 10
    }
  })

  data.getDelinquentReport.debtors.forEach(debtor => {
    expect(debtor.totalDebt).toBeGreaterThanOrEqual(1000)
  })
})
```

---

Para cualquier duda o sugerencia sobre el módulo de informes, contactar al equipo de backend.

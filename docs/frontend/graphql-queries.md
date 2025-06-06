# ASAM Backend - Colección de Queries y Mutations

Este archivo contiene todas las queries y mutations del sistema listas para copiar y usar.

## Autenticación

### Login
```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    user {
      id
      username
      role
      isActive
      lastLogin
    }
    accessToken
    refreshToken
    expiresAt
  }
}
```

### Logout
```graphql
mutation Logout {
  logout {
    success
    message
    error
  }
}
```

### Refresh Token
```graphql
mutation RefreshToken($input: RefreshTokenInput!) {
  refreshToken(input: $input) {
    accessToken
    refreshToken
    expiresAt
  }
}
```

## Queries de Miembros

### Obtener un Miembro
```graphql
query GetMember($id: ID!) {
  getMember(id: $id) {
    miembro_id
    numero_socio
    tipo_membresia
    nombre
    apellidos
    calle_numero_piso
    codigo_postal
    poblacion
    provincia
    pais
    estado
    fecha_alta
    fecha_baja
    fecha_nacimiento
    documento_identidad
    correo_electronico
    profesion
    nacionalidad
    observaciones
  }
}
```

### Listar Miembros
```graphql
query ListMembers($filter: MemberFilter) {
  listMembers(filter: $filter) {
    nodes {
      miembro_id
      numero_socio
      nombre
      apellidos
      estado
      tipo_membresia
      fecha_alta
      correo_electronico
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
  }
}
```

### Buscar Miembros
```graphql
query SearchMembers($criteria: String!) {
  searchMembers(criteria: $criteria) {
    miembro_id
    numero_socio
    nombre
    apellidos
    estado
    tipo_membresia
    correo_electronico
  }
}
```

## Mutations de Miembros

### Crear Miembro
```graphql
mutation CreateMember($input: CreateMemberInput!) {
  createMember(input: $input) {
    miembro_id
    numero_socio
    tipo_membresia
    nombre
    apellidos
    estado
    fecha_alta
    correo_electronico
  }
}
```

### Actualizar Miembro
```graphql
mutation UpdateMember($input: UpdateMemberInput!) {
  updateMember(input: $input) {
    miembro_id
    nombre
    apellidos
    correo_electronico
    calle_numero_piso
    codigo_postal
    poblacion
    provincia
    pais
    documento_identidad
    profesion
    observaciones
  }
}
```

### Cambiar Estado de Miembro
```graphql
mutation ChangeMemberStatus($id: ID!, $status: MemberStatus!) {
  changeMemberStatus(id: $id, status: $status) {
    miembro_id
    nombre
    apellidos
    estado
  }
}
```

### Eliminar Miembro
```graphql
mutation DeleteMember($id: ID!) {
  deleteMember(id: $id) {
    success
    message
    error
  }
}
```

## Queries de Familias

### Obtener una Familia
```graphql
query GetFamily($id: ID!) {
  getFamily(id: $id) {
    id
    numero_socio
    esposo_nombre
    esposo_apellidos
    esposa_nombre
    esposa_apellidos
    miembro_origen {
      miembro_id
      nombre
      apellidos
    }
    familiares {
      id
      nombre
      apellidos
      fecha_nacimiento
      dni_nie
      correo_electronico
    }
  }
}
```

### Listar Familias
```graphql
query ListFamilies($filter: FamilyFilter) {
  listFamilies(filter: $filter) {
    nodes {
      id
      numero_socio
      esposo_nombre
      esposo_apellidos
      esposa_nombre
      esposa_apellidos
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
  }
}
```

### Obtener Miembros de Familia
```graphql
query GetFamilyMembers($familyId: ID!) {
  getFamilyMembers(familyId: $familyId) {
    id
    nombre
    apellidos
    fecha_nacimiento
    dni_nie
    correo_electronico
  }
}
```

## Mutations de Familias

### Crear Familia
```graphql
mutation CreateFamily($input: CreateFamilyInput!) {
  createFamily(input: $input) {
    id
    numero_socio
    esposo_nombre
    esposo_apellidos
    esposa_nombre
    esposa_apellidos
    miembro_origen {
      miembro_id
      nombre
      apellidos
    }
  }
}
```

### Actualizar Familia
```graphql
mutation UpdateFamily($input: UpdateFamilyInput!) {
  updateFamily(input: $input) {
    id
    esposo_nombre
    esposo_apellidos
    esposo_documento_identidad
    esposo_correo_electronico
    esposa_nombre
    esposa_apellidos
    esposa_documento_identidad
    esposa_correo_electronico
  }
}
```

### Añadir Miembro a Familia
```graphql
mutation AddFamilyMember($family_id: ID!, $familiar: FamiliarInput!) {
  addFamilyMember(family_id: $family_id, familiar: $familiar) {
    id
    familiares {
      id
      nombre
      apellidos
      fecha_nacimiento
      dni_nie
      correo_electronico
    }
  }
}
```

### Eliminar Miembro de Familia
```graphql
mutation RemoveFamilyMember($familiar_id: ID!) {
  removeFamilyMember(familiar_id: $familiar_id) {
    success
    message
    error
  }
}
```

## Queries de Pagos

### Obtener un Pago
```graphql
query GetPayment($id: ID!) {
  getPayment(id: $id) {
    id
    member {
      miembro_id
      nombre
      apellidos
    }
    family {
      id
      numero_socio
    }
    amount
    payment_date
    status
    payment_method
    notes
  }
}
```

### Obtener Pagos de Miembro
```graphql
query GetMemberPayments($memberId: ID!) {
  getMemberPayments(memberId: $memberId) {
    id
    amount
    payment_date
    status
    payment_method
    notes
  }
}
```

### Obtener Pagos de Familia
```graphql
query GetFamilyPayments($familyId: ID!) {
  getFamilyPayments(familyId: $familyId) {
    id
    amount
    payment_date
    status
    payment_method
    notes
  }
}
```

### Obtener Estado de Pago
```graphql
query GetPaymentStatus($id: ID!) {
  getPaymentStatus(id: $id)
}
```

## Mutations de Pagos

### Registrar Pago
```graphql
mutation RegisterPayment($input: PaymentInput!) {
  registerPayment(input: $input) {
    id
    amount
    payment_date
    status
    payment_method
    notes
    member {
      miembro_id
      nombre
      apellidos
    }
    family {
      id
      numero_socio
    }
  }
}
```

### Actualizar Pago
```graphql
mutation UpdatePayment($id: ID!, $input: PaymentInput!) {
  updatePayment(id: $id, input: $input) {
    id
    amount
    payment_date
    status
    payment_method
    notes
  }
}
```

### Cancelar Pago
```graphql
mutation CancelPayment($id: ID!, $reason: String!) {
  cancelPayment(id: $id, reason: $reason) {
    success
    message
    error
  }
}
```

### Registrar Cuotas Masivas
```graphql
mutation RegisterFee($year: Int!, $month: Int!, $base_amount: Float!) {
  registerFee(year: $year, month: $month, base_amount: $base_amount) {
    success
    message
    error
  }
}
```

## Queries de Flujo de Caja

### Obtener una Transacción
```graphql
query GetCashFlow($id: ID!) {
  getCashFlow(id: $id) {
    id
    amount
    date
    operation_type
    detail
    member {
      miembro_id
      nombre
      apellidos
    }
    family {
      id
      numero_socio
    }
    payment {
      id
      status
    }
  }
}
```

### Obtener Balance
```graphql
query GetBalance {
  getBalance
}
```

### Listar Transacciones
```graphql
query GetTransactions($filter: TransactionFilter) {
  getTransactions(filter: $filter) {
    nodes {
      id
      amount
      date
      operation_type
      detail
      member {
        miembro_id
        nombre
        apellidos
      }
      family {
        id
        numero_socio
      }
      payment {
        id
        status
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
  }
}
```

## Mutations de Flujo de Caja

### Registrar Transacción
```graphql
mutation RegisterTransaction($input: TransactionInput!) {
  registerTransaction(input: $input) {
    id
    amount
    date
    operation_type
    detail
  }
}
```

### Actualizar Transacción
```graphql
mutation UpdateTransaction($id: ID!, $input: TransactionInput!) {
  updateTransaction(id: $id, input: $input) {
    id
    amount
    date
    operation_type
    detail
  }
}
```

### Ajustar Balance
```graphql
mutation AdjustBalance($amount: Float!, $reason: String!) {
  adjustBalance(amount: $amount, reason: $reason) {
    success
    message
    error
  }
}
```

## Ejemplos de Variables

### Login
```json
{
  "input": {
    "username": "usuario@ejemplo.com",
    "password": "contraseña123"
  }
}
```

### Crear Miembro
```json
{
  "input": {
    "numero_socio": "2023-001",
    "tipo_membresia": "INDIVIDUAL",
    "nombre": "Juan",
    "apellidos": "Pérez García",
    "calle_numero_piso": "Calle Principal 123, 2º A",
    "codigo_postal": "07001",
    "poblacion": "Palma de Mallorca",
    "provincia": "Islas Baleares",
    "pais": "España",
    "fecha_nacimiento": "1980-01-15T00:00:00Z",
    "documento_identidad": "12345678X",
    "correo_electronico": "juan.perez@ejemplo.com",
    "profesion": "Ingeniero",
    "nacionalidad": "Española",
    "observaciones": "Miembro fundador"
  }
}
```

### Filtro de Miembros
```json
{
  "filter": {
    "estado": "ACTIVE",
    "tipo_membresia": "INDIVIDUAL",
    "search_term": "Juan",
    "pagination": {
      "page": 1,
      "pageSize": 20
    },
    "sort": {
      "field": "NOMBRE",
      "direction": "ASC"
    }
  }
}
```

### Registrar Pago
```json
{
  "input": {
    "member_id": "1",
    "amount": 50.00,
    "payment_method": "TRANSFERENCIA",
    "notes": "Cuota mensual enero 2024"
  }
}
```

### Filtro de Transacciones
```json
{
  "filter": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z",
    "operation_type": "MEMBERSHIP_FEE",
    "pagination": {
      "page": 1,
      "pageSize": 50
    },
    "sort": {
      "field": "DATE",
      "direction": "DESC"
    }
  }
}
```

### Crear Familia
```json
{
  "input": {
    "numero_socio": "2024-F001",
    "miembro_origen_id": "1",
    "esposo_nombre": "Juan",
    "esposo_apellidos": "García López",
    "esposo_fecha_nacimiento": "1975-05-20T00:00:00Z",
    "esposo_documento_identidad": "12345678X",
    "esposo_correo_electronico": "juan.garcia@ejemplo.com",
    "esposa_nombre": "María",
    "esposa_apellidos": "Martínez Sánchez",
    "esposa_fecha_nacimiento": "1978-10-15T00:00:00Z",
    "esposa_documento_identidad": "87654321Y",
    "esposa_correo_electronico": "maria.martinez@ejemplo.com"
  }
}
```

### Añadir Familiar
```json
{
  "family_id": "1",
  "familiar": {
    "nombre": "Ana",
    "apellidos": "García Martínez",
    "fecha_nacimiento": "2005-03-10T00:00:00Z",
    "dni_nie": "12345678A",
    "correo_electronico": "ana.garcia@ejemplo.com",
    "parentesco": "Hija"
  }
}
```

# Guía para Desarrolladores Frontend

## Introducción

Esta guía proporciona información detallada para desarrolladores frontend que necesitan interactuar con la API GraphQL del backend de ASAM. La API está diseñada para proporcionar todas las funcionalidades necesarias para la gestión de miembros, familias, pagos y transacciones de la asociación.

## Información General

- **Endpoint de Producción**: `https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql`
- **Endpoint de Desarrollo**: `http://localhost:8080/graphql`
- **Playground**: `http://localhost:8080/playground` (disponible solo en entorno de desarrollo)
- **Autenticación**: Bearer Token JWT

### Configuración Rápida

```javascript
// Apollo Client
const GRAPHQL_ENDPOINT = process.env.NODE_ENV === 'production' 
  ? 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
  : 'http://localhost:8080/graphql';

const client = new ApolloClient({
  uri: GRAPHQL_ENDPOINT,
  headers: {
    authorization: localStorage.getItem('token') || '',
  },
});
```

## Autenticación

### Iniciar Sesión

```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    user {
      id
      username
      role
      isActive
      lastLogin
      emailVerified
      emailVerifiedAt
    }
    accessToken
    refreshToken
    expiresAt
  }
}

# Variables
{
  "input": {
    "username": "tu_usuario",
    "password": "tu_contraseña"
  }
}
```

### Renovar Token

```graphql
mutation RefreshToken($input: RefreshTokenInput!) {
  refreshToken(input: $input) {
    accessToken
    refreshToken
    expiresAt
  }
}

# Variables
{
  "input": {
    "refreshToken": "tu_refresh_token"
  }
}
```

### Cerrar Sesión

```graphql
mutation Logout {
  logout {
    success
    message
    error
  }
}
```

## Gestión de Usuarios

### Obtener Usuario Actual

```graphql
query GetCurrentUser {
  getCurrentUser {
    id
    username
    role
    isActive
    lastLogin
    emailVerified
    emailVerifiedAt
  }
}
```

### Obtener un Usuario (Admin)

```graphql
query GetUser($id: ID!) {
  getUser(id: $id) {
    id
    username
    role
    isActive
    lastLogin
    emailVerified
    emailVerifiedAt
  }
}
```

### Listar Usuarios (Admin)

```graphql
query ListUsers($page: Int, $pageSize: Int) {
  listUsers(page: $page, pageSize: $pageSize) {
    id
    username
    role
    isActive
    lastLogin
    emailVerified
  }
}
```

### Crear Usuario (Admin)

```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    username
    role
    isActive
  }
}

# Variables
{
  "input": {
    "username": "nuevo.usuario@ejemplo.com",
    "password": "ContraseñaSegura123!",
    "role": "user"
  }
}
```

### Actualizar Usuario (Admin)

```graphql
mutation UpdateUser($input: UpdateUserInput!) {
  updateUser(input: $input) {
    id
    username
    role
    isActive
  }
}

# Variables
{
  "input": {
    "id": "1",
    "username": "usuario.actualizado@ejemplo.com",
    "role": "admin",
    "isActive": true
  }
}
```

### Eliminar Usuario (Admin)

```graphql
mutation DeleteUser($id: ID!) {
  deleteUser(id: $id) {
    success
    message
    error
  }
}
```

### Cambiar Contraseña

```graphql
mutation ChangePassword($input: ChangePasswordInput!) {
  changePassword(input: $input) {
    success
    message
    error
  }
}

# Variables
{
  "input": {
    "currentPassword": "contraseñaActual123",
    "newPassword": "nuevaContraseñaSegura456!"
  }
}
```

### Resetear Contraseña de Usuario (Admin)

```graphql
mutation ResetUserPassword($userId: ID!, $newPassword: String!) {
  resetUserPassword(userId: $userId, newPassword: $newPassword) {
    success
    message
    error
  }
}
```

## Verificación de Email

### Enviar Email de Verificación

```graphql
mutation SendVerificationEmail {
  sendVerificationEmail {
    success
    message
    error
  }
}
```

### Verificar Email

```graphql
mutation VerifyEmail($token: String!) {
  verifyEmail(token: $token) {
    success
    message
    error
  }
}
```

### Reenviar Email de Verificación

```graphql
mutation ResendVerificationEmail($email: String!) {
  resendVerificationEmail(email: $email) {
    success
    message
    error
  }
}
```

## Recuperación de Contraseña

### Solicitar Reseteo de Contraseña

```graphql
mutation RequestPasswordReset($email: String!) {
  requestPasswordReset(email: $email) {
    success
    message
    error
  }
}
```

### Resetear Contraseña con Token

```graphql
mutation ResetPasswordWithToken($token: String!, $newPassword: String!) {
  resetPasswordWithToken(token: $token, newPassword: $newPassword) {
    success
    message
    error
  }
}
```

## Gestión de Miembros

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

# Variables
{
  "id": "1"
}
```

### Listar Miembros con Paginación y Filtros

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
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
  }
}

# Variables (ejemplo con filtrado y paginación)
{
  "filter": {
    "estado": "ACTIVE",
    "tipo_membresia": "INDIVIDUAL",
    "search_term": "Juan",
    "pagination": {
      "page": 1,
      "pageSize": 10
    },
    "sort": {
      "field": "NOMBRE",
      "direction": "ASC"
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
  }
}

# Variables
{
  "criteria": "Juan"
}
```

### Crear un Miembro

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
  }
}

# Variables
{
  "input": {
    "numero_socio": "2025-001",
    "tipo_membresia": "INDIVIDUAL",
    "nombre": "Juan",
    "apellidos": "Pérez García",
    "calle_numero_piso": "Calle Principal 123, 2º",
    "codigo_postal": "08001",
    "poblacion": "Barcelona",
    "provincia": "Barcelona",
    "pais": "España",
    "fecha_nacimiento": "1980-01-15T00:00:00Z",
    "documento_identidad": "12345678X",
    "correo_electronico": "juan.perez@ejemplo.com",
    "profesion": "Ingeniero",
    "nacionalidad": "Española",
    "observaciones": "Miembro recomendado por María García"
  }
}
```

### Actualizar un Miembro

```graphql
mutation UpdateMember($input: UpdateMemberInput!) {
  updateMember(input: $input) {
    miembro_id
    nombre
    apellidos
    correo_electronico
    calle_numero_piso
  }
}

# Variables
{
  "input": {
    "miembro_id": "1",
    "calle_numero_piso": "Nueva Calle 456, 3º",
    "codigo_postal": "08002",
    "correo_electronico": "nuevo.email@ejemplo.com"
  }
}
```

### Cambiar Estado de un Miembro

```graphql
mutation ChangeMemberStatus($id: ID!, $status: MemberStatus!) {
  changeMemberStatus(id: $id, status: $status) {
    miembro_id
    nombre
    apellidos
    estado
  }
}

# Variables
{
  "id": "1",
  "status": "INACTIVE"
}
```

### Eliminar un Miembro

```graphql
mutation DeleteMember($id: ID!) {
  deleteMember(id: $id) {
    success
    message
    error
  }
}

# Variables
{
  "id": "1"
}
```

## Gestión de Familias

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
    }
  }
}

# Variables
{
  "id": "1"
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

# Variables
{
  "filter": {
    "search_term": "García",
    "pagination": {
      "page": 1,
      "pageSize": 10
    },
    "sort": {
      "field": "ESPOSO_NOMBRE",
      "direction": "ASC"
    }
  }
}
```

### Obtener Miembros de una Familia

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

# Variables
{
  "familyId": "1"
}
```

### Crear una Familia

```graphql
mutation CreateFamily($input: CreateFamilyInput!) {
  createFamily(input: $input) {
    id
    numero_socio
    esposo_nombre
    esposo_apellidos
    esposa_nombre
    esposa_apellidos
  }
}

# Variables
{
  "input": {
    "numero_socio": "2025-F001",
    "miembro_origen_id": "1",
    "esposo_nombre": "Juan",
    "esposo_apellidos": "García López",
    "esposa_nombre": "María",
    "esposa_apellidos": "Martínez Sánchez",
    "esposo_fecha_nacimiento": "1975-05-20T00:00:00Z",
    "esposo_documento_identidad": "12345678X",
    "esposo_correo_electronico": "juan.garcia@ejemplo.com",
    "esposa_fecha_nacimiento": "1978-10-15T00:00:00Z",
    "esposa_documento_identidad": "87654321Y",
    "esposa_correo_electronico": "maria.martinez@ejemplo.com"
  }
}
```

### Actualizar una Familia

```graphql
mutation UpdateFamily($input: UpdateFamilyInput!) {
  updateFamily(input: $input) {
    id
    esposo_nombre
    esposo_apellidos
    esposa_nombre
    esposa_apellidos
  }
}

# Variables
{
  "input": {
    "familia_id": "1",
    "esposo_correo_electronico": "nuevo.email@ejemplo.com",
    "esposa_documento_identidad": "98765432Z"
  }
}
```

### Añadir Miembro a una Familia

```graphql
mutation AddFamilyMember($familyId: ID!, $familiar: FamiliarInput!) {
  addFamilyMember(family_id: $familyId, familiar: $familiar) {
    id
    familiares {
      id
      nombre
      apellidos
      fecha_nacimiento
    }
  }
}

# Variables
{
  "familyId": "1",
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

### Eliminar Miembro de una Familia

```graphql
mutation RemoveFamilyMember($familiarId: ID!) {
  removeFamilyMember(familiar_id: $familiarId) {
    success
    message
    error
  }
}

# Variables
{
  "familiarId": "1"
}
```

## Gestión de Pagos

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

# Variables
{
  "id": "1"
}
```

### Obtener Pagos de un Miembro

```graphql
query GetMemberPayments($memberId: ID!) {
  getMemberPayments(memberId: $memberId) {
    id
    amount
    payment_date
    status
    payment_method
  }
}

# Variables
{
  "memberId": "1"
}
```

### Obtener Estado de un Pago

```graphql
query GetPaymentStatus($id: ID!) {
  getPaymentStatus(id: $id)
}

# Variables
{
  "id": "1"
}
```

### Registrar un Pago

```graphql
mutation RegisterPayment($input: PaymentInput!) {
  registerPayment(input: $input) {
    id
    amount
    payment_date
    status
    payment_method
    notes
  }
}

# Variables
{
  "input": {
    "member_id": "1",
    "amount": 50.0,
    "payment_method": "TRANSFERENCIA",
    "notes": "Pago de cuota mensual"
  }
}
```

### Generar Cuotas Anuales

```graphql
mutation GenerateAnnualFees($input: GenerateAnnualFeesInput!) {
  generateAnnualFees(input: $input) {
    year
    membership_fee_id
    payments_generated
    payments_existing
    total_members
    total_expected_amount
    details {
      member_id
      member_number
      member_name
      amount
      was_created
      error
    }
  }
}

# Variables
{
  "input": {
    "year": 2025,
    "base_fee_amount": 100.0,
    "family_fee_extra": 50.0
  }
}
```

## Gestión de Transacciones (Flujo de Caja)

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

# Variables
{
  "filter": {
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-12-31T23:59:59Z",
    "operation_type": "MEMBERSHIP_FEE",
    "pagination": {
      "page": 1,
      "pageSize": 20
    },
    "sort": {
      "field": "DATE",
      "direction": "DESC"
    }
  }
}
```

### Registrar una Transacción

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

# Variables
{
  "input": {
    "operation_type": "OTHER_INCOME",
    "amount": 500.0,
    "date": "2025-06-15T14:30:00Z",
    "detail": "Donación anónima"
  }
}
```

## Tipos de Datos

### User (Usuario del Sistema)

```graphql
type User {
  id: ID!
  username: String!
  role: UserRole!
  isActive: Boolean!
  lastLogin: Time
  emailVerified: Boolean!
  emailVerifiedAt: Time
}
```

- `id`: Identificador único del usuario
- `username`: Nombre de usuario (generalmente el email)
- `role`: Rol del usuario (admin o user)
- `isActive`: Indica si el usuario está activo
- `lastLogin`: Fecha y hora del último inicio de sesión
- `emailVerified`: Indica si el email ha sido verificado
- `emailVerifiedAt`: Fecha y hora de verificación del email

## Enumeraciones (Enums)

### MembershipType (Tipo de Membresía)
- `INDIVIDUAL`: Miembro individual
- `FAMILY`: Miembro familiar

### MemberStatus (Estado del Miembro)
- `ACTIVE`: Miembro activo
- `INACTIVE`: Miembro inactivo

### OperationType (Tipo de Operación)
- `MEMBERSHIP_FEE`: Cuota de membresía
- `CURRENT_EXPENSE`: Gasto corriente
- `FUND_DELIVERY`: Entrega de fondos
- `OTHER_INCOME`: Otros ingresos

### PaymentStatus (Estado del Pago)
- `PENDING`: Pendiente
- `PAID`: Pagado
- `CANCELLED`: Cancelado

### SortDirection (Dirección de Ordenación)
- `ASC`: Ascendente
- `DESC`: Descendente

### UserRole (Rol de Usuario)
- `admin`: Administrador con acceso completo
- `user`: Usuario regular con acceso limitado

## Filtros y Paginación

### PaginationInput
```json
{
  "page": 1,       // Número de página (empieza en 1)
  "pageSize": 10   // Elementos por página
}
```

### SortInput
```json
{
  "field": "NOMBRE",     // Campo a ordenar
  "direction": "ASC"     // Dirección (ASC o DESC)
}
```

### MemberFilter
```json
{
  "estado": "ACTIVE",            // Estado (ACTIVE o INACTIVE)
  "tipo_membresia": "INDIVIDUAL", // Tipo (INDIVIDUAL o FAMILY)
  "search_term": "Juan",         // Término de búsqueda
  "pagination": { ... },         // Objeto PaginationInput
  "sort": { ... }                // Objeto SortInput
}
```

### FamilyFilter
```json
{
  "search_term": "García",       // Término de búsqueda
  "pagination": { ... },         // Objeto PaginationInput
  "sort": { ... }                // Objeto SortInput
}
```

### TransactionFilter
```json
{
  "start_date": "2023-01-01T00:00:00Z",    // Fecha de inicio
  "end_date": "2023-12-31T23:59:59Z",      // Fecha de fin
  "operation_type": "MEMBERSHIP_FEE",      // Tipo de operación
  "pagination": { ... },                   // Objeto PaginationInput
  "sort": { ... }                          // Objeto SortInput
}
```

## Manejo de Errores

La API devuelve errores en el siguiente formato:

```json
{
  "errors": [
    {
      "message": "Mensaje de error",
      "path": ["ruta", "de", "la", "operación"],
      "extensions": {
        "code": "CÓDIGO_ERROR",
        "field": "campo_con_error",
        "details": {
          "campo1": "mensaje de error para campo1",
          "campo2": "mensaje de error para campo2"
        }
      }
    }
  ],
  "data": null
}
```

### Códigos de Error Comunes

- `UNAUTHORIZED`: Error de autenticación
- `FORBIDDEN`: Error de permisos
- `NOT_FOUND`: Recurso no encontrado
- `VALIDATION_ERROR`: Error de validación
- `INTERNAL_ERROR`: Error interno del servidor
- `BUSINESS_ERROR`: Error de lógica de negocio

## Consideraciones para Desarrollo Frontend

1. **Endpoints**:
   - **Producción**: `https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql`
   - **Desarrollo**: `http://localhost:8080/graphql`
   - Configurar variables de entorno para cambiar entre entornos

2. **Autenticación**:
   - Guardar tokens (access y refresh) en almacenamiento seguro
   - Incluir el token de acceso en cada petición autenticada
   - Implementar renovación automática cuando expire el token
   - Verificar si el email del usuario está verificado después del login
   - Mostrar avisos para usuarios con email no verificado

3. **Gestión de Usuarios**:
   - Solo los administradores pueden crear, actualizar y eliminar usuarios
   - Los usuarios regulares solo pueden cambiar su propia contraseña
   - Implementar flujo completo de verificación de email
   - Incluir recuperación de contraseña en la página de login

4. **Optimización de Queries**:
   - Solicitar solo los campos necesarios para cada vista
   - Utilizar paginación para listas grandes
   - Implementar búsqueda y filtrado en el cliente
   - Cachear resultados con Apollo Client

5. **Manejo de Fechas**:
   - Todas las fechas se envían y reciben en formato ISO 8601
   - Considerar zonas horarias para presentación al usuario

6. **Validación de Formularios**:
   - Implementar validación en el cliente antes de enviar al servidor
   - Manejar errores de validación del servidor apropiadamente
   - Validación de contraseñas: mínimo 8 caracteres, mayúsculas, minúsculas, números y caracteres especiales

7. **Seguridad**:
   - No almacenar información sensible en almacenamiento no seguro
   - Implementar cierre de sesión automático por inactividad
   - Validar permisos en el cliente basado en rol de usuario
   - Usar HTTPS siempre en producción

## Ejemplos de Uso en Frameworks Frontend

### Ejemplo con React y Apollo Client

```jsx
import { gql, useMutation } from '@apollo/client';
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

// Configuración del cliente Apollo
const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL || 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
});

const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('accessToken');
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  }
});

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache()
});

// Mutation de login
const LOGIN_MUTATION = gql`
  mutation Login($username: String!, $password: String!) {
    login(input: { username: $username, password: $password }) {
      accessToken
      refreshToken
      user {
        id
        username
        role
        emailVerified
      }
    }
  }
`;

function LoginForm() {
  const [login, { data, loading, error }] = useMutation(LOGIN_MUTATION);
  
  const handleSubmit = (e) => {
    e.preventDefault();
    const username = e.target.username.value;
    const password = e.target.password.value;
    
    login({ variables: { username, password } })
      .then(({ data }) => {
        // Guardar tokens en localStorage o mejor en un almacenamiento seguro
        localStorage.setItem('accessToken', data.login.accessToken);
        localStorage.setItem('refreshToken', data.login.refreshToken);
        
        // Redireccionar o actualizar estado
      })
      .catch(err => {
        // Manejar error de login
        console.error('Error de login:', err);
      });
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {error && <p className="error">Error: {error.message}</p>}
      <div>
        <label>Usuario:</label>
        <input type="text" name="username" required />
      </div>
      <div>
        <label>Contraseña:</label>
        <input type="password" name="password" required />
      </div>
      <button type="submit" disabled={loading}>
        {loading ? 'Iniciando sesión...' : 'Iniciar sesión'}
      </button>
    </form>
  );
}
```

### Ejemplo con Vue y GraphQL

```vue
<template>
  <div>
    <h2>Lista de Miembros</h2>
    <div class="filters">
      <input v-model="searchTerm" placeholder="Buscar..." @input="updateFilters" />
      <select v-model="memberStatus" @change="updateFilters">
        <option value="">Todos</option>
        <option value="ACTIVE">Activos</option>
        <option value="INACTIVE">Inactivos</option>
      </select>
    </div>
    
    <table v-if="members.length">
      <thead>
        <tr>
          <th>Número Socio</th>
          <th>Nombre</th>
          <th>Apellidos</th>
          <th>Estado</th>
          <th>Acciones</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="member in members" :key="member.miembro_id">
          <td>{{ member.numero_socio }}</td>
          <td>{{ member.nombre }}</td>
          <td>{{ member.apellidos }}</td>
          <td>{{ member.estado === 'ACTIVE' ? 'Activo' : 'Inactivo' }}</td>
          <td>
            <button @click="viewMember(member.miembro_id)">Ver</button>
            <button @click="editMember(member.miembro_id)">Editar</button>
          </td>
        </tr>
      </tbody>
    </table>
    
    <div class="pagination">
      <button :disabled="page === 1" @click="prevPage">Anterior</button>
      <span>Página {{ page }} de {{ totalPages }}</span>
      <button :disabled="!hasNextPage" @click="nextPage">Siguiente</button>
    </div>
    
    <div v-if="loading" class="loading">Cargando...</div>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import gql from 'graphql-tag';

const LIST_MEMBERS_QUERY = gql`
  query ListMembers($filter: MemberFilter) {
    listMembers(filter: $filter) {
      nodes {
        miembro_id
        numero_socio
        nombre
        apellidos
        estado
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        totalCount
      }
    }
  }
`;

export default {
  setup() {
    const page = ref(1);
    const pageSize = ref(10);
    const searchTerm = ref('');
    const memberStatus = ref('');
    const totalCount = ref(0);
    const hasNextPage = ref(false);
    
    const filter = reactive({
      pagination: { page: page.value, pageSize: pageSize.value },
      search_term: searchTerm.value,
      estado: memberStatus.value || null,
      sort: { field: "NOMBRE", direction: "ASC" }
    });
    
    const { result, loading, error, refetch } = useQuery(
      LIST_MEMBERS_QUERY,
      { filter },
      { fetchPolicy: 'network-only' }
    );
    
    const members = ref([]);
    
    const updateFilters = () => {
      filter.search_term = searchTerm.value;
      filter.estado = memberStatus.value || null;
      filter.pagination.page = 1;
      page.value = 1;
      refetch();
    };
    
    const prevPage = () => {
      page.value--;
      filter.pagination.page = page.value;
      refetch();
    };
    
    const nextPage = () => {
      page.value++;
      filter.pagination.page = page.value;
      refetch();
    };
    
    const viewMember = (id) => {
      // Implementar navegación a página de detalle
      console.log('Ver miembro:', id);
    };
    
    const editMember = (id) => {
      // Implementar navegación a página de edición
      console.log('Editar miembro:', id);
    };
    
    // Actualizar datos cuando cambia el resultado
    watch(result, () => {
      if (result.value && result.value.listMembers) {
        members.value = result.value.listMembers.nodes;
        hasNextPage.value = result.value.listMembers.pageInfo.hasNextPage;
        totalCount.value = result.value.listMembers.pageInfo.totalCount;
      }
    });
    
    onMounted(() => {
      refetch();
    });
    
    const totalPages = computed(() => Math.ceil(totalCount.value / pageSize.value));
    
    return {
      members,
      loading,
      error,
      page,
      searchTerm,
      memberStatus,
      hasNextPage,
      totalCount,
      updateFilters,
      prevPage,
      nextPage,
      viewMember,
      editMember,
      totalPages
    };
  }
};
</script>
```

## Soporte y Contacto

Para cualquier duda sobre la API o para reportar errores, por favor contactar al equipo de desarrollo.

## Documentación Completa

Para una documentación más detallada y específica para el desarrollo frontend, consulta:

- 📚 [Documentación Frontend Completa](./frontend/README.md)
- 🎉 [Actualizaciones de Junio 2025](./frontend/updates-june-2025.md)
- 🚀 [Guía de Inicio Rápido](./frontend/quick-reference.md)
- 📝 [Colección de Queries y Mutations](./frontend/graphql-queries.md)
- ⚛️ [Integración con React](./frontend/react-integration-guide.md)
- 🕹️ [Integración con Vue](./frontend/vue-integration-guide.md)
- ⚠️ [Manejo de Errores](./frontend/error-handling-guide.md)
- 🔒 [Guía de Seguridad](./frontend/security-guide.md)

### Endpoint de Producción

```
https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql
```

## Compatibilidad con Apollo Client

El backend maneja automáticamente el campo `__typename` que Apollo Client añade a todas las consultas y mutaciones. No es necesario realizar ninguna configuración especial en el frontend ni desactivar `__typename` en Apollo Client.

Para más detalles, consulta la [documentación de compatibilidad con Apollo Client](../apollo-client-compatibility.md).

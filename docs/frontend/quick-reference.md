# ASAM Backend - Referencia Rápida para Frontend

## 🚀 Inicio Rápido

### Endpoint Principal
```
Desarrollo: http://localhost:8080/graphql
Producción: https://api.asam.org/graphql
Playground: http://localhost:8080/playground (solo dev)
```

### Headers Requeridos
```http
Content-Type: application/json
Authorization: Bearer <access_token>
```

## 🔐 Autenticación

### Login
```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    user { id username role }
    accessToken
    refreshToken
    expiresAt
  }
}

# Variables
{ "input": { "username": "email@ejemplo.com", "password": "contraseña" } }
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
```

### Tokens
- **Access Token**: 15 minutos
- **Refresh Token**: 7 días

## 📊 Queries Principales

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
    }
    pageInfo {
      hasNextPage
      totalCount
    }
  }
}

# Filtros disponibles
{
  "filter": {
    "estado": "ACTIVE",
    "tipo_membresia": "INDIVIDUAL",
    "search_term": "Juan",
    "pagination": { "page": 1, "pageSize": 20 },
    "sort": { "field": "NOMBRE", "direction": "ASC" }
  }
}
```

### Obtener Balance
```graphql
query GetBalance {
  getBalance
}
```

## ✏️ Mutations Principales

### Crear Miembro
```graphql
mutation CreateMember($input: CreateMemberInput!) {
  createMember(input: $input) {
    miembro_id
    numero_socio
  }
}
```

### Registrar Pago
```graphql
mutation RegisterPayment($input: PaymentInput!) {
  registerPayment(input: $input) {
    id
    amount
    status
  }
}
```

## 🔢 Enums Importantes

```graphql
enum MemberStatus {
  ACTIVE
  INACTIVE
}

enum MembershipType {
  INDIVIDUAL
  FAMILY
}

enum PaymentStatus {
  PENDING
  PAID
  CANCELLED
}

enum UserRole {
  ADMIN
  USER
}

enum OperationType {
  MEMBERSHIP_FEE
  CURRENT_EXPENSE
  FUND_DELIVERY
  OTHER_INCOME
}
```

## ❌ Códigos de Error

| Código | Descripción | HTTP Status |
|--------|-------------|-------------|
| `UNAUTHORIZED` | Token inválido/expirado | 401 |
| `FORBIDDEN` | Sin permisos | 403 |
| `NOT_FOUND` | Recurso no encontrado | 404 |
| `VALIDATION_ERROR` | Datos inválidos | 400 |
| `DUPLICATE_ENTRY` | Entrada duplicada | 409 |
| `RATE_LIMIT_EXCEEDED` | Límite excedido | 429 |
| `INTERNAL_ERROR` | Error del servidor | 500 |

## 🚦 Rate Limiting

- **Límite**: 10 requests/segundo por IP
- **Burst**: 20 requests máximo
- **Headers de respuesta**:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
  - `X-RateLimit-Reset`

## 🛡️ Permisos por Rol

| Acción | ADMIN | USER |
|--------|-------|------|
| Ver datos | ✅ | ✅ |
| Crear/Editar | ✅ | ❌ |
| Eliminar | ✅ | ❌ |
| Gestionar pagos | ✅ | ❌ |
| Ver reportes | ✅ | ✅ |

## 📝 Formato de Fechas

Todas las fechas en formato ISO 8601:
```
2023-12-15T10:30:00Z
```

## 🔧 Apollo Client Setup

```javascript
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL
});

const authLink = setContext((_, { headers }) => ({
  headers: {
    ...headers,
    authorization: localStorage.getItem('accessToken') 
      ? `Bearer ${localStorage.getItem('accessToken')}` 
      : "",
  }
}));

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache()
});
```

## 📞 Soporte

- **Email**: soporte@asam.org
- **Docs Completa**: [Ver documentación completa](./README.md)
- **Issues**: [GitHub](https://github.com/javicabdev/asam-backend/issues)

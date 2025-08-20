# ASAM Backend - Referencia Rápida para Frontend

## 🚀 Inicio Rápido

### Endpoint Principal
```
Desarrollo: http://localhost:8080/graphql
Producción: https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql
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
    user { 
      id 
      username 
      role 
      emailVerified 
    }
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

### Cambiar Contraseña
```graphql
mutation ChangePassword($input: ChangePasswordInput!) {
  changePassword(input: $input) {
    success
    message
  }
}

# Variables
{ 
  "input": { 
    "currentPassword": "actual123", 
    "newPassword": "nueva456!" 
  } 
}
```

### Recuperar Contraseña
```graphql
mutation RequestPasswordReset($email: String!) {
  requestPasswordReset(email: $email) {
    success
    message
  }
}
```

### Tokens
- **Access Token**: 15 minutos
- **Refresh Token**: 7 días

## 👤 Gestión de Usuarios

### Obtener Usuario Actual
```graphql
query GetCurrentUser {
  getCurrentUser {
    id
    username
    role
    emailVerified
    emailVerifiedAt
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
  }
}

# Variables
{
  "input": {
    "username": "nuevo@ejemplo.com",
    "password": "ContraseñaSegura123!",
    "role": "user"
  }
}
```

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
  admin
  user
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
| Crear/Editar miembros | ✅ | ❌ |
| Eliminar | ✅ | ❌ |
| Gestionar pagos | ✅ | ❌ |
| Ver reportes | ✅ | ✅ |
| Gestionar usuarios | ✅ | ❌ |
| Cambiar su contraseña | ✅ | ✅ |

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
  uri: 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
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

## 🔐 Ejemplo de Manejo de Auth con React

```javascript
// hooks/useAuth.js
import { useState, useEffect } from 'react';
import { useMutation } from '@apollo/client';
import { LOGIN_MUTATION, REFRESH_TOKEN_MUTATION } from './queries';

export const useAuth = () => {
  const [login] = useMutation(LOGIN_MUTATION);
  const [refreshToken] = useMutation(REFRESH_TOKEN_MUTATION);
  
  const signIn = async (username, password) => {
    const { data } = await login({ 
      variables: { input: { username, password } } 
    });
    
    localStorage.setItem('accessToken', data.login.accessToken);
    localStorage.setItem('refreshToken', data.login.refreshToken);
    localStorage.setItem('expiresAt', data.login.expiresAt);
    
    return data.login;
  };
  
  const refresh = async () => {
    const storedRefreshToken = localStorage.getItem('refreshToken');
    const { data } = await refreshToken({
      variables: { input: { refreshToken: storedRefreshToken } }
    });
    
    localStorage.setItem('accessToken', data.refreshToken.accessToken);
    localStorage.setItem('refreshToken', data.refreshToken.refreshToken);
    localStorage.setItem('expiresAt', data.refreshToken.expiresAt);
    
    return data.refreshToken;
  };
  
  const signOut = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('expiresAt');
  };
  
  return { signIn, refresh, signOut };
};
```

## 📞 Soporte

- **Email**: soporte@asam.org
- **Docs Completa**: [Ver documentación completa](./README.md)
- **Issues**: [GitHub](https://github.com/javicabdev/asam-backend/issues)

# Autenticación en ASAM Backend

Este documento describe cómo utilizar el sistema de autenticación implementado en el backend de ASAM.

## Arquitectura

El sistema de autenticación sigue una arquitectura basada en JWT (JSON Web Tokens) y está compuesto por:

1. **Modelos de Dominio**: `User` que contiene la información del usuario
2. **Servicios**: `AuthService` que implementa la lógica de autenticación
3. **Resolvers GraphQL**: Conectan las mutaciones GraphQL con los servicios
4. **Middleware**: Verifica los tokens JWT en las peticiones
5. **Repositorios**: Almacenan y validan los tokens

## Endpints GraphQL

### Login

```graphql
mutation Login($username: String!, $password: String!) {
  login(input: {username: $password, password: $password}) {
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
mutation RefreshToken($refreshToken: JWT!) {
  refreshToken(input: {refreshToken: $refreshToken}) {
    accessToken
    refreshToken
    expiresAt
  }
}
```

## Uso de Tokens

### Envío de Tokens

El token de acceso debe enviarse en todas las peticiones que requieran autenticación usando el header HTTP `Authorization`:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Flujo de Autenticación

1. El cliente llama a la mutación `login` con credenciales válidas
2. El servidor devuelve un `accessToken` y un `refreshToken`
3. El cliente almacena ambos tokens
4. El cliente usa el `accessToken` en sus peticiones
5. Cuando el `accessToken` expira, el cliente usa el `refreshToken` para obtener nuevos tokens
6. Al cerrar sesión, el cliente llama a la mutación `logout` y elimina los tokens

## Roles y Autorización

El sistema soporta dos roles:

- **ADMIN**: Acceso completo a todas las funcionalidades
- **USER**: Acceso restringido a funcionalidades específicas

## Middleware de Autenticación

El middleware de autenticación se encarga de:

1. Extraer el token del header `Authorization`
2. Validar el token usando el servicio de autenticación
3. Cargar la información del usuario en el contexto
4. Permitir o denegar el acceso a los resolvers

## Regeneración del Código

Después de realizar cambios en el esquema GraphQL, es necesario regenerar el código usando:

```
go run github.com/99designs/gqlgen generate
```

## Ejemplo de Cliente

```javascript
// Login
async function login(username, password) {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: `
        mutation Login($username: String!, $password: String!) {
          login(input: {username: $username, password: $password}) {
            accessToken
            refreshToken
            expiresAt
            user { id username role }
          }
        }
      `,
      variables: { username, password }
    })
  });
  
  const data = await response.json();
  return data.data.login;
}

// Realizar petición autenticada
async function fetchData(query, variables, token) {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ query, variables })
  });
  
  return await response.json();
}

// Refrescar token
async function refreshToken(refreshToken) {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: `
        mutation RefreshToken($refreshToken: JWT!) {
          refreshToken(input: {refreshToken: $refreshToken}) {
            accessToken
            refreshToken
            expiresAt
          }
        }
      `,
      variables: { refreshToken }
    })
  });
  
  const data = await response.json();
  return data.data.refreshToken;
}

// Logout
async function logout(token) {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      query: `
        mutation {
          logout {
            success
            message
            error
          }
        }
      `
    })
  });
  
  const data = await response.json();
  return data.data.logout;
}
```

## Siguiente Pasos

1. Implementar creación de usuarios administrativos
2. Añadir middleware de autorización basado en roles
3. Implementar límites de intentos de login
4. Añadir autenticación de dos factores

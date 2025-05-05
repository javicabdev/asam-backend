# Implementación del sistema de autenticación

Este documento describe los pasos necesarios para completar la implementación del sistema de autenticación en el backend de ASAM.

## Estructura actual

El sistema de autenticación está compuesto por:

1. **Mutaciones GraphQL** (schema.graphql):
   - `login(input: LoginInput!): AuthResponse!`
   - `logout: MutationResponse!`
   - `refreshToken(input: RefreshTokenInput!): TokenResponse!`

2. **Resolvers** (auth_resolver.go):
   - Implementan la lógica que conecta las mutaciones con los servicios

3. **Servicios** (auth_service.go):
   - Implementan la lógica de autenticación mediante JWT

4. **Middleware** (auth.go):
   - Verifica los tokens en las peticiones
   - Añade información del usuario al contexto

## Problemas actuales y soluciones

Después de implementar el sistema de autenticación básico, nos encontramos con algunos conflictos al generar el código con gqlgen:

1. **Conflictos de tipos**:
   - Los tipos definidos manualmente en `auth_models.go` entran en conflicto con los generados por gqlgen
   - **Solución**: Eliminar `auth_models.go` y usar los tipos generados por gqlgen

2. **Referencias a campos incorrectos**:
   - Campo `input.Address` vs `input.Address`
   - **Solución**: Actualizar las referencias para usar los nombres correctos de los campos

3. **Funciones no declaradas**:
   - `timestampToTime`
   - **Solución**: Añadir la implementación de esta función

## Pasos para completar la implementación

1. **Regenerar el código con gqlgen**:
   ```bash
   go run github.com/99designs/gqlgen generate
   ```

2. **Ajustar los resolvers para usar los tipos generados**:
   - Actualizar las referencias a tipos y campos para que correspondan con los generados

3. **Implementar el middleware de autorización**:
   - Añadir lógica para verificar permisos según el rol del usuario

4. **Añadir manejo de errores específico para autenticación**:
   - Crear tipos de error para problemas como token expirado, credenciales inválidas, etc.

## Configuración de JWT

El sistema usa JWT para la autenticación con dos tipos de tokens:

1. **Access Token**:
   - Duración corta (15 minutos por defecto)
   - Usado para autenticar peticiones

2. **Refresh Token**:
   - Duración larga (7 días por defecto)
   - Usado para obtener nuevos access tokens sin necesidad de reiniciar sesión

## Seguridad

Para mejorar la seguridad del sistema, se recomienda:

1. **Limitar intentos de login**:
   - Implementar bloqueo temporal después de múltiples intentos fallidos

2. **Rotación de refresh tokens**:
   - Cada vez que se usa un refresh token, generar uno nuevo

3. **Blacklist de tokens**:
   - Mantener una lista de tokens invalidados hasta su expiración

4. **Validación de CSRF**:
   - Para proteger contra ataques CSRF en clientes web

## Pruebas

Para probar el sistema de autenticación:

1. **Crear un usuario administrador**:
   ```sql
   INSERT INTO users (username, password, role, is_active, created_at, updated_at)
   VALUES ('admin', '$2a$10$kZPrU.JXxbJ8ydMp1Rwl1.tFgHIx3Wb5oWLl2NHQGr1rpyOQP8AwS', 'admin', true, NOW(), NOW());
   ```
   (La contraseña es 'password' hasheada con bcrypt)

2. **Probar la mutación login**:
   ```graphql
   mutation {
     login(input: {username: "admin", password: "password"}) {
       user {
         id
         username
         role
       }
       accessToken
       refreshToken
       expiresAt
     }
   }
   ```

3. **Probar endpoint protegido**:
   ```
   curl -X POST -H "Authorization: Bearer [token]" -d '{"query": "query { getBalance }"}' http://localhost:8080/graphql
   ```

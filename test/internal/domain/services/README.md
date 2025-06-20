# Tests del Sistema de Autenticación

Este directorio contiene los tests unitarios completos para el sistema de autenticación del backend ASAM.

## Estructura de Tests

```
test/internal/domain/services/
├── auth_service_test.go    # Tests del servicio de autenticación
└── user_service_test.go    # Tests del servicio de usuarios
```

## Cobertura de Tests

### auth_service_test.go
Tests completos para todas las funciones del servicio de autenticación:

#### Login
- ✅ Login exitoso con credenciales válidas
- ✅ Usuario no encontrado
- ✅ Contraseña incorrecta
- ✅ Usuario inactivo
- ✅ Error de base de datos
- ✅ Error al generar tokens
- ✅ Error al guardar refresh token
- ✅ Login con información de contexto (IP, User-Agent)

#### Logout
- ✅ Logout exitoso
- ✅ Token inválido
- ✅ Error al extraer claims
- ✅ UUID faltante en claims
- ✅ Error al eliminar refresh token

#### RefreshToken
- ✅ Refresh exitoso
- ✅ Token inválido
- ✅ Token no encontrado en BD
- ✅ Usuario no encontrado
- ✅ Error al generar nuevos tokens

#### ValidateToken
- ✅ Validación exitosa
- ✅ Token inválido
- ✅ User ID faltante en claims
- ✅ Usuario no encontrado

### user_service_test.go
Tests completos para la gestión de usuarios:

#### CreateUser
- ✅ Creación exitosa de usuario
- ✅ Validación de username (vacío, muy corto, muy largo, caracteres inválidos)
- ✅ Validación de password (vacío, muy corto, muy largo, sin mayúsculas, sin minúsculas, sin números)
- ✅ Username ya existe
- ✅ Error de base de datos

#### UpdateUser
- ✅ Actualización exitosa
- ✅ Actualización con nuevo password
- ✅ Usuario no encontrado
- ✅ Username ya tomado por otro usuario

#### DeleteUser
- ✅ Eliminación exitosa (soft delete)
- ✅ Usuario no encontrado

#### GetUser
- ✅ Obtención exitosa
- ✅ Usuario no encontrado

#### ChangePassword
- ✅ Cambio exitoso de contraseña
- ✅ Contraseña actual incorrecta
- ✅ Nueva contraseña inválida

#### ResetPassword
- ✅ Reset exitoso (función admin)
- ✅ Usuario no encontrado
- ✅ Nueva contraseña inválida

## Ejecución de Tests

### Ejecutar todos los tests de autenticación
```powershell
# Desde la raíz del proyecto
.\scripts\test-auth-complete.ps1
```

### Ejecutar solo tests del auth service
```powershell
.\scripts\test-auth-service.ps1
```

### Ejecutar tests manualmente
```bash
# Auth service tests
go test -v ./test/internal/domain/services/auth_service_test.go

# User service tests
go test -v ./test/internal/domain/services/user_service_test.go

# Con cobertura
go test -v -cover ./test/internal/domain/services/...

# Con reporte de cobertura detallado
go test -v -coverprofile=coverage.out ./test/internal/domain/services/...
go tool cover -html=coverage.out -o coverage.html
```

## Mocks Utilizados

Los tests utilizan mocks para simular las dependencias:

1. **MockUserRepository**: Simula el repositorio de usuarios
2. **MockTokenRepository**: Simula el repositorio de tokens
3. **MockJWTUtil**: Simula las utilidades JWT
4. **MockLogger**: Logger de prueba (ya existente en test/mocks.go)

## Características de los Tests

### Aislamiento
- Cada test es completamente independiente
- Los mocks permiten simular diferentes escenarios sin afectar la BD

### Cobertura Completa
- Todos los métodos públicos están testeados
- Se cubren casos de éxito y todos los casos de error
- Se verifican las validaciones de entrada

### Verificación de Comportamiento
- Se verifica que los passwords se hasheen correctamente
- Se verifica que los tokens se generen con la información correcta
- Se verifica que las operaciones de BD se llamen con los parámetros correctos
- Se verifica que los passwords nunca se devuelvan en las respuestas

### Manejo de Errores
- Se verifican todos los tipos de errores del sistema
- Se asegura que los mensajes de error sean apropiados
- Se verifica que los códigos de error sean correctos

## Próximos Pasos

1. **Tests de Integración**: Crear tests que prueben el flujo completo con BD real
2. **Tests de Rate Limiting**: Añadir tests para el LoginRateLimiter
3. **Tests de Middleware**: Verificar que el middleware de auth funcione correctamente
4. **Benchmarks**: Añadir benchmarks para operaciones críticas como login y validación de tokens
5. **Tests E2E**: Tests completos desde GraphQL hasta la BD

## Notas Importantes

- Los tests usan contraseñas conocidas para facilitar las pruebas
- El hash de prueba `$2a$10$kZPrU.JXxbJ8ydMp1Rwl1...` corresponde a "password123"
- Los tests verifican que las contraseñas nunca se devuelvan en las respuestas
- Se usa `mock.AnythingOfType("*models.User")` para validar tipos en runtime

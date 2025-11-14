# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Versionado Semántico](https://semver.org/lang/es/).

## [1.6.0] - 2025-11-14

### Changed

#### ⚠️ BREAKING: DeleteUser ahora es Hard Delete
- **Cambio de comportamiento**: `deleteUser` mutation ahora elimina usuarios **permanentemente** de la base de datos
  - **Antes**: Desactivaba el usuario (soft delete) poniendo `is_active = false`
  - **Ahora**: Elimina el registro completamente (hard delete) junto con todos sus tokens
- **Borrado en cascada**: Elimina automáticamente RefreshTokens y VerificationTokens asociados
- **Protecciones implementadas**:
  - No permite eliminar usuarios con Member asociado (constraint `OnDelete:RESTRICT`)
  - Previene eliminar el último administrador del sistema
  - Validación mejorada con conteo de administradores activos
- **Impacto en frontend**:
  - La API GraphQL no cambia, sigue siendo `deleteUser(id: ID!)`
  - Se recomienda actualizar mensajes de confirmación para advertir sobre borrado permanente
  - Implementar manejo de errores específicos para restricciones

### Fixed

#### DeleteUser Database Error
- Corregido error `"Error updating user"` al intentar eliminar usuarios
  - **Causa**: El método Update intentaba guardar asociaciones preloaded (Member)
  - **Solución**: Implementado método `Delete()` específico con transacción atómica
- Implementada transacción para garantizar atomicidad en borrado

### Technical

#### Repository Layer
- Agregado método `Delete(ctx context.Context, userID uint) error` en `UserRepository` interface
- Implementación en `user_repository.go` usa transacciones GORM para borrado atómico
- Mejor manejo de errores con detección de violaciones de constraints

## [1.5.1] - 2025-11-13

### Added

#### Actualización Automática de Pagos Pendientes
- **Sincronización inteligente**: Al cambiar el monto de una cuota anual (ej: 35€ → 40€), los pagos pendientes se actualizan automáticamente
- **Respeto a pagos confirmados**: Solo actualiza pagos en estado PENDING, los pagos PAID no se modifican
- **Auditoría clara**: Nota "Cuota anual actualizada automáticamente" en pagos modificados
- Elimina inconsistencias entre cuotas y pagos pendientes

### Fixed

#### Tests de Integración
- Corregido error de compilación en `payment_cashflow_sync_test.go`
  - Variable global `testDB` reemplazada por patrón correcto `setupTestDB(t)`
  - Agregadas funciones helper para repositorios en tests
  - Modelo `Member` corregido con todos los campos obligatorios

#### GraphQL Resolver - GenerateAnnualFees
- Corregido error "the requested element is null" en campo `error` nullable
  - Solo crea puntero cuando hay error real
  - Retorna `nil` cuando no hay error (cumple schema GraphQL)

#### Métricas de Defaulters
- Eliminada referencia a columna `deleted_at` inexistente en tabla `members`
  - Tabla `members` no usa soft delete
  - Error: "column m.deleted_at does not exist" ya no ocurre

#### GraphQL Resolver - UpdatePayment
- Corregido error "member is null" al actualizar pagos
  - Resolver ahora recarga el payment con todas sus relaciones después de actualizar
  - Campo `member: Member!` siempre retorna objeto completo (no null)
  - Previene pérdida de información de socio y cuota al editar

### Technical

#### Tests
- Implementado método `Update()` en `MockPaymentRepository` para permitir tests de actualización
- Tests de idempotencia ahora verifican correctamente la actualización de pagos

#### Mejoras en Integridad de Datos
- `generatePaymentForMember()` ahora detecta cambios en monto y actualiza pagos pendientes
- Uso correcto de `Preload()` en GORM para cargar relaciones requeridas por GraphQL
- Consistencia mejorada entre estado de cuotas y pagos generados

### Casos de Uso

#### Escenario: Ajuste de Cuota Anual
```
1. Admin genera cuotas 2025 con 35€ → Crea pagos pendientes de 35€
2. Admin cambia cuotas 2025 a 40€ → Actualiza automáticamente pagos pendientes a 40€
3. Los pagos ya confirmados mantienen su monto original
```

#### Antes vs Después
**Antes (v1.5.0)**:
- Cambiar monto de cuota no actualizaba pagos pendientes
- Socio veía monto antiguo (35€) aunque la cuota fuera 40€
- Requerías eliminar y recrear pagos manualmente

**Después (v1.5.1)**:
- Cambiar monto de cuota actualiza automáticamente pagos pendientes
- Socio siempre ve el monto correcto
- Sincronización transparente y automática

---

## [1.5.0] - 2025-11-12

### Added

#### Sincronización Bidireccional Payment-CashFlow con Transacciones ACID
- **Sincronización automática** entre Payment y CashFlow en ambas direcciones
- **Transacciones ACID** que garantizan consistencia de datos
- **Idempotencia** en operaciones de confirmación de pagos
- **Logging de auditoría** con prefijo `[SYNC]` para troubleshooting

#### Nuevos Métodos de Repositorio
- `UpdatePaymentAndSyncCashFlow()` - Actualiza payment y sincroniza su cashflow en transacción
- `UpdateCashFlowAndSyncPayment()` - Actualiza cashflow y sincroniza su payment en transacción
- `ConfirmPaymentWithTransaction()` - Confirma payment y crea cashflow atómicamente

#### Tests de Integración
- `TestPaymentCashFlowSync_UpdatePayment` - Verifica sincronización payment → cashflow
- `TestPaymentCashFlowSync_UpdateCashFlow` - Verifica sincronización cashflow → payment
- `TestConfirmPaymentWithTransaction` - Verifica confirmación atómica con transacción

### Fixed

#### Reset Automático de Verificación de Email
- **Reseteo automático** de `emailVerified` a `false` al cambiar email
- **Reseteo automático** de `emailVerifiedAt` a `null` al cambiar email
- **Validación de disponibilidad** del nuevo email antes de actualizar
- **Logging de auditoría** para cambios de email con valores antiguo y nuevo
- Normalización automática de emails (trim + lowercase)

### Changed

#### Resolvers GraphQL
- `updatePayment` ahora usa método sincronizado para mantener consistencia con cashflow
- `updateTransaction` ahora usa método sincronizado para mantener consistencia con payment
- `updateUser` ahora resetea verificación automáticamente al cambiar email

### Technical

#### Mejoras en Integridad de Datos
- Eliminada función obsoleta `createCashFlowForPayment` (reemplazada por transacciones)
- Todos los métodos de sincronización usan `db.Transaction()` para garantizar atomicidad
- Rollback automático en caso de error durante sincronización
- Sin cambios requeridos en frontend - sincronización transparente

#### Mejoras en User Service
- Nueva función `updateEmail()` que maneja toda la lógica de actualización de email
- Validación de duplicados de email antes de actualizar
- Consistencia mejorada en normalización de emails

### Security

#### Garantías de Consistencia
- **Atomicidad**: Payment y CashFlow se actualizan juntos o ninguno
- **Consistencia**: Imposible que Payment y CashFlow tengan valores diferentes
- **Aislamiento**: Transacciones protegen contra condiciones de carrera
- **Durabilidad**: Cambios confirmados permanecen en la base de datos

#### Auditoría de Email
- Log detallado cuando se cambia el email de un usuario
- Tracking de email antiguo y nuevo para investigación
- Reset automático de verificación previene acceso con email no verificado

---

## [1.4.0] - 2025-11-12

### Added

#### Sliding Expiration para Sesiones de Usuario
- **Sistema de expiración deslizante (sliding expiration)** que extiende automáticamente las sesiones cuando el usuario está activo
- Las sesiones se extienden automáticamente con cada refresh de token, mejorando la experiencia de usuario
- **Límite absoluto de 30 días** desde el login inicial (configurable)
- **Timeout de inactividad de 7 días** - sesiones expiran si no se usan durante este período
- **Ventana de extensión de 24 horas** - cada refresh extiende la sesión por este tiempo
- Campo `last_used_at` actualizado automáticamente en cada uso del token

#### Nuevas Variables de Entorno
- `TOKEN_SLIDING_EXPIRATION=true` - Habilitar/deshabilitar sliding expiration
- `TOKEN_SLIDING_WINDOW=24h` - Tiempo de extensión en cada refresh
- `TOKEN_ABSOLUTE_MAX_LIFETIME=720h` - Límite absoluto de sesión (30 días)
- `TOKEN_INACTIVITY_TIMEOUT=168h` - Timeout por inactividad (7 días)

### Technical

#### Repositorio de Tokens
- Nuevo método `GetRefreshToken(uuid)` - Obtiene información completa del token
- Nuevo método `ExtendTokenExpiration(uuid, newExpires)` - Extiende la expiración del token

#### Servicio de Autenticación
- Lógica de sliding expiration implementada en `RefreshToken()`
- Método `shouldApplySlidingExpiration()` - Verifica políticas de extensión
- Método `createNewRefreshTokenWithSlidingExpiration()` - Crea tokens con extensión
- Logging detallado de extensiones y límites alcanzados

### Security

#### Políticas de Seguridad
- **Límite absoluto**: Fuerza nuevo login después de 30 días, independientemente de la actividad
- **Detección de inactividad**: Expira sesiones no utilizadas en 7+ días
- **Trazabilidad mejorada**: Campo `last_used_at` permite auditoría de uso de sesiones
- **Configuración flexible**: Permite ajustar políticas según necesidades de seguridad

### Casos de Uso

#### Escenario 1: Usuario Activo Diario
```
Día 1: Login → Token expira en 24h
Día 2: Refresh → Token se extiende +24h (expira día 3)
Día 3-29: Uso continuo → Token se extiende cada vez
Día 30: Refresh → Límite absoluto alcanzado → REQUIERE LOGIN
```

#### Escenario 2: Usuario Inactivo
```
Día 1: Login → Token expira en 24h
Días 2-7: Sin actividad
Día 8: Usuario intenta acceder → Token expiró por inactividad → REQUIERE LOGIN
```

#### Escenario 3: Configuración Deshabilitada
```
TOKEN_SLIDING_EXPIRATION=false
Comportamiento tradicional: Token expira en tiempo fijo (JWT_REFRESH_TTL)
```

### Configuration Examples

#### Aplicación Pública (Alta Seguridad)
```env
TOKEN_SLIDING_EXPIRATION=true
TOKEN_SLIDING_WINDOW=12h          # Extensión corta
TOKEN_ABSOLUTE_MAX_LIFETIME=168h  # 7 días máximo
TOKEN_INACTIVITY_TIMEOUT=24h      # 1 día sin uso
```

#### Aplicación Interna (Mayor Conveniencia)
```env
TOKEN_SLIDING_EXPIRATION=true
TOKEN_SLIDING_WINDOW=24h          # Extensión de 1 día
TOKEN_ABSOLUTE_MAX_LIFETIME=720h  # 30 días máximo
TOKEN_INACTIVITY_TIMEOUT=168h     # 7 días sin uso
```

---

## [1.3.0] - 2025-11-11

### Added
- Parámetro opcional `amount` en mutación `confirmPayment`
  - Permite actualizar el monto del pago al momento de confirmarlo
  - Si no se proporciona, mantiene el valor actual del pago
  - Validación automática que el monto sea mayor que cero
  - Actualización condicional: solo modifica si el valor es diferente al actual

### Technical
- Actualizado schema GraphQL con nuevo parámetro opcional `amount: Float`
- Lógica de validación de monto en `payment_service.go`
- Interfaz `PaymentService` actualizada con nuevo parámetro
- Código generado por gqlgen actualizado automáticamente
- Resolver GraphQL modificado para pasar el parámetro al servicio

---

## [1.2.0] - 2025-11-10

### Fixed
- Corregidos nombres de columnas en query de métricas de morosos (defaulters)
  - `m.numero_socio` → `m.membership_number`
  - `m.correo_electronico` → `m.email`
  - `m.full_name` → `CONCAT(m.name, ' ', m.surnames)`
- Organización de scripts temporales en `cmptemp/` para mejorar estructura del proyecto
- Solo `cmd/api` y `cmd/migrate` permanecen en el repositorio

### Technical
- Configuración de linter optimizada para escanear solo directorios relevantes
- Scripts de desarrollo movidos a `cmptemp/` (ignorados por git)
- Mantenimiento de `cmd/migrate` en repositorio para compatibilidad con CI/CD

---

## [1.1.0] - 2025-11-10

### Added

#### Datos Históricos en Altas de Socios
- Campo opcional `fecha_alta` en `CreateMemberInput` para especificar fechas de alta históricas
- Generación automática de pagos pendientes para **todos los años** desde la fecha de alta hasta el año actual
- Validación que impide crear socio si faltan cuotas anuales en el rango de años requerido
- Mensajes de error claros indicando qué años de cuotas faltan

#### Query para Listar Cuotas Anuales
- Nueva query `listAnnualFees`: retorna todas las cuotas anuales sin paginación (límite 1000)
- Nueva query `listMembershipFees(page, pageSize)`: listado con paginación opcional
- Ordenamiento por año descendente (más reciente primero)
- Solo accesible para usuarios con rol ADMIN

#### Filtrado Inteligente en Generación de Cuotas
- `GenerateAnnualFees` ahora respeta la fecha de alta del socio
- No genera pagos para años anteriores a la fecha de alta del socio
- Evita crear obligaciones de pago incorrectas para socios con altas históricas

### Changed
- **Comportamiento de creación de socios**: Al crear un socio, ahora se generan múltiples pagos pendientes (uno por cada año desde su fecha de alta hasta el actual), en lugar de solo un pago para el año actual

### Technical
- Nuevo método `FindAll(limit, offset)` en `MembershipFeeRepository` con paginación
- Nuevo método `ListMembershipFees(page, pageSize)` en `PaymentService`
- Actualizado schema GraphQL con nuevos campos y queries
- Método `createPendingPayment` refactorizado para generar múltiples pagos
- Método `generatePaymentForMember` actualizado con filtrado por fecha de alta
- Mocks de tests actualizados con método `FindAll`

### Casos de Uso

#### Ejemplo 1: Alta Histórica (2022)
```
Socio dado de alta: 15/03/2022
Año actual: 2025
Resultado: Se generan 4 pagos pendientes (2022, 2023, 2024, 2025)
```

#### Ejemplo 2: Validación de Cuotas Faltantes
```
Intento de alta: 01/01/2020
Cuotas disponibles: 2020, 2024, 2025
Resultado: ERROR - Faltan cuotas para 2021, 2022, 2023
```

#### Ejemplo 3: Generación Masiva Inteligente
```
GenerateAnnualFees(2021)
- Socio A (alta 2020): ✓ Se genera pago de 2021
- Socio B (alta 2022): ✗ NO se genera pago (alta posterior)
```

---

## [1.0.0] - 2025-11-08

### Initial Release
- Sistema completo de gestión de socios ASAM
- Gestión de miembros individuales y familiares
- Sistema de pagos y cuotas
- Gestión de flujo de caja (cash flow)
- Autenticación y autorización con JWT
- API GraphQL completa
- Panel de administración
- Métricas y logs de auditoría
- Tests unitarios e integración
- Documentación completa

---

## Leyenda

- `Added`: Nuevas funcionalidades
- `Changed`: Cambios en funcionalidad existente
- `Deprecated`: Funcionalidades obsoletas (próximas a eliminar)
- `Removed`: Funcionalidades eliminadas
- `Fixed`: Corrección de bugs
- `Security`: Correcciones de seguridad
- `Technical`: Cambios técnicos internos

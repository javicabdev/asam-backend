# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Versionado Semántico](https://semver.org/lang/es/).

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

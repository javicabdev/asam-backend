# Control de Acceso Basado en Socio - Resumen de Implementación

## Descripción General
Se ha implementado un sistema de control de acceso que diferencia entre usuarios administradores y usuarios regulares asociados a socios.

### Principios Implementados:
- **Usuarios ADMIN**: Acceso completo de lectura/escritura a todos los datos
- **Usuarios USER**: Solo acceso de lectura a sus propios datos de socio
- **Validación en Login**: Usuarios USER deben tener socio asociado para autenticarse

## Cambios Realizados

### 1. Modelo de Datos (Commit 1)
- Agregado campo `MemberID *uint` en el modelo User
- Creada relación 1:1 entre User y Member
- Agregados métodos de validación en el modelo

### 2. Migración de Base de Datos (Commit 2)
- Nueva columna `member_id` en tabla `users`
- Índice único para garantizar relación 1:1
- Foreign key constraint con tabla `members`

### 3. Servicios de Autenticación (Commit 3)
- **AuthService**: Valida que USER tenga MemberID en login
- **UserService**: Valida coherencia rol-memberID al crear/actualizar
- Agregado método `FindByMemberID` en UserRepository

### 4. Middleware de Autorización (Commit 4)
Nuevos helpers implementados:
- `GetMemberIDFromContext()`: Obtiene MemberID del usuario actual
- `CanAccessMember()`: Verifica acceso a un socio específico
- `CanAccessFamily()`: Verifica acceso a una familia
- `CanAccessPayment()`: Verifica acceso a un pago
- `IsUserMember()`: Verifica si es USER con socio
- `GetCurrentUserMember()`: Obtiene MemberID si existe

### 5. Resolvers GraphQL (Commit 5)
Actualización de queries con control de acceso:

#### Queries de Member:
- **getMember**: Aplica `CanAccessMember()`
- **listMembers**: ADMIN ve todo, USER solo su registro
- **searchMembers**: ADMIN busca todo, USER solo en su registro

#### Queries de Family:
- **getFamily**: Aplica `CanAccessFamily()`
- **listFamilies**: ADMIN ve todo, USER solo donde es miembro origen
- **getFamilyMembers**: Verifica permisos de familia

#### Queries de Payment:
- **getPayment**: Aplica `CanAccessPayment()`
- **getMemberPayments**: Verifica permisos del miembro
- **getFamilyPayments**: Verifica permisos de familia
- **getPaymentStatus**: Verifica permisos del pago

#### Queries de CashFlow:
- **getCashFlow**: Solo ADMIN
- **getBalance**: Solo ADMIN
- **getTransactions**: Solo ADMIN

#### Mutations:
- Todas las mutations requieren rol ADMIN

### 6. Campo Member en User
- Agregado resolver para el campo `member` del tipo User
- Aplica control de acceso al cargar datos del socio

## Comportamiento del Sistema

### Para Administradores:
1. Pueden autenticarse sin socio asociado
2. Acceso completo a todos los datos
3. Pueden crear, actualizar y eliminar cualquier registro
4. Pueden ver transacciones y balance general

### Para Usuarios Regulares:
1. DEBEN tener socio asociado para hacer login
2. Solo pueden ver sus propios datos en modo lectura:
   - Su información personal como socio
   - Sus pagos
   - Familias donde son miembro origen
3. NO pueden:
   - Ver datos de otros socios
   - Modificar ningún dato (solo lectura)
   - Ver transacciones o balance general
   - Acceder si no tienen socio asociado

## Consideraciones de Seguridad

1. **Validación en Login**: Primera línea de defensa
2. **Middleware de Autorización**: Segunda validación en cada request
3. **Control Granular**: Cada query verifica permisos específicos
4. **Sin Estados Ambiguos**: USER sin socio no puede autenticarse

## Testing

Se han incluido tests unitarios para:
- Helpers de autorización del middleware
- Control de acceso en resolvers principales
- Casos de éxito y error para ADMIN y USER

## Próximos Pasos Sugeridos

1. **Interfaz de Admin**: Crear UI para asociar usuarios a socios
2. **Migración de Datos**: Script para asociar usuarios existentes
3. **Auditoría**: Log de accesos para seguridad
4. **Expansión**: Incluir cónyuges/familiares en permisos de familia

# Gestión de Usuarios y Sesiones

Este directorio contiene scripts para la gestión de usuarios y sesiones en ASAM Backend.

## Arquitectura de Autenticación

El sistema utiliza JWT con un diseño robusto que incluye:

### 1. Tabla `users`
Almacena la información básica del usuario:
- Credenciales (username, password hasheado)
- Rol (admin/user)
- Estado (activo/inactivo)
- Última conexión

### 2. Tabla `refresh_tokens` (NUEVA)
Gestiona las sesiones activas con:
- UUID único por token
- Referencia al usuario
- Fecha de expiración
- Información del dispositivo (opcional)
- IP y User-Agent
- Última vez usado

## Ventajas del diseño con tabla separada

1. **Múltiples sesiones simultáneas**: Un usuario puede estar conectado en varios dispositivos
2. **Revocación selectiva**: Cerrar sesión en un dispositivo sin afectar otros
3. **Auditoría de sesiones**: Ver todas las sesiones activas de un usuario
4. **Limpieza automática**: Eliminar tokens expirados sin tocar la tabla users
5. **Mejor seguridad**: Información de sesión separada de datos del usuario
6. **Escalabilidad**: Fácil agregar nuevos campos sin modificar users

## Aplicar las migraciones

```powershell
# Aplicar la nueva migración para crear refresh_tokens
.\scripts\migrate.ps1

# Verificar que se creó correctamente
.\scripts\migrate.ps1 development version
```

## Scripts disponibles

### 1. auto-create-test-users.go
Crea automáticamente usuarios de prueba para desarrollo.

```powershell
# Se ejecuta automáticamente al iniciar Docker
# O manualmente:
go run scripts/user-management/auto-create-test-users.go
```

### 2. manage-users.go
Herramienta interactiva para gestión manual de usuarios.

```powershell
.\scripts\user-management\manage-users.ps1
```

Funcionalidades:
- Crear nuevos usuarios
- Listar usuarios existentes
- Eliminar usuarios
- Resetear contraseñas

### 3. manage-sessions.go (Por implementar)
Herramienta para gestión de sesiones activas.

Funcionalidades futuras:
- Ver sesiones activas por usuario
- Revocar sesión específica
- Revocar todas las sesiones de un usuario
- Limpiar tokens expirados

## Flujo de autenticación

1. **Login**: 
   - Usuario envía credenciales
   - Sistema valida y genera access token (15 min) y refresh token (7 días)
   - Refresh token se guarda en tabla `refresh_tokens` con info del dispositivo

2. **Uso de API**:
   - Cliente envía access token en header Authorization
   - Si expira, usa refresh token para obtener nuevo access token

3. **Logout**:
   - Elimina el refresh token específico de la sesión
   - Otros dispositivos siguen conectados

4. **Logout global**:
   - Elimina todos los refresh tokens del usuario
   - Fuerza re-login en todos los dispositivos

## Mantenimiento

### Limpieza automática de tokens expirados
Configura un cron job o tarea programada:

```go
// Ejemplo de función para limpieza (a ejecutar diariamente)
func CleanupExpiredTokens() {
    db.Where("expires_at < ?", time.Now().Unix()).
       Delete(&models.RefreshToken{})
}
```

## Seguridad

1. **Contraseñas**: Hasheadas con bcrypt (cost 10)
2. **Tokens JWT**: 
   - Access token: Corta duración, contiene claims mínimos
   - Refresh token: UUID aleatorio, almacenado hasheado
3. **Sesiones**: Información de dispositivo para detectar accesos sospechosos
4. **Revocación**: Inmediata al eliminar de BD

## Próximas mejoras

1. **Notificaciones de seguridad**: Email cuando se detecta login desde nuevo dispositivo
2. **Límite de sesiones**: Máximo N sesiones simultáneas por usuario
3. **Geolocalización**: Detectar y alertar logins desde ubicaciones inusuales
4. **2FA**: Autenticación de dos factores
5. **API de sesiones**: Endpoint para que usuarios vean/gestionen sus sesiones

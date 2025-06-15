# PASOS PARA SOLUCIONAR EL PROBLEMA

## El problema
El modelo `User` tiene campos para refresh tokens que no existen en la base de datos, causando el error al crear usuarios.

## La solución implementada
Hemos implementado una arquitectura robusta con tabla `refresh_tokens` separada que permite:
- Múltiples sesiones por usuario
- Gestión de dispositivos
- Revocación selectiva de sesiones
- Mejor seguridad y auditoría

## Pasos a seguir:

### 1. Aplicar la nueva migración

```powershell
# Desde la raíz del proyecto
.\scripts\migrate.ps1

# Verificar que se aplicó correctamente
.\scripts\migrate.ps1 development version
```

Esto creará la tabla `refresh_tokens` con todos los índices necesarios.

### 2. Reiniciar el entorno de desarrollo

```powershell
# Opción 1: Usando Make
make dev-restart

# Opción 2: Usando Docker Compose
docker-compose down
docker-compose up -d
```

### 3. Los usuarios de prueba se crearán automáticamente

El script `auto-create-test-users.go` se ejecutará automáticamente y creará:
- admin@asam.org / admin123
- user@asam.org / admin123

### 4. Herramientas de gestión disponibles

```powershell
# Gestión de usuarios
.\scripts\user-management\manage-users.ps1

# Gestión de sesiones (NUEVO)
.\scripts\user-management\manage-sessions.ps1
```

## Verificación

Para verificar que todo funciona correctamente:

```powershell
# Ver logs del contenedor
docker-compose logs -f api

# Acceder a la base de datos
docker-compose exec postgres psql -U postgres -d asam_db

# En psql, verificar las tablas
\dt
\d users
\d refresh_tokens
```

## Beneficios del nuevo diseño

1. **Seguridad mejorada**: Sesiones independientes por dispositivo
2. **Mejor UX**: Usuario puede ver y gestionar sus sesiones activas
3. **Escalabilidad**: Fácil agregar nuevas funcionalidades
4. **Cumplimiento**: Auditoría completa de accesos
5. **Mantenimiento**: Limpieza automática de tokens expirados

## Próximos pasos opcionales

1. Crear endpoint GraphQL para que usuarios vean sus sesiones
2. Agregar notificaciones de login desde nuevo dispositivo
3. Implementar límite de sesiones simultáneas
4. Agregar 2FA

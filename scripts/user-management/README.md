# Gestión de Usuarios ASAM

Este directorio contiene scripts para gestionar usuarios del sistema ASAM.

## Uso

### Windows
```batch
run.bat
```

### Linux/Mac
```bash
go run manage_users.go
```

## Funcionalidades

1. **Crear usuarios**: Con nombre de usuario, contraseña y rol (admin/user)
2. **Listar usuarios**: Ver todos los usuarios del sistema
3. **Activar/Desactivar usuarios**: Cambiar el estado de un usuario
4. **Cambiar contraseña**: Actualizar la contraseña de un usuario

## Notas de Seguridad

- Este script debe ejecutarse SOLO desde el servidor
- NO exponer estos scripts a través de la API
- Las contraseñas se hashean automáticamente con bcrypt
- Requiere acceso directo a la base de datos

## Usuarios por defecto recomendados

1. **admin** - Usuario administrador principal
2. **user1**, **user2**, etc. - Usuarios normales según necesidad
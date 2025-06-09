# Gestión de Usuarios ASAM

Este directorio contiene herramientas para gestionar usuarios del sistema ASAM.

## Herramientas Disponibles

### 1. manage_users.go
Herramienta interactiva para gestión manual de usuarios.

**Uso:**
```batch
# Windows
run.bat

# Linux/Mac/Docker
go run manage_users.go
```

**Funcionalidades:**
- Crear usuarios con nombre, contraseña y rol (admin/user)
- Listar todos los usuarios del sistema
- Activar/Desactivar usuarios
- Cambiar contraseñas

### 2. auto-create-test-users.go
Script automatizado para crear usuarios de prueba sin interacción.

**Uso:**
```bash
# Dentro del contenedor Docker
go run scripts/user-management/auto-create-test-users.go

# Desde el host
docker-compose exec -T api go run scripts/user-management/auto-create-test-users.go
```

**Usuarios creados:**
- admin@asam.org / admin123 (rol: admin)
- user@asam.org / admin123 (rol: user)

Este script es ejecutado automáticamente por `start-docker.ps1`.

## Integración con Docker

Al iniciar el proyecto con `start-docker.ps1`, se ejecuta automáticamente el script `auto-create-test-users.go` para crear los usuarios de prueba necesarios.

## Notas de Seguridad

- Estos scripts deben ejecutarse SOLO desde el servidor o contenedor
- NO exponer estas herramientas a través de la API
- Las contraseñas se hashean automáticamente con bcrypt
- Requieren acceso directo a la base de datos
- Los usuarios de prueba con contraseñas simples solo deben usarse en desarrollo
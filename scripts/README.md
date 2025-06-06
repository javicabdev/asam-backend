# Scripts de Base de Datos

Esta carpeta contiene scripts útiles para la gestión de la base de datos.

## create-test-users.sql

Crea usuarios de prueba con contraseñas predefinidas:

- **admin@asam.org** / admin123 (rol ADMIN)
- **user@asam.org** / admin123 (rol USER)

### Uso con Docker

```bash
# Desde el directorio asam-backend
docker-compose exec -T postgres psql -U postgres -d asam_db < scripts/create-test-users.sql
```

### Uso directo con psql

```bash
psql -U postgres -d asam_db < scripts/create-test-users.sql
```

## Crear usuarios adicionales

Para crear usuarios adicionales, usa el siguiente formato SQL:

```sql
INSERT INTO users (username, password, role, is_active, created_at, updated_at)
VALUES (
    'nuevo-usuario@asam.org',
    '$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG',  -- 'admin123' hasheado
    'ADMIN',  -- o 'USER'
    true,
    NOW(),
    NOW()
);
```

### Hashes de contraseñas precalculados

- **admin123**: `$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG`

Para generar nuevos hashes, necesitarías modificar el backend para incluir una función de utilidad.

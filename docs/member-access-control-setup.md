# Implementación de Control de Acceso Basado en Socio

## Resumen de Cambios

Se ha implementado un sistema de control de acceso diferenciado entre usuarios administradores y usuarios regulares asociados a socios.

## Migraciones de Base de Datos

### Ejecutar la migración

```powershell
# Desde la raíz del proyecto
migrate -path ./migrations -database "postgres://user:password@localhost:5432/asam_db?sslmode=disable" up
```

La migración `000009_add_member_id_to_users` agrega:
- Campo `member_id` en tabla `users` 
- Índice único para garantizar relación 1:1
- Foreign key constraint con tabla `members`

### Rollback si es necesario

```powershell
migrate -path ./migrations -database "postgres://user:password@localhost:5432/asam_db?sslmode=disable" down 1
```

## Seeds de Prueba

### Ejecutar el seed de usuario con socio

Agrega el siguiente código a tu archivo de seeds principal o ejecuta directamente:

```go
// En cmptemp/seed/main.go o similar
package main

import (
    "log"
    "github.com/javicabdev/asam-backend/internal/adapters/db/seeds"
    // ... otros imports
)

func main() {
    // ... conexión a BD ...
    
    // Ejecutar seed de usuario con socio
    if err := seeds.SeedUserWithMember(db); err != nil {
        log.Fatal("Error seeding user with member:", err)
    }
    
    // Opcional: asociar usuarios existentes
    if err := seeds.SeedUserMemberAssociations(db); err != nil {
        log.Fatal("Error associating users:", err)
    }
}
```

### Datos de prueba creados

1. **Usuario Admin** (ya existente):
   - Username: `admin`
   - Password: `admin123`
   - Rol: ADMIN
   - Sin socio asociado

2. **Usuario con Socio**:
   - Username: `juan.perez`
   - Password: `password123`
   - Rol: USER
   - Socio: Juan Pérez García (M001)

3. **Socio sin Usuario** (para asociación futura):
   - Nombre: María González López
   - Número: M002
   - Disponible para asociar a un nuevo usuario

## Probar el Sistema

### 1. Login como Admin

```graphql
mutation {
  login(input: {
    username: "admin"
    password: "admin123"
  }) {
    user {
      id
      username
      role
      member {
        miembro_id
        nombre
      }
    }
    accessToken
  }
}
```

Resultado esperado: Login exitoso, `member` es null

### 2. Login como Usuario con Socio

```graphql
mutation {
  login(input: {
    username: "juan.perez"
    password: "password123"
  }) {
    user {
      id
      username
      role
      member {
        miembro_id
        nombre
        apellidos
      }
    }
    accessToken
  }
}
```

Resultado esperado: Login exitoso, `member` contiene datos del socio

### 3. Queries como Usuario Regular

```graphql
# Ver sus propios datos
query {
  getMember(id: "1") {  # Usar el ID real del socio
    miembro_id
    nombre
    apellidos
    correo_electronico
  }
}

# Listar miembros (solo verá el suyo)
query {
  listMembers {
    nodes {
      miembro_id
      nombre
      apellidos
    }
    pageInfo {
      totalCount
    }
  }
}
```

### 4. Intentar acceso no autorizado

```graphql
# Como usuario regular, intentar ver otro socio
query {
  getMember(id: "2") {  # ID de otro socio
    nombre
  }
}
```

Resultado esperado: Error de autorización

## Crear Usuario para Socio Existente (Admin)

```graphql
mutation {
  createUser(input: {
    username: "maria.gonzalez"
    email: "maria.gonzalez@example.com"
    password: "password123"
    role: user
    memberId: "2"  # ID del socio M002
  }) {
    id
    username
    member {
      nombre
      apellidos
    }
  }
}
```

## Verificación de Permisos

### Como Admin puedes:
- ✅ Ver todos los socios
- ✅ Crear/editar/eliminar socios
- ✅ Ver todas las transacciones
- ✅ Gestionar usuarios

### Como Usuario regular puedes:
- ✅ Ver tus datos de socio
- ✅ Ver tus pagos
- ✅ Ver familias donde eres miembro origen
- ❌ Ver otros socios
- ❌ Modificar datos
- ❌ Ver transacciones generales

## Troubleshooting

### Error: "Tu usuario no está asociado a ningún socio"

1. Verifica que el usuario tenga rol USER
2. Asocia el usuario a un socio usando la mutation updateUser (como admin)

### Error: "No tienes permiso para acceder a este socio"

Esto es normal para usuarios regulares intentando acceder a datos de otros socios.

### Asociar usuario existente a socio

```graphql
mutation {
  updateUser(input: {
    id: "USER_ID"
    memberId: "MEMBER_ID"
  }) {
    id
    username
    member {
      nombre
    }
  }
}
```

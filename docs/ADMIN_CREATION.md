# Creación de Usuarios Administradores

## Seguridad Importante

**NUNCA incluyas contraseñas en archivos de migración SQL o en el código fuente**, incluso si están hasheadas. Las contraseñas deben manejarse siempre a través de variables de entorno o sistemas de gestión de secretos.

## Proceso para Crear un Usuario Administrador

### 1. Ejecutar las migraciones

Primero asegúrate de que la base de datos está actualizada:

```bash
go run cmptemp/migrate/main.go -cmd up
```

### 2. Configurar las credenciales del administrador

Edita tu archivo `.env.development` (NO commitear con valores reales):

```env
ADMIN_EMAIL=admin@tudominio.com
ADMIN_PASSWORD=ContraseñaSegura123!
ADMIN_USERNAME=admin  # Opcional, si no se proporciona usa el email como username
```

**Variables explicadas:**
- `ADMIN_EMAIL`: Email real del administrador (obligatorio) - se usará para notificaciones
- `ADMIN_PASSWORD`: Contraseña inicial del administrador (obligatorio)
- `ADMIN_USERNAME`: Nombre de usuario para login (opcional)
  - Si no se proporciona, usará el email como username
  - Puedes usar un username corto como "admin" para facilitar el login
  - El username es lo que se usará para iniciar sesión en el sistema

**Requisitos de contraseña:**
- Mínimo 8 caracteres
- Al menos una mayúscula
- Al menos una minúscula
- Al menos un número

### 3. Crear el usuario administrador

Ejecuta el comando de creación:

```bash
# Para entorno local
go run cmptemp/create-admin/main.go -env local

# Para entorno Aiven
go run cmptemp/create-admin/main.go -env aiven

# Para actualizar un admin existente
go run cmptemp/create-admin/main.go -env local -force
```

### 4. Verificación de email

El usuario administrador creado tendrá:
- `email_verified = false`
- `email_verified_at = NULL`

Deberá verificar su email en el primer inicio de sesión.

## Alternativa con Variables de Entorno Directas

También puedes pasar las credenciales directamente sin modificar el archivo `.env`:

```bash
# Con username personalizado (recomendado para facilitar el login)
ADMIN_EMAIL=admin@tudominio.com \
ADMIN_PASSWORD=ContraseñaSegura123! \
ADMIN_USERNAME=admin \
go run cmptemp/create-admin/main.go -env local

# Sin username (usará el email como username)
ADMIN_EMAIL=admin@tudominio.com \
ADMIN_PASSWORD=ContraseñaSegura123! \
go run cmptemp/create-admin/main.go -env local
```

## Ejemplos de Uso

### Ejemplo 1: Admin con username corto
```bash
ADMIN_EMAIL=juan.perez@empresa.com \
ADMIN_PASSWORD=SuperSegura2024! \
ADMIN_USERNAME=juanp \
go run cmptemp/create-admin/main.go -env local
```
Resultado:
- Login con: `juanp`
- Email de notificaciones: `juan.perez@empresa.com`

### Ejemplo 2: Admin usando email como username
```bash
ADMIN_EMAIL=admin@empresa.com \
ADMIN_PASSWORD=AdminPass2024! \
go run cmptemp/create-admin/main.go -env local
```
Resultado:
- Login con: `admin@empresa.com`
- Email de notificaciones: `admin@empresa.com`

## Mejores Prácticas de Seguridad

1. **Nunca commitees archivos `.env` con credenciales reales**
2. **Usa contraseñas fuertes y únicas** para cada entorno
3. **Rota las credenciales regularmente**
4. **Habilita 2FA** cuando sea posible
5. **Audita los accesos** de administrador
6. **Limita el número de usuarios administradores**

## Para Producción

En producción, considera usar:
- Gestores de secretos (AWS Secrets Manager, HashiCorp Vault, etc.)
- Variables de entorno del sistema CI/CD
- Kubernetes Secrets si usas K8s
- Nunca hardcodees credenciales en el código o archivos de configuración

## Troubleshooting

### Error: "Admin user already exists"
Usa la flag `-force` para actualizar el usuario existente:
```bash
go run cmptemp/create-admin/main.go -env local -force
```

### Error: "ADMIN_EMAIL environment variable is required"
Asegúrate de que las variables están configuradas en tu archivo `.env.development` o pásalas directamente en el comando.

### Error: "Password does not meet complexity requirements"
La contraseña debe tener al menos 8 caracteres, una mayúscula, una minúscula y un número.

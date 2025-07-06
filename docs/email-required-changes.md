# Pasos para aplicar los cambios de Email obligatorio

Este documento describe los pasos necesarios para hacer el campo email obligatorio en la tabla users.

## 1. Detener el contenedor de la aplicación (no la base de datos)

```powershell
docker stop asam-backend-api
```

## 2. Ejecutar el script de actualización de emails

Este script actualizará los usuarios existentes con el email correcto:

```powershell
.\scripts\update-test-users-email.ps1
```

## 3. Ejecutar la migración

Ejecuta la migración para hacer el campo email obligatorio:

```powershell
docker exec -i asam-postgres psql -U postgres -d asam_db -f /tmp/migrate.sql
```

Pero primero, copia el archivo de migración:

```powershell
docker cp migrations/000007_make_email_required.up.sql asam-postgres:/tmp/migrate.sql
docker exec -i asam-postgres psql -U postgres -d asam_db -f /tmp/migrate.sql
```

## 4. Reiniciar la aplicación

```powershell
docker start asam-backend-api
```

O simplemente reiniciar todo el stack:

```powershell
docker-compose down
docker-compose up -d
```

## 5. Verificar que todo funcione

Intenta hacer login y enviar un email de verificación. Ahora debería funcionar correctamente.

## Cambios realizados en el código:

1. **Modelo User** (`internal/domain/models/user.go`):
   - Campo `Email` cambió de `*string` a `string` (obligatorio)

2. **Migración** (`migrations/000007_make_email_required.up.sql`):
   - Actualiza emails NULL con el username
   - Hace el campo NOT NULL

3. **Script de usuarios de prueba** (`scripts/user-management/auto-create-test-users/auto-create-test-users.go`):
   - Actualizado para usar `user.Email` en lugar de `&email`
   - Establece `javierfernandezc@gmail.com` como email por defecto

4. **Adaptadores de email** (SMTP y Mock):
   - Actualizados para usar `user.Email` directamente sin verificar nil

5. **Resolver de email** (`internal/adapters/gql/resolvers/email_resolver.go`):
   - Actualizado para verificar `user.Email == ""` en lugar de `user.Email == nil`

## Notas importantes:

- El email ahora es obligatorio para todos los usuarios
- Los usuarios existentes sin email serán actualizados para usar su username como email
- Los nuevos usuarios de prueba usarán `javierfernandezc@gmail.com` como email

## 🛠️ Solución de Problemas - ASAM Backend con Docker

## Error: go.mod requires go >= 1.24

Este error ocurre porque el proyecto requiere Go 1.24, pero es posible que no exista una imagen Docker oficial para esta versión todavía.

### Soluciones:

#### Opción 1: Usar Go 1.23 (Recomendado temporalmente)

Edita el archivo `go.mod` y cambia:
```
go 1.24
```
Por:
```
go 1.23
```

#### Opción 2: Usar la imagen más reciente de Go

Edita `Dockerfile.dev` y cambia:
```dockerfile
FROM golang:1.24-alpine AS builder
```
Por:
```dockerfile
FROM golang:alpine AS builder
```

## Error: The "POSTGRES_DB" variable is not set

Ya está solucionado en el archivo `.env`. Asegúrate de que contenga:
```env
POSTGRES_DB=asam_db
```

## Error: JWT_ACCESS_SECRET missing

El backend requiere variables JWT específicas. Asegúrate de que tu `.env` contenga:
```env
# JWT Configuration (REQUIRED)
JWT_ACCESS_SECRET=dev-access-secret-change-in-production
JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h  # 7 days
```

## Error: time: unknown unit "d" in duration

Go no reconoce "d" como unidad de tiempo. Usa:
- `168h` en lugar de `7d` (7 días = 168 horas)
- `720h` en lugar de `30d` (30 días = 720 horas)

**Solución rápida**: Ejecuta `fix-jwt-restart.ps1` o `fix-jwt-restart.bat`

## Pasos para reiniciar todo

1. **Detén todo**:
   ```bash
   docker-compose down -v
   ```

2. **Ejecuta el script de limpieza**:
   ```bash
   .\clean-start-docker.ps1
   ```

## Scripts útiles

| Script | Descripción |
|--------|-------------|
| `clean-start-docker.ps1/bat` | Limpia todo y reinicia desde cero |
| `fix-jwt-restart.ps1/bat` | Aplica corrección JWT y ejecuta migraciones |
| `quick-restart.ps1/bat` | Reinicio rápido del API con nueva configuración |
| `quick-fix.bat` | Reinicio muy rápido (solo restart) |
| `restart-api.ps1/bat` | Reinicia API y ejecuta migraciones |
| `check-docker.ps1` | Verifica el estado de los servicios |
| `check-env.ps1` | Muestra las variables de entorno |
| `check-all.bat` | Verificación completa del sistema |
| `create-users.ps1/bat` | Crea usuarios de prueba manualmente |

## Verificar el estado

Ejecuta:
```bash
.\check-all.bat
```

Este script te mostrará:
- Si Docker está corriendo
- Estado de PostgreSQL
- Estado del API
- Si las tablas existen
- Si el GraphQL Playground es accesible

## Comandos útiles

### Ver logs en tiempo real
```bash
# Todos los servicios
docker-compose logs -f

# Solo API
docker-compose logs -f api

# Solo PostgreSQL
docker-compose logs -f postgres
```

### Acceder a los contenedores
```bash
# Acceder al contenedor del API
docker-compose exec api sh

# Acceder a PostgreSQL
docker-compose exec postgres psql -U postgres -d asam_db
```

### Reiniciar servicios
```bash
# Reiniciar todo
docker-compose restart

# Reiniciar solo el API
docker-compose restart api
```

### Limpiar todo
```bash
# Detener y eliminar contenedores
docker-compose down

# Detener, eliminar contenedores Y volúmenes (borra datos)
docker-compose down -v

# Eliminar todas las imágenes no utilizadas
docker system prune -a
```

## Si nada funciona

1. Reinicia Docker Desktop
2. Ejecuta: `docker system prune -a` (esto limpiará todo)
3. Vuelve a ejecutar: `.\clean-start-docker.ps1`

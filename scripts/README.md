# Scripts de Migraciones para Producción

Este directorio contiene scripts para ejecutar migraciones de base de datos en el entorno de producción.

## Requisitos previos

1. **Google Cloud CLI** instalado y configurado:
   ```bash
   # Instalar desde: https://cloud.google.com/sdk/docs/install
   
   # Autenticarse
   gcloud auth login
   
   # Configurar proyecto
   gcloud config set project YOUR_PROJECT_ID
   ```

2. **Permisos necesarios**:
   - Acceso a Google Secret Manager
   - Permisos para leer los secretos de la base de datos

3. **Go** instalado en tu máquina local

## Uso

### Windows (PowerShell)

```powershell
# Ejecutar migraciones hacia arriba (por defecto)
.\scripts\run-production-migrations.ps1

# Ver la versión actual
.\scripts\run-production-migrations.ps1 -Command version

# Revertir todas las migraciones (¡CUIDADO!)
.\scripts\run-production-migrations.ps1 -Command down
```

### Linux/Mac

```bash
# Dar permisos de ejecución
chmod +x scripts/run-production-migrations.sh

# Ejecutar migraciones hacia arriba (por defecto)
./scripts/run-production-migrations.sh

# Ver la versión actual
./scripts/run-production-migrations.sh version

# Revertir todas las migraciones (¡CUIDADO!)
./scripts/run-production-migrations.sh down
```

## Alternativas

### 1. Desde GitHub Actions (Recomendado)

Ve a GitHub → Actions → "Deploy to Google Cloud Run" y activa las migraciones:
- Environment: `production`
- Run database migrations: ✅

### 2. Con variables de entorno locales

Si tienes las credenciales directamente (no recomendado para producción):

```bash
export DB_HOST="tu-host.aivencloud.com"
export DB_PORT="14276"
export DB_USER="avnadmin"
export DB_PASSWORD="tu-password"
export DB_NAME="defaultdb"
export DB_SSL_MODE="require"

go run cmd/migrate/main.go -cmd up
```

## Seguridad

- **NUNCA** hardcodees credenciales en los scripts
- **NUNCA** commits archivos con credenciales
- Usa siempre Google Secret Manager para producción
- Los scripts limpian las variables de entorno después de usarlas

## Troubleshooting

### Error: "No se pudieron obtener todas las credenciales"

1. Verifica que estás autenticado:
   ```bash
   gcloud auth list
   ```

2. Verifica el proyecto actual:
   ```bash
   gcloud config get-value project
   ```

3. Lista los secretos disponibles:
   ```bash
   gcloud secrets list
   ```

### Error: "Error al conectar a la base de datos"

1. Verifica que tu IP está en la lista blanca de Aiven
2. Verifica que las credenciales son correctas
3. Prueba la conexión con psql:
   ```bash
   psql "postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=require"
   ```

## Mejores prácticas

1. **Siempre** ejecuta `version` antes de hacer cambios para saber el estado actual
2. **Siempre** haz backup antes de ejecutar migraciones en producción
3. **Nunca** ejecutes `down` en producción sin un plan de rollback
4. **Considera** ejecutar las migraciones durante ventanas de mantenimiento

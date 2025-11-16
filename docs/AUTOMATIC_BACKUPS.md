# Servicio de Backups Automáticos

El sistema de backups automáticos permite crear copias de seguridad de la base de datos de forma periódica sin intervención manual.

## Características

- **Backups periódicos**: Configurable (por defecto cada 12 horas)
- **Retención automática**: Mantiene solo las N copias más recientes (por defecto 2)
- **Integración con Google Drive**: Los backups de producción se sincronizan automáticamente
- **Verificación de integridad**: Valida que Docker esté disponible antes de crear backups
- **Logs detallados**: Registra cada operación de backup y limpieza

## Configuración

El servicio se configura mediante variables de entorno:

### Variables de Configuración

```bash
# Habilitar el servicio de backups
BACKUP_ENABLED=true

# Intervalo entre backups (formato: 1h, 12h, 24h, etc.)
BACKUP_INTERVAL=12h

# Directorio donde se guardarán los backups
# Para producción (sincronizado con Google Drive):
BACKUP_DIR=/Users/javierfernandezcabanas/Library/CloudStorage/GoogleDrive-javierfernandezc@gmail.com/My Drive/Babacar/asam-db-backups

# Para desarrollo local:
BACKUP_DIR=/Users/javierfernandezcabanas/asam-backups/local

# Número máximo de backups a mantener
BACKUP_MAX_RETENTION=2

# Identificador del entorno (afecta el nombre de los archivos)
BACKUP_ENVIRONMENT=production
```

## Uso

### 1. Habilitar el Servicio

Edita tu archivo `.env` y agrega/modifica:

```bash
BACKUP_ENABLED=true
BACKUP_INTERVAL=12h
BACKUP_DIR=/Users/javierfernandezcabanas/Library/CloudStorage/GoogleDrive-javierfernandezc@gmail.com/My Drive/Babacar/asam-db-backups
BACKUP_MAX_RETENTION=2
BACKUP_ENVIRONMENT=production
```

### 2. Iniciar la Aplicación

El servicio se iniciará automáticamente cuando arranque el backend:

```bash
go run cmd/api/main.go
```

Verás un mensaje en los logs:

```
Database backup service started
  interval=12h0m0s
  backup_dir=/Users/.../asam-db-backups
  max_retention=2
```

### 3. Verificar Funcionamiento

El servicio:
1. Crea un backup **inmediatamente** al iniciar
2. Luego crea backups cada 12 horas (o el intervalo configurado)
3. Mantiene solo las 2 copias más recientes (o el número configurado)
4. Elimina automáticamente los backups antiguos

## Formato de los Archivos

Los backups se crean con el siguiente formato:

```
backup_<environment>_<timestamp>.dump
```

Ejemplo:
```
backup_production_20251116_143000.dump
backup_production_20251117_023000.dump
```

## Logs

El servicio registra todas las operaciones:

```
[INFO] Starting database backup...
[INFO] Backup created successfully
  filename=backup_production_20251116_143000.dump
  duration=2.5s
[INFO] Cleaning up old backups
  max_retention=2
[INFO] Removed old backup
  file=backup_production_20251115_143000.dump
[INFO] Cleanup completed
  removed=1
  remaining=2
```

## Requisitos

- **Docker**: Debe estar corriendo (usa contenedor `postgres:17-alpine` para pg_dump)
- **Permisos de escritura**: En el directorio de backups
- **Google Drive sincronizado**: Si usas la carpeta de Google Drive

## Ventajas vs Script Manual

### Script Manual (`backup-unified.sh`)
- ✅ Control total sobre cuándo se ejecuta
- ✅ Confirmación interactiva
- ❌ Requiere intervención manual
- ❌ Puede olvidarse de ejecutar

### Servicio Automático
- ✅ **Totalmente automático**
- ✅ **No requiere intervención**
- ✅ **Se ejecuta mientras la app está corriendo**
- ✅ **Logs centralizados**
- ❌ Se detiene si la aplicación se detiene

## Solución de Problemas

### Error: "Docker is not running"

**Solución**: Inicia Docker Desktop
```bash
open -a Docker
```

### Error: "Failed to create backup directory"

**Solución**: Verifica los permisos del directorio
```bash
mkdir -p /ruta/al/directorio/backups
chmod 750 /ruta/al/directorio/backups
```

### Los backups no aparecen en Google Drive

**Solución**:
1. Verifica que Google Drive esté sincronizado
2. Comprueba que la ruta en `BACKUP_DIR` sea correcta
3. Verifica en los logs que el backup se creó exitosamente

### Backup manual de emergencia

Si necesitas un backup inmediato, aún puedes usar el script manual:

```bash
./scripts/backup-unified.sh production
```

## Configuraciones Recomendadas

### Para Producción
```bash
BACKUP_ENABLED=true
BACKUP_INTERVAL=12h
BACKUP_DIR=/Users/javierfernandezcabanas/Library/CloudStorage/GoogleDrive-javierfernandezc@gmail.com/My Drive/Babacar/asam-db-backups
BACKUP_MAX_RETENTION=2
BACKUP_ENVIRONMENT=production
```

### Para Desarrollo
```bash
BACKUP_ENABLED=false  # Normalmente deshabilitado en desarrollo
BACKUP_INTERVAL=24h
BACKUP_DIR=/Users/javierfernandezcabanas/asam-backups/local
BACKUP_MAX_RETENTION=3
BACKUP_ENVIRONMENT=development
```

### Para Testing
```bash
BACKUP_ENABLED=true
BACKUP_INTERVAL=5m   # Cada 5 minutos para testing
BACKUP_DIR=/tmp/asam-backups-test
BACKUP_MAX_RETENTION=2
BACKUP_ENVIRONMENT=test
```

## Seguridad

- Los backups contienen **toda la información** de la base de datos
- El directorio de backups debe tener **permisos restrictivos** (750)
- Los backups de producción en Google Drive tienen la **seguridad de Google**
- **No compartas** los archivos de backup públicamente

## Monitoreo

Para verificar que el servicio funciona correctamente:

1. **Revisa los logs** al iniciar la aplicación
2. **Verifica el directorio** de backups periódicamente
3. **Comprueba las timestamps** de los archivos

```bash
# Listar backups
ls -lht "/Users/.../asam-db-backups/"

# Ver los logs de la aplicación
tail -f logs/app.log | grep -i backup
```

## Restauración

Para restaurar desde un backup automático, usa el script de restore:

```bash
./scripts/restore-unified.sh production backup_production_20251116_143000.dump
```

---

**Última actualización**: 2025-11-16

# Copias de Seguridad Automáticas

El backend incluye un servicio de copias de seguridad automáticas que ejecuta `pg_dump` periódicamente contra la base de datos PostgreSQL (Aiven) y almacena los dumps en el backend de almacenamiento configurado.

## Arquitectura

El servicio de backup (`internal/domain/services/backup_service.go`) se ejecuta como una goroutine dentro del proceso principal de la API. Al arrancar:

1. Realiza un backup inmediato
2. Programa backups periódicos según el intervalo configurado
3. Limpia backups antiguos según la política de retención

Cada backup genera un archivo con formato `backup_{environment}_{timestamp}.dump` usando `pg_dump` en formato custom (`-F c`).

## Backends de almacenamiento

El servicio soporta dos backends de almacenamiento, configurados mediante la variable `BACKUP_STORAGE_TYPE`:

### Filesystem (`filesystem`)

Guarda los dumps en una ruta local. Esta ruta puede apuntar a una carpeta sincronizada con Google Drive, lo que permite tener copias de seguridad en la nube sin coste adicional.

**Variable de configuración:** `BACKUP_DIR`

Ejemplo de ruta (ajusta usuario y cuenta de Google Drive):
```
BACKUP_DIR=/Users/tu_usuario/Library/CloudStorage/GoogleDrive-tu_cuenta@gmail.com/My Drive/asam-db-backups
```

### Google Cloud Storage (`gcs`)

Sube los dumps a un bucket de Google Cloud Storage. Requiere autenticación con GCP (Application Default Credentials o cuenta de servicio).

**Variables de configuración:**
- `BACKUP_GCS_BUCKET` — nombre del bucket (por defecto: `asam-db-backups`)
- `BACKUP_GCS_PREFIX` — prefijo opcional para organizar los backups dentro del bucket

> **Nota:** Para la capa gratuita de GCS (5 GB), crear el bucket en región US.

## Variables de entorno

| Variable | Descripción | Valor por defecto |
|---|---|---|
| `BACKUP_ENABLED` | Habilitar/deshabilitar backups automáticos | `false` |
| `BACKUP_INTERVAL` | Intervalo entre backups | `24h` |
| `BACKUP_STORAGE_TYPE` | Backend de almacenamiento (`filesystem` o `gcs`) | `filesystem` |
| `BACKUP_DIR` | Ruta local para backups (solo filesystem) | Carpeta Google Drive |
| `BACKUP_GCS_BUCKET` | Nombre del bucket GCS | `asam-db-backups` |
| `BACKUP_GCS_PREFIX` | Prefijo dentro del bucket GCS | (vacío) |
| `BACKUP_MAX_RETENTION` | Numero maximo de backups a mantener (0 = deshabilitado) | `0` |
| `BACKUP_ENVIRONMENT` | Identificador de entorno en el nombre del archivo | `production` |

## Retención

Hay dos estrategias de retención:

1. **In-app** (`BACKUP_MAX_RETENTION > 0`): El servicio elimina los backups mas antiguos, manteniendo solo los N mas recientes.
2. **GCS Lifecycle Policies** (`BACKUP_MAX_RETENTION = 0`, recomendado para GCS): Se delega la limpieza a las políticas de ciclo de vida del bucket, lo que permite mayor flexibilidad (por ejemplo, mover a Nearline tras 30 días y eliminar tras 90).

## Backup manual

Ademas de los backups automáticos, el servicio expone el método `BackupNow(ctx)` para disparar un backup manualmente.

## Notas sobre Aiven

Aiven proporciona sus propios backups automáticos diarios con retención configurable según el plan contratado. Los backups descritos en este documento son **complementarios** y proporcionan una copia adicional bajo nuestro control, util para:

- Tener backups fuera de la infraestructura de Aiven
- Poder restaurar en un entorno local o en otro proveedor
- Cumplir políticas de retención propias

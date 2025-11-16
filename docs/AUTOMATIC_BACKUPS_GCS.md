# Backups Automáticos con Google Cloud Storage

Esta guía explica cómo configurar backups automáticos de la base de datos en **Google Cloud Storage (GCS)** para aplicaciones ejecutándose en Google Cloud Run, Cloud Functions, o GKE.

## ¿Por qué usar Google Cloud Storage?

Cuando tu aplicación corre en Google Cloud (no en tu Mac local), **no tiene acceso a Google Drive**. Google Cloud Storage es la solución nativa de Google Cloud para almacenar archivos:

✅ **Integración nativa** con Google Cloud Run/GKE
✅ **No requiere configuración compleja** de permisos
✅ **Más seguro y confiable** que Google Drive API
✅ **Versionado automático** de archivos
✅ **Acceso desde la consola web** de Google Cloud
✅ **Más económico** para backups

## Configuración Paso a Paso

### 1. Crear un Bucket en Google Cloud Storage

#### ⚠️ IMPORTANTE: Usar Región US para Capa Gratuita

La **capa gratuita de GCS (5 GB)** solo aplica en **regiones de Estados Unidos**.

**Regiones gratuitas:**
- `us-east1` (Carolina del Sur) ✅ Recomendada
- `us-west1` (Oregón)
- `us-central1` (Iowa)

**Regiones con costo desde el primer byte:**
- `europe-southwest1` (Madrid) ❌
- `europe-west1` (Bélgica) ❌

**Nota:** Aunque tu app corra en Europa, usar región US está bien porque:
- El backup es un proceso en background (no afecta usuarios)
- Se ejecuta solo 1 vez al día
- La latencia adicional (150-200ms) es irrelevante para backups

#### Opción A: Desde la Consola Web

1. Ve a https://console.cloud.google.com/storage
2. Click en "CREATE BUCKET"
3. Configuración:
   - **Name**: `asam-db-backups`
   - **Location type**: `Region`
   - **Location**: `us-east1` ⭐ (región gratuita)
   - **Storage class**: `Standard`
   - **Access control**: `Uniform`
   - **Protection tools**: NO habilites versionado (usaremos lifecycle)
4. Click "CREATE"

#### Opción B: Desde gcloud CLI (Recomendado)

```bash
# Crear bucket en región US gratuita
gcloud storage buckets create gs://asam-db-backups \
  --location=us-east1 \
  --uniform-bucket-level-access

# NO habilitar versionado - usaremos lifecycle policy en su lugar
```

### 2. Configurar Permisos (IAM)

La aplicación necesita permisos para escribir en el bucket:

#### Si usas Cloud Run con Service Account por defecto:

```bash
# Obtener el email del service account
gcloud run services describe asam-backend \
  --region=europe-southwest1 \
  --format='value(spec.template.spec.serviceAccountName)'

# Dar permisos al service account
gcloud storage buckets add-iam-policy-binding gs://asam-db-backups \
  --member="serviceAccount:YOUR-SERVICE-ACCOUNT@PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

#### Si usas Cloud Run con el Compute Engine default service account:

```bash
# El service account suele ser: PROJECT-NUMBER-compute@developer.gserviceaccount.com
gcloud storage buckets add-iam-policy-binding gs://asam-db-backups \
  --member="serviceAccount:PROJECT-NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

### 3. Configurar Lifecycle Policy (IMPORTANTE)

**Estrategia KISS:** Deja que GCS maneje la retención automáticamente.

#### ¿Por qué usar Lifecycle Policy?

❌ **Evitar:** Código custom que borra backups (punto de fallo, más complejidad)
✅ **Preferir:** Lifecycle nativo de GCS (confiable, probado, funciona siempre)

#### Configurar retención de 30 días

```bash
# Crear archivo lifecycle.json
cat > lifecycle.json <<'EOF'
{
  "lifecycle": {
    "rule": [
      {
        "action": {
          "type": "Delete"
        },
        "condition": {
          "age": 30,
          "matchesPrefix": ["production/backup_"]
        }
      }
    ]
  }
}
EOF

# Aplicar política al bucket
gcloud storage buckets update gs://asam-db-backups \
  --lifecycle-file=lifecycle.json
```

**Resultado:**
- GCS **automáticamente** borra backups con más de 30 días
- Siempre tienes los **últimos 30 días** de backups
- Funciona **aunque tu app esté caída**
- **0 código** que mantener

#### Verificar que la política está activa

```bash
gcloud storage buckets describe gs://asam-db-backups \
  --format="json(lifecycle)"
```

### 4. Configurar Variables de Entorno en Cloud Run

Actualiza las variables de entorno de tu servicio en Cloud Run:

```bash
gcloud run services update asam-backend \
  --region=europe-southwest1 \
  --update-env-vars="\
BACKUP_ENABLED=true,\
BACKUP_INTERVAL=24h,\
BACKUP_STORAGE_TYPE=gcs,\
BACKUP_GCS_BUCKET=asam-db-backups,\
BACKUP_GCS_PREFIX=production,\
BACKUP_MAX_RETENTION=0,\
BACKUP_ENVIRONMENT=production"
```

**Nota:** `BACKUP_MAX_RETENTION=0` desactiva la retención en código (usamos GCS lifecycle)

O desde la consola web:
1. Ve a Cloud Run → tu servicio
2. Click "EDIT & DEPLOY NEW REVISION"
3. Tab "VARIABLES & SECRETS"
4. Agrega las variables:
   ```
   BACKUP_ENABLED=true
   BACKUP_INTERVAL=24h
   BACKUP_STORAGE_TYPE=gcs
   BACKUP_GCS_BUCKET=asam-db-backups
   BACKUP_GCS_PREFIX=production
   BACKUP_MAX_RETENTION=0
   BACKUP_ENVIRONMENT=production
   ```

### 4. Desplegar y Verificar

Después de actualizar las variables de entorno, el servicio se reiniciará automáticamente.

#### Verificar en los Logs:

```bash
# Ver logs en tiempo real
gcloud run services logs tail asam-backend --region=europe-southwest1

# Buscar logs de backup
gcloud run services logs read asam-backend \
  --region=europe-southwest1 \
  --limit=50 | grep -i backup
```

Deberías ver mensajes como:

```
Connected to Google Cloud Storage
  bucket=asam-db-backups
  prefix=production

Database backup service started
  interval=1h0m0s
  storage_type=gcs
  max_retention=2

Starting database backup...
Backup file created locally
  size_bytes=1234567

Uploading backup to Google Cloud Storage...
  bucket=asam-db-backups
  object=production/backup_production_20251116_143000.dump

Backup uploaded successfully to GCS
  bytes_written=1234567
  duration=2.5s
```

#### Verificar Backups en GCS:

```bash
# Listar backups
gcloud storage ls gs://asam-db-backups/production/

# Ver detalles de un backup
gcloud storage ls -l gs://asam-db-backups/production/backup_production_*.dump
```

O desde la consola web: https://console.cloud.google.com/storage/browser/asam-db-backups

## Configuración de Variables

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `BACKUP_ENABLED` | Habilitar backups | `true` |
| `BACKUP_INTERVAL` | Intervalo entre backups | `1h` |
| `BACKUP_STORAGE_TYPE` | Tipo de almacenamiento | `gcs` |
| `BACKUP_GCS_BUCKET` | Nombre del bucket | `asam-db-backups` |
| `BACKUP_GCS_PREFIX` | Prefijo/carpeta en el bucket | `production` |
| `BACKUP_MAX_RETENTION` | Máximo de backups a mantener | `2` |
| `BACKUP_ENVIRONMENT` | Identificador del entorno | `production` |

## Estructura de Archivos en GCS

Con la configuración anterior, los backups se guardarán así:

```
gs://asam-db-backups/
└── production/
    ├── backup_production_20251116_143000.dump
    └── backup_production_20251116_153000.dump
```

Si no usas `BACKUP_GCS_PREFIX`, se guardarán en la raíz del bucket:

```
gs://asam-db-backups/
├── backup_production_20251116_143000.dump
└── backup_production_20251116_153000.dump
```

## Acceder a los Backups

### Desde la Consola Web

1. Ve a https://console.cloud.google.com/storage
2. Click en tu bucket `asam-db-backups`
3. Navega a la carpeta `production/`
4. Click en el archivo → "Download"

### Desde gcloud CLI

```bash
# Descargar un backup específico
gcloud storage cp gs://asam-db-backups/production/backup_production_20251116_143000.dump ./

# Descargar el backup más reciente
LATEST=$(gcloud storage ls gs://asam-db-backups/production/ | sort -r | head -n1)
gcloud storage cp $LATEST ./
```

### Restaurar un Backup

Una vez descargado, usa el script de restore:

```bash
# Descargar backup
gcloud storage cp gs://asam-db-backups/production/backup_production_20251116_143000.dump \
  ~/Downloads/

# Restaurar
./scripts/restore-unified.sh production backup_production_20251116_143000.dump
```

## Costos

Google Cloud Storage es muy económico para backups:

**Standard Storage (europe-southwest1)**:
- Almacenamiento: ~$0.02 USD/GB/mes
- Transferencia de salida (descarga): ~$0.12 USD/GB

**Ejemplo** con backups de 100 MB cada uno:
- 2 backups × 0.1 GB = 0.2 GB
- Costo mensual: ~$0.004 USD/mes
- Costo anual: ~$0.05 USD/año

## Seguridad

### Lifecycle Policies (Opcional)

Puedes configurar políticas para eliminar backups antiguos automáticamente:

```bash
# Crear archivo lifecycle.json
cat > lifecycle.json <<EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {
          "age": 30,
          "matchesPrefix": ["production/backup_"]
        }
      }
    ]
  }
}
EOF

# Aplicar política
gcloud storage buckets update gs://asam-db-backups \
  --lifecycle-file=lifecycle.json
```

Esto eliminará backups de más de 30 días automáticamente (además de la retención de 2 backups que ya hace la app).

### Encriptación

Por defecto, GCS encripta todos los archivos. Puedes usar:

1. **Encriptación por defecto de Google** (recomendado, sin configuración)
2. **Customer-managed encryption keys (CMEK)** para mayor control
3. **Client-side encryption** para máxima seguridad

Para la mayoría de casos, la encriptación por defecto es suficiente.

## Monitoreo

### Ver Logs de Backups

```bash
# Logs en tiempo real
gcloud run services logs tail asam-backend \
  --region=europe-southwest1 \
  --filter="textPayload:backup"

# Últimos 10 backups
gcloud storage ls -l gs://asam-db-backups/production/ | head -n 10
```

### Alertas (Opcional)

Puedes configurar alertas en Cloud Monitoring para notificarte si:
- No se ha creado un backup en las últimas 2 horas
- El tamaño del backup es anormalmente pequeño/grande
- Hay errores en los logs de backup

## Comparación: GCS vs Google Drive

| Característica | Google Cloud Storage | Google Drive |
|----------------|---------------------|---------------|
| **Integración con Cloud** | ✅ Nativa | ❌ Requiere OAuth2 |
| **Configuración** | ✅ Simple | ❌ Compleja |
| **Confiabilidad** | ✅ 99.95% SLA | ⚠️ Sin SLA |
| **Costos** | ✅ $0.02/GB/mes | ✅ 15 GB gratis |
| **Acceso programático** | ✅ API simple | ⚠️ OAuth complejo |
| **Versionado** | ✅ Built-in | ⚠️ Manual |
| **Lifecycle policies** | ✅ Automático | ❌ No disponible |
| **Logs y métricas** | ✅ Cloud Monitoring | ❌ No disponible |

## Solución de Problemas

### Error: "failed to create GCS client"

**Causa**: No hay credenciales de Google Cloud configuradas.

**Solución**: Verifica que el service account de Cloud Run tiene permisos:

```bash
gcloud storage buckets get-iam-policy gs://asam-db-backups
```

### Error: "failed to access bucket"

**Causa**: El bucket no existe o no tienes permisos.

**Solución**:
1. Verifica que el bucket existe: `gcloud storage buckets describe gs://asam-db-backups`
2. Verifica permisos (ver sección "Configurar Permisos")

### Los backups no aparecen en GCS

**Solución**:
1. Verifica los logs: `gcloud run services logs read asam-backend --region=europe-southwest1`
2. Verifica la variable `BACKUP_ENABLED=true`
3. Verifica la variable `BACKUP_STORAGE_TYPE=gcs`

### Docker no disponible en Cloud Run

**Importante**: Cloud Run tiene Docker disponible para ejecutar pg_dump. Si ves errores de Docker, contacta con soporte de Google Cloud.

## Desarrollo Local con GCS

Si quieres probar GCS en local:

```bash
# Autenticarte con gcloud
gcloud auth application-default login

# Configurar .env local
BACKUP_ENABLED=true
BACKUP_STORAGE_TYPE=gcs
BACKUP_GCS_BUCKET=asam-db-backups
BACKUP_GCS_PREFIX=development

# Ejecutar
go run cmd/api/main.go
```

Para desarrollo local, es más fácil usar `BACKUP_STORAGE_TYPE=filesystem`.

---

**Última actualización**: 2025-11-16

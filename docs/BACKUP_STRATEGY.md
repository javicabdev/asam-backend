# Estrategia de Backups (KISS)

Esta guía explica la estrategia recomendada de backups siguiendo los principios **KISS** (Keep It Simple, Stupid) y **YAGNI** (You Aren't Gonna Need It).

## ⚠️ El Problema de la Ventana de Retención

### Estrategia INCORRECTA ❌

```
Backup cada 1 hora + mantener solo 2 copias
```

**¿Por qué es peligrosa?**

Imagina este escenario:

- **1:00 PM:** Creas Backup A
- **2:00 PM:** Creas Backup B (tienes A y B)
- **3:00 PM:** Creas Backup C → **Borras A automáticamente**

A las **3:01 PM** descubres que un error grave corrompió la base de datos a las **12:30 PM**.

¿Qué copias tienes? Solo las de 2:00 PM y 3:00 PM. **Ambas están corruptas.**

**Has perdido toda tu información**, a pesar de hacer backups cada hora.

### Tu ventana de protección

Con esta estrategia solo te proteges contra errores ocurridos en las **últimas 2 horas**.

## ✅ Estrategia Recomendada (KISS)

### Principios

1. **YAGNI (Frecuencia)**: ¿Realmente necesitas un backup cada hora para una base de datos pequeña con poco movimiento?
   - **NO**. Un backup **una vez al día** es mucho más razonable.

2. **KISS (Retención)**: El problema no es el espacio (tienes 5 GB gratis en GCS).
   - Guarda **más copias**, pero espaciadas en el tiempo.
   - Usa **herramientas nativas** (GCS Lifecycle) en lugar de código custom.

### Configuración Óptima

#### Para producción (Google Cloud Storage)

```bash
# Variables de entorno
BACKUP_ENABLED=true
BACKUP_INTERVAL=24h                    # 1 backup diario
BACKUP_STORAGE_TYPE=gcs
BACKUP_GCS_BUCKET=asam-db-backups
BACKUP_GCS_PREFIX=production
BACKUP_MAX_RETENTION=0                 # Deshabilitado - usar lifecycle
BACKUP_ENVIRONMENT=production
```

#### Lifecycle Policy en GCS

Configura una regla simple en tu bucket:

1. **Acción:** Delete (Borrar)
2. **Condición:** Age (Antigüedad)
3. **Valor:** 30 días

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

### Resultado de esta estrategia

✅ **Cada día** se guarda un nuevo backup en GCS
✅ **GCS automáticamente** borra backups con más de 30 días
✅ Siempre tienes los **últimos 30 días** de backups
✅ Si descubres un error que ocurrió hace 3 semanas, **tienes la copia**

### Costos

Con backups de 5 MB cada uno:

```
5 MB/día × 30 días = 150 MB total
```

**150 MB** está muy por debajo de los **5,000 MB (5 GB)** gratuitos.

**Costo: $0 USD** 🎉

## Comparación de Estrategias

| Característica | 1 hora + 2 copias ❌ | 24 horas + 30 días ✅ |
|----------------|---------------------|---------------------|
| **Frecuencia** | Cada 1 hora | Cada 24 horas |
| **Ventana protección** | 2 horas | 30 días |
| **Número de copias** | 2 | ~30 |
| **Espacio usado** | ~10 MB | ~150 MB |
| **Protección real** | Muy baja | Alta |
| **Complejidad** | Alta (código custom) | Baja (lifecycle nativo) |
| **Costo (5MB/backup)** | $0 | $0 |

## Regiones y Capa Gratuita de GCS

### ⚠️ IMPORTANTE: Región del Bucket

La **capa gratuita de GCS (5 GB)** solo aplica en **regiones de Estados Unidos**:

✅ **Regiones gratuitas:**
- `us-east1` (Carolina del Sur)
- `us-west1` (Oregón)
- `us-central1` (Iowa)

❌ **Regiones con costo desde el primer byte:**
- `europe-southwest1` (Madrid)
- `europe-west1` (Bélgica)
- `asia-southeast1` (Singapur)

### Crear bucket en región gratuita

```bash
# ✅ Correcto - Región US gratuita
gcloud storage buckets create gs://asam-db-backups \
  --location=us-east1 \
  --uniform-bucket-level-access

# ❌ Incorrecto - Región EU con costo
gcloud storage buckets create gs://asam-db-backups \
  --location=europe-southwest1 \
  --uniform-bucket-level-access
```

### ¿Latencia desde Europa?

Aunque tu app corra en `europe-southwest1` (Madrid), **no hay problema**:

- El backup es un proceso en **background** (no afecta a usuarios)
- Se ejecuta **una vez al día** (no es crítico)
- La latencia adicional (150-200ms) es **irrelevante** para backups

**Ahorro vs latencia:** Preferir región US gratuita es la opción correcta.

## Configuración Completa Paso a Paso

### 1. Crear bucket en región US

```bash
gcloud storage buckets create gs://asam-db-backups \
  --location=us-east1 \
  --uniform-bucket-level-access
```

### 2. Configurar lifecycle policy

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

# Aplicar
gcloud storage buckets update gs://asam-db-backups \
  --lifecycle-file=lifecycle.json
```

### 3. Dar permisos al service account

```bash
gcloud storage buckets add-iam-policy-binding gs://asam-db-backups \
  --member="serviceAccount:YOUR-SERVICE-ACCOUNT@PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

### 4. Configurar variables en Cloud Run

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

## Verificación

### Ver lifecycle policy activa

```bash
gcloud storage buckets describe gs://asam-db-backups \
  --format="json(lifecycle)"
```

### Listar backups actuales

```bash
gcloud storage ls -l gs://asam-db-backups/production/
```

### Ver logs de backup en Cloud Run

```bash
gcloud run services logs tail asam-backend \
  --region=europe-southwest1 \
  --filter="textPayload:backup"
```

## Preguntas Frecuentes

### ¿Por qué no hacer backups más frecuentes?

Para una base de datos pequeña (~10k registros) con movimiento moderado:
- **1 backup/día** es suficiente
- Backups más frecuentes solo añaden complejidad sin beneficio real
- **YAGNI** (You Aren't Gonna Need It)

### ¿Qué pasa si necesito recuperar datos de hace 31 días?

No podrás. Pero pregúntate:
- ¿Cuándo ha pasado esto en los últimos 2 años?
- ¿Vale la pena la complejidad de guardar backups de 1 año?
- **YAGNI** - Si nunca ha pasado, probablemente nunca pasará

Si realmente necesitas retención de 1 año:
- Cambia `"age": 30` → `"age": 365`
- Espacio: 5 MB × 365 = ~1.8 GB (aún dentro de 5 GB gratuitos)

### ¿Por qué usar GCS lifecycle en lugar de código?

**Principio KISS:**
- GCS lifecycle es **nativo**, **confiable**, **probado**
- Tu código custom es **un punto más de fallo**
- GCS lifecycle funciona **aunque tu app esté caída**
- Menos código = menos bugs

---

**Principios aplicados:** KISS, YAGNI, DRY

**Última actualización:** 2025-11-16

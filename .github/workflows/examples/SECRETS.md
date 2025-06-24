# Configuración de Secrets

Este documento explica cómo configurar los secrets necesarios para el despliegue del ASAM Backend.

## 🔐 Arquitectura de Secrets

El proyecto utiliza **Google Secret Manager** para almacenar información sensible de forma segura. Los secrets se referencian en Cloud Run pero nunca se exponen en el código o en los logs.

## 📋 Secrets Requeridos

### En GitHub Actions

Solo necesitas configurar estos dos secrets en GitHub:

1. **`GCP_PROJECT_ID`**: ID de tu proyecto en Google Cloud Platform
2. **`GCP_SA_KEY`**: JSON de la cuenta de servicio con permisos para:
   - Cloud Run Admin
   - Artifact Registry Writer
   - Service Account User
   - Secret Manager Viewer (para las migraciones)

### En Google Secret Manager

Todos los demás secrets se almacenan en Google Secret Manager:

#### Base de Datos (Aiven PostgreSQL)
- `db-host`: Host de la base de datos
- `db-port`: Puerto (generalmente 14276 para Aiven)
- `db-user`: Usuario de la base de datos
- `db-password`: Contraseña de la base de datos
- `db-name`: Nombre de la base de datos

#### Autenticación JWT
- `jwt-access-secret`: Secret para tokens de acceso (generado automáticamente)
- `jwt-refresh-secret`: Secret para refresh tokens (generado automáticamente)

#### Credenciales de Admin
- `admin-user`: Usuario administrador inicial
- `admin-password`: Contraseña del administrador

#### SMTP (Opcional)
- `smtp-user`: Usuario SMTP (email)
- `smtp-password`: Contraseña de aplicación SMTP

## 🚀 Configuración Paso a Paso

### 1. Crear Cuenta de Servicio en GCP

```bash
# Crear cuenta de servicio
gcloud iam service-accounts create asam-backend-sa \
  --display-name="ASAM Backend Service Account" \
  --project=TU_PROJECT_ID

# Asignar roles necesarios
gcloud projects add-iam-policy-binding TU_PROJECT_ID \
  --member="serviceAccount:asam-backend-sa@TU_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding TU_PROJECT_ID \
  --member="serviceAccount:asam-backend-sa@TU_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding TU_PROJECT_ID \
  --member="serviceAccount:asam-backend-sa@TU_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding TU_PROJECT_ID \
  --member="serviceAccount:asam-backend-sa@TU_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.viewer"

# Crear y descargar key JSON
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=asam-backend-sa@TU_PROJECT_ID.iam.gserviceaccount.com
```

### 2. Configurar Secrets en GitHub

1. Ve a tu repositorio en GitHub
2. Settings > Secrets and variables > Actions
3. Crea los siguientes secrets:
   - `GCP_PROJECT_ID`: Tu ID de proyecto
   - `GCP_SA_KEY`: Contenido del archivo `sa-key.json`

### 3. Configurar Secrets en Google Secret Manager

#### Opción A: Usando PowerShell (Windows)

```powershell
# Establecer proyecto
$env:PROJECT_ID = "TU_PROJECT_ID"

# Ejecutar script
.\scripts\create-secrets.ps1
```

#### Opción B: Usando Bash (Linux/Mac)

```bash
# Establecer proyecto
export PROJECT_ID="TU_PROJECT_ID"

# Ejecutar script
chmod +x scripts/create-secrets.sh
./scripts/create-secrets.sh
```

#### Opción C: Manualmente

```bash
# Habilitar API
gcloud services enable secretmanager.googleapis.com

# Crear cada secret
echo -n "valor-del-secret" | gcloud secrets create nombre-del-secret --data-file=-

# Ejemplo para la base de datos
echo -n "pg-asam-asam-backend-db.l.aivencloud.com" | gcloud secrets create db-host --data-file=-
echo -n "14276" | gcloud secrets create db-port --data-file=-
# ... continuar con los demás
```

## 🔍 Verificar Configuración

### Listar secrets creados

```bash
gcloud secrets list --project=TU_PROJECT_ID
```

### Verificar permisos de Cloud Run

```bash
# Obtener cuenta de servicio de Cloud Run
SERVICE_ACCOUNT=$(gcloud iam service-accounts list \
  --filter="displayName:Compute Engine default service account" \
  --format="value(email)")

# Verificar permisos en un secret
gcloud secrets get-iam-policy db-password --project=TU_PROJECT_ID
```

### Test de conexión local

```bash
# Exportar secrets para prueba local
export DB_HOST=$(gcloud secrets versions access latest --secret=db-host)
export DB_PORT=$(gcloud secrets versions access latest --secret=db-port)
export DB_USER=$(gcloud secrets versions access latest --secret=db-user)
export DB_PASSWORD=$(gcloud secrets versions access latest --secret=db-password)
export DB_NAME=$(gcloud secrets versions access latest --secret=db-name)

# Ejecutar test
go run cmd/test-db-connection/main.go
```

## 🚨 Seguridad

1. **Nunca** incluyas valores de secrets en el código
2. **Nunca** hagas commit de archivos `.env` con valores reales
3. **Siempre** usa `--set-secrets` en Cloud Run, no `--set-env-vars` para información sensible
4. **Rota** los secrets periódicamente, especialmente JWT secrets
5. **Limita** el acceso a Secret Manager solo a las cuentas de servicio necesarias

## 📝 Notas Importantes

- Los secrets JWT se generan automáticamente con valores aleatorios fuertes
- La contraseña de admin debe ser cambiada después del primer login
- Los secrets SMTP son opcionales; si no se configuran, las notificaciones estarán deshabilitadas
- Cloud Run cachea los secrets, los cambios pueden tardar unos minutos en aplicarse

## 🆘 Troubleshooting

### Error: "Permission denied accessing secret"

```bash
# Otorgar permisos a la cuenta de servicio de Cloud Run
gcloud secrets add-iam-policy-binding SECRET_NAME \
  --member="serviceAccount:SERVICE_ACCOUNT_EMAIL" \
  --role="roles/secretmanager.secretAccessor"
```

### Error: "Secret not found"

Verifica que el nombre del secret en Cloud Run coincida exactamente con el nombre en Secret Manager.

### Ver logs de Cloud Run

```bash
gcloud run services logs read asam-backend --region=europe-west1 --limit=50
```

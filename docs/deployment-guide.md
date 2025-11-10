# Guía de Despliegue a Producción

Esta guía detalla el proceso completo para desplegar ASAM Backend en Google Cloud Run.

## 📋 Pre-requisitos

1. **Google Cloud Platform**
   - Proyecto de GCP creado
   - Billing habilitado
   - gcloud CLI instalado y configurado

2. **Base de Datos**
   - PostgreSQL en Aiven configurado
   - Credenciales de acceso disponibles

3. **Herramientas locales**
   - Go 1.24+
   - Git
   - PowerShell (Windows) o Bash (Linux/Mac)

## 🚀 Proceso de Despliegue

### Paso 1: Preparar el proyecto GCP

```bash
# Establecer proyecto
gcloud config set project TU_PROJECT_ID

# Habilitar APIs necesarias
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com

# Crear Artifact Registry
gcloud artifacts repositories create asam-backend \
  --repository-format=docker \
  --location=europe-west1 \
  --description="ASAM Backend Docker images"
```

### Paso 2: Configurar cuenta de servicio

```bash
# Crear cuenta de servicio
gcloud iam service-accounts create asam-backend-sa \
  --display-name="ASAM Backend Service Account"

# Asignar roles
PROJECT_ID=$(gcloud config get-value project)
SA_EMAIL="asam-backend-sa@${PROJECT_ID}.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/secretmanager.viewer"

# Crear key JSON
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=${SA_EMAIL}
```

### Paso 3: Configurar GitHub Secrets

1. Ve a **Settings > Secrets and variables > Actions**
2. Crea estos secrets:
   - `GCP_PROJECT_ID`: Tu ID de proyecto
   - `GCP_SA_KEY`: Contenido de `sa-key.json`

### Paso 4: Configurar Google Secret Manager

#### Windows (PowerShell):
```powershell
$env:PROJECT_ID = "TU_PROJECT_ID"
.\scripts\create-secrets.ps1
```

#### Linux/Mac (Bash):
```bash
export PROJECT_ID="TU_PROJECT_ID"
chmod +x scripts/create-secrets.sh
./scripts/create-secrets.sh
```

### Paso 5: Ejecutar migraciones iniciales

```bash
# Obtener credenciales de la base de datos
export DB_HOST=$(gcloud secrets versions access latest --secret=db-host)
export DB_PORT=$(gcloud secrets versions access latest --secret=db-port)
export DB_USER=$(gcloud secrets versions access latest --secret=db-user)
export DB_PASSWORD=$(gcloud secrets versions access latest --secret=db-password)
export DB_NAME=$(gcloud secrets versions access latest --secret=db-name)
export DB_SSL_MODE=require

# Ejecutar migraciones
go run cmd/migrate/main.go -cmd up
```

### Paso 6: Desplegar aplicación

#### Opción A: Despliegue manual desde GitHub Actions

1. Ve a **Actions** en tu repositorio
2. Selecciona **Deploy to Google Cloud Run**
3. Click en **Run workflow**
4. Selecciona opciones:
   - Environment: `production`
   - Run migrations: `false` (ya las ejecutamos)
5. Click **Run workflow**

#### Opción B: Despliegue automático con tag

```bash
# Crear tag de versión
git tag v1.0.0
git push origin v1.0.0
```

### Paso 7: Verificar despliegue

```bash
# Obtener URL del servicio
SERVICE_URL=$(gcloud run services describe asam-backend \
  --region=europe-west1 \
  --format='value(status.url)')

# Test de salud
curl $SERVICE_URL/health

# Ver logs
gcloud run services logs read asam-backend \
  --region=europe-west1 \
  --limit=50
```

## 🔍 Monitoreo Post-Despliegue

### Métricas en Cloud Console

1. Ve a [Cloud Run](https://console.cloud.google.com/run)
2. Selecciona tu servicio
3. Revisa:
   - Request count
   - Latency
   - CPU utilization
   - Memory utilization

### Alertas recomendadas

```bash
# Crear alerta para errores 5xx
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="ASAM Backend - High Error Rate" \
  --condition-display-name="5xx errors > 1%" \
  --condition-filter='resource.type="cloud_run_revision" AND metric.type="run.googleapis.com/request_count" AND metric.labels.response_code_class="5xx"'
```

## 📝 Checklist de Producción

- [ ] Migraciones ejecutadas exitosamente
- [ ] Health check respondiendo OK
- [ ] GraphQL playground accesible (si está habilitado)
- [ ] Logs sin errores críticos
- [ ] Métricas de CPU < 80%
- [ ] Métricas de memoria < 80%
- [ ] Latencia < 500ms p95
- [ ] Usuario admin puede hacer login
- [ ] Backup de base de datos configurado

## 🆘 Troubleshooting

### La aplicación no arranca

```bash
# Ver logs detallados
gcloud run services logs read asam-backend \
  --region=europe-west1 \
  --limit=100 \
  --format="table(timestamp,severity,textPayload)"
```

### Error de conexión a base de datos

1. Verificar secrets:
   ```bash
   gcloud secrets versions access latest --secret=db-host
   gcloud secrets versions access latest --secret=db-password
   ```

2. Test de conexión local:
   ```bash
   go run cmd/test-db-connection/main.go
   ```

### Performance issues

1. Escalar instancias:
   ```bash
   gcloud run services update asam-backend \
     --region=europe-west1 \
     --min-instances=1 \
     --max-instances=10
   ```

2. Aumentar recursos:
   ```bash
   gcloud run services update asam-backend \
     --region=europe-west1 \
     --cpu=2 \
     --memory=1Gi
   ```

## 🔄 Actualizaciones

Para actualizar la aplicación:

1. Hacer cambios en el código
2. Commit y push a `main`
3. Crear nuevo tag:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```
4. El workflow de release se ejecutará automáticamente

## 🔐 Seguridad

1. **Rotar secrets periódicamente**:
   ```bash
   # Actualizar un secret
   echo -n "nuevo-valor" | gcloud secrets versions add SECRET_NAME --data-file=-
   ```

2. **Revisar permisos IAM**:
   ```bash
   gcloud projects get-iam-policy PROJECT_ID
   ```

3. **Habilitar auditoría**:
   ```bash
   gcloud run services update asam-backend \
     --region=europe-west1 \
     --update-labels=audit=enabled
   ```

## 📞 Soporte

Para problemas o dudas:
1. Revisar logs en Cloud Console
2. Consultar documentación en `/docs`
3. Abrir issue en GitHub

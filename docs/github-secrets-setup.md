# Configuración de Secretos para CI/CD en GitHub Actions

Los secretos viven en dos sitios:

- **GitHub (Actions secrets):** solo lo que los workflows necesitan para autenticarse en GCP y en la API de Aiven.
- **GCP Secret Manager:** los secretos de la aplicación (base de datos, JWT, administrador, email). El despliegue los inyecta en Cloud Run; no son secretos de GitHub.

## Secretos de GitHub

### Google Cloud Platform
- `GCP_PROJECT_ID`: El ID de tu proyecto de GCP (el nuevo proyecto creado para asam-backend).
- `GCP_SA_KEY`: La clave JSON de la cuenta de servicio de GCP con permisos para desplegar en Cloud Run, acceder a Container Registry y leer Secret Manager.

### Keep-alive del plan gratuito de Aiven
Los usa el workflow [`aiven-keepalive.yml`](../.github/workflows/aiven-keepalive.yml), que consulta la API de Aiven dos veces por semana para que el servicio gratuito no se apague por inactividad, y falla (GitHub te avisa por email) si el servicio está apagado o el token deja de valer.
- `AIVEN_API_TOKEN`: token personal de Aiven. Créalo en la consola de Aiven: icono de usuario → **Tokens** → **Generate token**. Si el formulario lo permite, dale caducidad con renovación por uso ("extend when used") y solo permisos de lectura. Se muestra una sola vez: pégalo directamente aquí, en ningún otro sitio.
- `AIVEN_PROJECT`: nombre del proyecto de Aiven.
- `AIVEN_SERVICE`: nombre del servicio PostgreSQL en ese proyecto.

Proyecto y servicio van como secretos porque juntos revelan el host de la base de datos. Tras crearlos, lanza el workflow a mano (**Actions → Aiven keep-alive → Run workflow**): el log debe mostrar solo `Service state: RUNNING` (o `REBALANCING`, transitorio).

## Secretos de la aplicación (GCP Secret Manager)

Viven en Secret Manager, en el proyecto de GCP. [`cloud-run-deploy.yml`](../.github/workflows/cloud-run-deploy.yml) los monta en Cloud Run como variables de entorno (`--set-secrets`, versión `latest`), y su job de migraciones los lee con `gcloud secrets versions access`.

| Secreto en Secret Manager | Variable de entorno en la app |
|---|---|
| `db-host` | `DB_HOST` |
| `db-port` | `DB_PORT` |
| `db-user` | `DB_USER` |
| `db-password` | `DB_PASSWORD` |
| `db-name` | `DB_NAME` |
| `jwt-access-secret` | `JWT_ACCESS_SECRET` |
| `jwt-refresh-secret` | `JWT_REFRESH_SECRET` |
| `admin-user` | `ADMIN_USER` |
| `admin-password` | `ADMIN_PASSWORD` |
| `mailersend-api-key` | `MAILERSEND_API_KEY` |

El resto de la configuración de producción no es secreta y va como `--set-env-vars` en el mismo workflow (por ejemplo `DB_SSL_MODE=require` o `DB_MAX_OPEN_CONNS`).

Para cambiar un valor (por ejemplo, al rotar la contraseña de la base de datos), añade una versión nueva sin que el valor quede en pantalla ni en el historial del shell:

```sh
read -rs VALUE && printf '%s' "$VALUE" | gcloud secrets versions add db-password --data-file=- ; unset VALUE
```

Después vuelve a desplegar con `cloud-run-deploy.yml`, para que todas las instancias arranquen con la versión nueva.

## Cómo configurar los secretos en GitHub

1. Ve a tu repositorio en GitHub
2. Haz clic en "Settings" (Configuración)
3. En el menú lateral, selecciona "Secrets and variables" > "Actions"
4. Haz clic en "New repository secret"
5. Añade cada secreto con su nombre y valor correspondiente

## Variables del repositorio (opcional, no sensibles)

En **Settings → Secrets and variables → Actions → Variables** puedes definir valores que no son secretos pero que los workflows necesitan, por ejemplo:

- **`DAST_TARGET_URL`**: URL base del API desplegado para el workflow [DAST](../.github/workflows/dast.yml) (escaneo programado o por defecto). Sin esta variable, el escaneo automático fallará hasta que la configures o pases `target_url` al ejecutar el workflow a mano.

## Crear una cuenta de servicio en Google Cloud Platform

Para obtener el archivo de clave JSON necesario para el secreto `GCP_SA_KEY`:

1. Ve a la consola de Google Cloud: https://console.cloud.google.com/
2. Asegúrate de seleccionar el nuevo proyecto que creaste para asam-backend
3. Ve a "IAM & Admin" > "Service Accounts"
4. Crea una nueva cuenta de servicio con un nombre descriptivo (ej. "github-actions-deploy")
5. Asigna los siguientes roles:
   - Cloud Run Admin
   - Cloud Build Service Account
   - Service Account User
   - Storage Admin (para acceder a Container Registry)
   - Secret Manager Secret Accessor (el job de migraciones lee los secretos de la base de datos)
6. Crea una clave para esta cuenta de servicio (JSON)
7. Descarga el archivo JSON 
8. Copia todo el contenido del archivo y pégalo como valor del secreto `GCP_SA_KEY`

## Configuración adicional en Google Cloud

Asegúrate de tener habilitadas las siguientes APIs en tu proyecto:
- Cloud Run API
- Cloud Build API
- Container Registry API
- Cloud Resource Manager API
- Secret Manager API

La cuenta de servicio con la que se ejecuta Cloud Run también necesita el rol **Secret Manager Secret Accessor**, para resolver los `--set-secrets` al arrancar.

Puedes habilitarlas en: https://console.cloud.google.com/apis/dashboard

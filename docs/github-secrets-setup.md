# Configuración de Secretos para CI/CD en GitHub Actions

Para que el flujo de trabajo de CI/CD funcione correctamente, necesitas configurar los siguientes secretos en tu repositorio de GitHub:

## Secretos necesarios

### Google Cloud Platform
- `GCP_PROJECT_ID`: El ID de tu proyecto de GCP (el nuevo proyecto creado para asam-backend).
- `GCP_SA_KEY`: La clave JSON de la cuenta de servicio de GCP con permisos para desplegar en Cloud Run y acceder a Container Registry.

### Base de datos Aiven
- `AIVEN_DB_HOST`: Hostname del servidor PostgreSQL en Aiven (ej. pg-asam-asam-backend-db.l.aivencloud.com)
- `AIVEN_DB_PORT`: Puerto del servidor PostgreSQL (ej. 14276)
- `AIVEN_DB_USER`: Usuario para conectarse a PostgreSQL (ej. avnadmin)
- `AIVEN_DB_PASSWORD`: Contraseña para conectarse a PostgreSQL
- `AIVEN_DB_NAME`: Name de la base de datos (ej. asam-backend-db)

### Seguridad de la aplicación
- `JWT_ACCESS_SECRET`: Clave secreta para la generación de tokens JWT de acceso
- `JWT_REFRESH_SECRET`: Clave secreta para la generación de tokens JWT de actualización
- `ADMIN_USER`: Usuario administrador para acceder a endpoints de monitoreo en producción
- `ADMIN_PASSWORD`: Contraseña del administrador (debe ser muy segura)

### Configuración de Email (opcional)
- `SMTP_SERVER`: Servidor SMTP para envío de correos
- `SMTP_PORT`: Puerto del servidor SMTP
- `SMTP_USER`: Usuario SMTP
- `SMTP_PASSWORD`: Contraseña SMTP

## Cómo configurar los secretos en GitHub

1. Ve a tu repositorio en GitHub
2. Haz clic en "Settings" (Configuración)
3. En el menú lateral, selecciona "Secrets and variables" > "Actions"
4. Haz clic en "New repository secret"
5. Añade cada secreto con su nombre y valor correspondiente

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
6. Crea una clave para esta cuenta de servicio (JSON)
7. Descarga el archivo JSON 
8. Copia todo el contenido del archivo y pégalo como valor del secreto `GCP_SA_KEY`

## Configuración adicional en Google Cloud

Asegúrate de tener habilitadas las siguientes APIs en tu proyecto:
- Cloud Run API
- Cloud Build API
- Container Registry API
- Cloud Resource Manager API

Puedes habilitarlas en: https://console.cloud.google.com/apis/dashboard

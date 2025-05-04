# Guía para crear y configurar un nuevo proyecto en Google Cloud Platform

Esta guía te ayudará a crear un nuevo proyecto en Google Cloud Platform (GCP) específicamente para desplegar asam-backend en Cloud Run y conectarlo a tu base de datos PostgreSQL en Aiven.

## 1. Crear un nuevo proyecto en GCP

1. Ve a la consola de Google Cloud: https://console.cloud.google.com/
2. Haz clic en el selector de proyectos en la parte superior
3. Selecciona "Nuevo proyecto"
4. En la ventana "Nuevo proyecto":
   - **Nombre del proyecto**: asam-backend (o el nombre que prefieras)
   - **Organización**: Selecciona la organización si aplica, o deja en "Sin organización"
   - **Ubicación**: Deja el valor predeterminado o selecciona una carpeta específica
5. Haz clic en "Crear"
6. Espera mientras se crea el proyecto (puede tomar unos segundos)
7. Una vez creado, serás redirigido al dashboard del nuevo proyecto

## 2. Habilitar las APIs necesarias

Para que Cloud Run funcione correctamente, necesitas habilitar varias APIs:

1. En la consola de GCP, ve a "APIs y servicios" > "Biblioteca"
2. Busca y habilita las siguientes APIs (una por una):
   - **Cloud Run API**
   - **Cloud Build API**
   - **Container Registry API**
   - **Cloud Resource Manager API**
   - **Secret Manager API** (opcional, si planeas usar Secret Manager para gestionar secretos)

Para cada API:
1. Busca el nombre de la API en la barra de búsqueda
2. Haz clic en el resultado correspondiente
3. Haz clic en "Habilitar"
4. Espera a que se complete la activación

## 3. Crear una cuenta de servicio

Para que GitHub Actions pueda desplegar en tu proyecto de GCP, necesitas crear una cuenta de servicio con los permisos adecuados:

1. En la consola de GCP, ve a "IAM y administración" > "Cuentas de servicio"
2. Haz clic en "Crear cuenta de servicio"
3. En "Detalles de la cuenta de servicio":
   - **Nombre de la cuenta de servicio**: github-actions-deploy
   - **ID de la cuenta de servicio**: se generará automáticamente
   - **Descripción**: Cuenta de servicio para GitHub Actions CI/CD
4. Haz clic en "Crear y continuar"
5. En "Otorgar acceso a esta cuenta de servicio al proyecto", asigna los siguientes roles:
   - **Cloud Run Admin** (roles/run.admin)
   - **Cloud Build Service Account** (roles/cloudbuild.builds.builder)
   - **Service Account User** (roles/iam.serviceAccountUser)
   - **Storage Admin** (roles/storage.admin)
6. Haz clic en "Continuar"
7. En "Otorgar a los usuarios acceso a esta cuenta de servicio", puedes dejarlo vacío
8. Haz clic en "Listo"

## 4. Crear una clave para la cuenta de servicio

Para utilizar la cuenta de servicio en GitHub Actions, necesitas crear una clave:

1. En la lista de cuentas de servicio, busca la cuenta que acabas de crear
2. Haz clic en los tres puntos verticales en la columna "Acciones"
3. Selecciona "Administrar claves"
4. Haz clic en "Añadir clave" > "Crear nueva clave"
5. Selecciona "JSON" como tipo de clave
6. Haz clic en "Crear"
7. El archivo de clave se descargará automáticamente a tu computadora
8. Guarda este archivo en un lugar seguro, lo necesitarás para configurar los secretos en GitHub

## 5. Obtener el ID del proyecto

Necesitarás el ID del proyecto para configurarlo en GitHub Actions:

1. En la consola de GCP, ve al "Dashboard" del proyecto
2. Busca el "ID del proyecto" en la tarjeta de información del proyecto
3. Copia este ID, lo necesitarás para configurar el secreto `GCP_PROJECT_ID` en GitHub

## 6. Configurar los secretos en GitHub

Sigue las instrucciones en el archivo [github-secrets-setup.md](github-secrets-setup.md) para configurar los secretos en GitHub, utilizando:
- El ID del proyecto que acabas de copiar como valor para `GCP_PROJECT_ID`
- El contenido completo del archivo JSON de la clave de cuenta de servicio como valor para `GCP_SA_KEY`

## Nota sobre facturación

Para utilizar Cloud Run y otras funcionalidades de Google Cloud Platform, necesitas tener la facturación habilitada en tu proyecto. Si aún no has configurado la facturación:

1. En la consola de GCP, ve a "Facturación"
2. Asocia una cuenta de facturación a tu proyecto
3. Si no tienes una cuenta de facturación, tendrás que crear una y proporcionar la información de pago

Google Cloud ofrece una capa gratuita que incluye cierto nivel de uso de Cloud Run sin costo. Consulta la [documentación de precios de Cloud Run](https://cloud.google.com/run/pricing) para más detalles.
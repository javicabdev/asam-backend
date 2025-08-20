# Scripts de Configuración para Google Cloud Platform

Este directorio contiene scripts para ayudar con la configuración y verificación de secretos en Google Secret Manager para el proyecto ASAM Backend.

## pre-deploy-check.ps1 / pre-deploy-check.sh

Script de diagnóstico que verifica que todo esté configurado correctamente antes de ejecutar el workflow de despliegue.

### Uso

#### PowerShell (Windows)

```powershell
# Verificar configuración
.\pre-deploy-check.ps1 -ProjectId tu-project-id
```

#### Bash (Linux/Mac)

```bash
# Hacer el script ejecutable (solo la primera vez)
chmod +x pre-deploy-check.sh

# Verificar configuración
./pre-deploy-check.sh tu-project-id
```

### Qué verifica

1. **Google Cloud SDK**: Instalación y versión
2. **Autenticación**: Que estés autenticado en GCP
3. **Proyecto**: Que el proyecto esté configurado correctamente
4. **APIs**: Que las APIs necesarias estén habilitadas
5. **Cuenta de servicio**: Existencia y permisos correctos
6. **Secretos de BD**: Que todos los secretos necesarios existan
7. **Otros secretos**: JWT, admin credentials, etc.
8. **Imágenes Docker**: Qué imágenes están disponibles

## verify-db-secrets.ps1 / verify-db-secrets.sh

Scripts para verificar, crear y probar los secretos de base de datos necesarios para las migraciones en Google Secret Manager.

### Requisitos previos

- Google Cloud SDK instalado y configurado
- Acceso al proyecto de GCP con permisos para:
  - Leer y crear secretos (Secret Manager)
  - Gestionar cuentas de servicio (IAM)
- (Opcional) PostgreSQL client para probar conexiones

### Uso

#### PowerShell (Windows)

```powershell
# Verificar secretos existentes
.\verify-db-secrets.ps1 -ProjectId tu-project-id

# Crear secretos faltantes
.\verify-db-secrets.ps1 -ProjectId tu-project-id -CreateSecrets

# Probar conexión a la base de datos
.\verify-db-secrets.ps1 -ProjectId tu-project-id -TestConnection

# Todas las opciones juntas
.\verify-db-secrets.ps1 -ProjectId tu-project-id -CreateSecrets -TestConnection
```

#### Bash (Linux/Mac)

```bash
# Hacer el script ejecutable (solo la primera vez)
chmod +x verify-db-secrets.sh

# Verificar secretos existentes
./verify-db-secrets.sh tu-project-id

# Crear secretos faltantes
./verify-db-secrets.sh tu-project-id --create-secrets

# Probar conexión a la base de datos
./verify-db-secrets.sh tu-project-id --test-connection

# Ver ayuda
./verify-db-secrets.sh --help
```

### Funcionalidades

1. **Verificación de secretos**: Comprueba que todos los secretos necesarios existan en Google Secret Manager:
   - `db-host`: Host de la base de datos
   - `db-port`: Puerto de la base de datos
   - `db-user`: Usuario de la base de datos
   - `db-password`: Contraseña de la base de datos
   - `db-name`: Nombre de la base de datos

2. **Creación de secretos**: Si faltan secretos, el script puede crearlos interactivamente

3. **Verificación de permisos**: Comprueba que la cuenta de servicio `github-actions-deploy` tenga los permisos necesarios

4. **Test de conexión**: Si tienes `psql` instalado, puede probar la conexión a la base de datos

### Secretos requeridos

Los siguientes secretos deben estar configurados en Google Secret Manager para que las migraciones funcionen correctamente:

| Secreto | Descripción | Ejemplo |
|---------|-------------|---------|
| `db-host` | Host de PostgreSQL | `pg-asam-xxxx.aivencloud.com` |
| `db-port` | Puerto de PostgreSQL | `14276` |
| `db-user` | Usuario de PostgreSQL | `avnadmin` |
| `db-password` | Contraseña de PostgreSQL | `contraseña-segura` |
| `db-name` | Nombre de la base de datos | `defaultdb` |

### Solución de problemas

#### Error: "La cuenta de servicio no existe"

Crea la cuenta de servicio ejecutando:

```bash
gcloud iam service-accounts create github-actions-deploy \
  --display-name='GitHub Actions Deploy Service Account'
```

#### Error: "La cuenta de servicio NO tiene el rol secretmanager.secretAccessor"

Otorga los permisos necesarios:

```bash
gcloud projects add-iam-policy-binding TU-PROJECT-ID \
  --member='serviceAccount:github-actions-deploy@TU-PROJECT-ID.iam.gserviceaccount.com' \
  --role='roles/secretmanager.secretAccessor'
```

#### Error al conectar a la base de datos

1. Verifica que los valores de los secretos sean correctos
2. Asegúrate de que la IP desde donde ejecutas el script esté autorizada en Aiven
3. Verifica que el SSL mode sea el correcto (generalmente `require` para Aiven)

### Actualizar un secreto existente

Si necesitas actualizar un secreto (por ejemplo, después de cambiar la contraseña):

```bash
# Actualizar un secreto
echo "nuevo-valor" | gcloud secrets versions add nombre-del-secreto --data-file=-

# Ejemplo: actualizar la contraseña
echo "nueva-contraseña" | gcloud secrets versions add db-password --data-file=-
```

### Ver el valor de un secreto

Para ver el valor actual de un secreto:

```bash
gcloud secrets versions access latest --secret=nombre-del-secreto

# Ejemplo: ver el host actual
gcloud secrets versions access latest --secret=db-host
```

## Integración con GitHub Actions

Estos secretos son utilizados por el workflow `cloud-run-deploy.yml` para ejecutar las migraciones de base de datos cuando se despliega la aplicación en Google Cloud Run.

El workflow:
1. Obtiene los secretos desde Google Secret Manager
2. Los exporta como variables de entorno con dos prefijos (`DB_` y `POSTGRES_`)
3. Ejecuta las migraciones usando estas variables

## Seguridad

- Los secretos están almacenados de forma segura en Google Secret Manager
- Solo las cuentas de servicio autorizadas pueden acceder a los secretos
- Los valores sensibles (como contraseñas) nunca se muestran en logs
- GitHub Actions usa máscaras para ocultar valores sensibles en los logs

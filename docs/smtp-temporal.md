# Configuración SMTP Temporal para Testing

Mientras registras el dominio, puedes usar estas opciones gratuitas para desarrollo:

## Opción 1: Mailtrap (Recomendado para desarrollo)
- Crea cuenta gratis en https://mailtrap.io
- Te dan credenciales SMTP de prueba
- Los emails no se envían realmente, pero puedes verlos en su interfaz

```env
SMTP_SERVER=smtp.mailtrap.io
SMTP_PORT=587
SMTP_USER=tu-usuario-mailtrap
SMTP_PASSWORD=tu-password-mailtrap
SMTP_USE_TLS=true
SMTP_FROM_EMAIL=noreply@asam.test
```

## Opción 2: Gmail (Para pruebas rápidas)
- Usa una cuenta Gmail personal
- Necesitas crear una "App Password"

1. Ve a https://myaccount.google.com/security
2. Activa verificación en 2 pasos
3. Genera una "App Password"

```env
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu-email@gmail.com
SMTP_PASSWORD=tu-app-password-16-caracteres
SMTP_USE_TLS=true
SMTP_FROM_EMAIL=tu-email@gmail.com
```

## Configurar en Secret Manager

```powershell
# PowerShell
$env:PROJECT_ID = "tu-proyecto-id"

# Crear secrets SMTP temporales
"smtp.gmail.com" | gcloud secrets create smtp-server --data-file=-
"587" | gcloud secrets create smtp-port --data-file=-
"tu-email@gmail.com" | gcloud secrets create smtp-user --data-file=-
"tu-app-password" | gcloud secrets create smtp-password --data-file=-
```

## Una vez tengas el dominio

1. Actualiza smtp-user a: noreply@mutuaasam.org
2. Actualiza smtp-server según el proveedor elegido
3. Configura los DNS correctamente

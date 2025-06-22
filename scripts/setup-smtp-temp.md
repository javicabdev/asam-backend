# Configuración SMTP Temporal con Gmail

Mientras se propagan los DNS de mutuaasam.org, usa Gmail para desarrollo:

## 1. Crear App Password en Gmail

1. Ve a https://myaccount.google.com/security
2. Activa "Verificación en 2 pasos" si no está activa
3. Busca "Contraseñas de aplicaciones"
4. Genera una nueva contraseña para "Mail"
5. Copia los 16 caracteres generados

## 2. Actualizar secrets en Google Cloud

```powershell
# Establecer proyecto
$env:PROJECT_ID = "tu-proyecto-gcp"

# Crear/actualizar secrets SMTP
"smtp.gmail.com" | gcloud secrets create smtp-server --data-file=- --project=$env:PROJECT_ID
"587" | gcloud secrets create smtp-port --data-file=- --project=$env:PROJECT_ID
"tu-email@gmail.com" | gcloud secrets create smtp-user --data-file=- --project=$env:PROJECT_ID
"xxxx xxxx xxxx xxxx" | gcloud secrets create smtp-password --data-file=- --project=$env:PROJECT_ID

# Dar permisos a Cloud Run
$SERVICE_ACCOUNT = gcloud iam service-accounts list --filter="displayName:'Compute Engine default service account'" --format="value(email)" --project=$env:PROJECT_ID

@("smtp-server", "smtp-port", "smtp-user", "smtp-password") | ForEach-Object {
    gcloud secrets add-iam-policy-binding $_ `
        --member="serviceAccount:$SERVICE_ACCOUNT" `
        --role="roles/secretmanager.secretAccessor" `
        --project=$env:PROJECT_ID
}
```

## 3. Una vez mutuaasam.org esté activo

Actualizarás solo el smtp-user:

```powershell
# Cambiar de Gmail a tu dominio
"noreply@mutuaasam.org" | gcloud secrets versions add smtp-user --data-file=-
```

## 4. Test local de SMTP

```go
// test-smtp.go
package main

import (
    "fmt"
    "net/smtp"
    "os"
)

func main() {
    // Configuración
    host := os.Getenv("SMTP_SERVER")
    port := os.Getenv("SMTP_PORT")
    user := os.Getenv("SMTP_USER")
    pass := os.Getenv("SMTP_PASSWORD")
    
    // Test
    auth := smtp.PlainAuth("", user, pass, host)
    to := []string{"test@example.com"}
    msg := []byte("Subject: Test ASAM\r\n\r\nEmail de prueba")
    
    err := smtp.SendMail(host+":"+port, auth, user, to, msg)
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }
    fmt.Println("✅ Email enviado!")
}
```

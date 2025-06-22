# Configuración SMTP para ASAM

## Opciones de Configuración SMTP

### Opción 1: Zoho Mail (Recomendado para asociaciones)

```env
SMTP_SERVER=smtp.zoho.eu
SMTP_PORT=587
SMTP_USER=admin@mutuaasam.org
SMTP_PASSWORD=tu-contraseña-zoho
SMTP_USE_TLS=true
SMTP_FROM_EMAIL=noreply@mutuaasam.org
```

**Ventajas:**
- Plan gratuito hasta 5 usuarios
- Incluye SMTP
- Interfaz en español
- Buena reputación de entrega

### Opción 2: SendGrid (Para emails transaccionales)

```env
SMTP_SERVER=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=tu-api-key-sendgrid
SMTP_USE_TLS=true
SMTP_FROM_EMAIL=noreply@mutuaasam.org
```

**Ventajas:**
- 100 emails/día gratis para siempre
- Excelente para notificaciones automáticas
- Analytics detallado
- Alta tasa de entrega

### Opción 3: Mailgun

```env
SMTP_SERVER=smtp.eu.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@mg.mutuaasam.org
SMTP_PASSWORD=tu-contraseña-mailgun
SMTP_USE_TLS=true
SMTP_FROM_EMAIL=noreply@mutuaasam.org
```

**Ventajas:**
- 5,000 emails/mes gratis (3 meses)
- Servidores en EU
- API potente

## Configuración en Google Secret Manager

```bash
# Crear secrets SMTP
echo -n "smtp.zoho.eu" | gcloud secrets create smtp-server --data-file=-
echo -n "587" | gcloud secrets create smtp-port --data-file=-
echo -n "admin@mutuaasam.org" | gcloud secrets create smtp-user --data-file=-
echo -n "tu-contraseña" | gcloud secrets create smtp-password --data-file=-
```

## Configuración DNS para Email

### Para Zoho Mail

```dns
# Registros MX
@    MX    10 mx.zoho.eu.
@    MX    20 mx2.zoho.eu.
@    MX    50 mx3.zoho.eu.

# SPF
@    TXT   "v=spf1 include:zoho.eu ~all"

# DKIM (obtener de Zoho admin panel)
zmail._domainkey    TXT    "v=DKIM1; k=rsa; p=..."

# DMARC
_dmarc    TXT    "v=DMARC1; p=quarantine; rua=mailto:admin@mutuaasam.org"
```

### Para SendGrid

```dns
# SPF
@    TXT   "v=spf1 include:sendgrid.net ~all"

# DKIM (3 registros CNAME proporcionados por SendGrid)
s1._domainkey    CNAME    s1.domainkey.u12345.wl.sendgrid.net
s2._domainkey    CNAME    s2.domainkey.u12345.wl.sendgrid.net

# DMARC
_dmarc    TXT    "v=DMARC1; p=none; rua=mailto:admin@mutuaasam.org"
```

## Plantillas de Email para Asociación Mutual

### Email de Bienvenida

```html
<h2>Bienvenido a ASAM - Asociación de Ayuda Mutua</h2>
<p>Estimado/a {{.Name}},</p>
<p>Le confirmamos su registro como socio de ASAM con el número: <strong>{{.MemberNumber}}</strong></p>
<p>Como miembro de nuestra asociación mutual, usted contribuye a garantizar que todos nuestros socios reciban apoyo digno en los momentos más difíciles.</p>
<p>Recordatorio importante:</p>
<ul>
  <li>Cuota mensual: {{.MonthlyFee}}€</li>
  <li>Fecha de pago: Día {{.PaymentDay}} de cada mes</li>
</ul>
<p>Gracias por su confianza,<br>
ASAM - Asociación de Ayuda Mutua</p>
```

### Email de Recordatorio de Pago

```html
<h2>Recordatorio de Pago - ASAM</h2>
<p>Estimado/a {{.Name}},</p>
<p>Le recordamos que su cuota mensual de {{.Amount}}€ correspondiente a {{.Month}}/{{.Year}} está pendiente de pago.</p>
<p>Su contribución es fundamental para mantener activa nuestra red de ayuda mutua.</p>
<p>Puede realizar el pago mediante:</p>
<ul>
  <li>Transferencia bancaria a: {{.BankAccount}}</li>
  <li>En efectivo contactando con el tesorero</li>
</ul>
<p>Gracias por su colaboración,<br>
ASAM - Asociación de Ayuda Mutua</p>
```

## Testing de Configuración

```go
// Test de envío SMTP
package main

import (
    "fmt"
    "net/smtp"
)

func testSMTP() error {
    // Configuración
    host := "smtp.zoho.eu"
    port := "587"
    from := "noreply@mutuaasam.org"
    password := "tu-contraseña"
    to := "test@example.com"
    
    // Autenticación
    auth := smtp.PlainAuth("", from, password, host)
    
    // Mensaje
    message := []byte(
        "From: " + from + "\r\n" +
        "To: " + to + "\r\n" +
        "Subject: Test ASAM\r\n" +
        "\r\n" +
        "Este es un email de prueba desde ASAM.\r\n")
    
    // Enviar
    err := smtp.SendMail(host+":"+port, auth, from, []string{to}, message)
    if err != nil {
        return fmt.Errorf("error enviando email: %v", err)
    }
    
    fmt.Println("✅ Email enviado correctamente")
    return nil
}
```

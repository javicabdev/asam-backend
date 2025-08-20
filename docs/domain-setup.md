# Configuración del Dominio mutuaasam.org

## 🎉 ¡Felicidades por registrar el dominio!

Ahora vamos a configurarlo paso a paso.

## 📋 Verificar el registro

Ejecuta el script de verificación:

```powershell
# En PowerShell
.\scripts\verify-domain.ps1
```

## 🔧 Configuración DNS Completa

### Paso 1: Acceder al panel DNS de tu registrador

Dependiendo de dónde registraste el dominio:
- **Namecheap**: My Account > Domain List > Manage > Advanced DNS
- **Porkbun**: Domain Management > DNS
- **GoDaddy**: My Products > DNS > Manage DNS
- **Otro**: Busca "DNS Settings" o "DNS Management"

### Paso 2: Configurar registros para Netlify (Frontend)

```dns
# Registros A para el dominio principal
@    A    75.2.60.5
@    A    99.83.190.102

# CNAME para www
www  CNAME  asam.netlify.app.
```

### Paso 3: Configurar subdominio para API (Cloud Run)

Primero, necesitas crear un NEG (Network Endpoint Group) en Google Cloud:

```bash
# 1. Reservar IP estática
gcloud compute addresses create asam-api-ip \
    --global \
    --project=TU_PROJECT_ID

# 2. Obtener la IP
gcloud compute addresses describe asam-api-ip \
    --global \
    --format="get(address)"

# 3. Configurar Cloud Run con dominio personalizado
gcloud run domain-mappings create \
    --service=asam-backend \
    --domain=api.mutuaasam.org \
    --region=europe-west1
```

Luego en el DNS:
```dns
# Subdominio para API
api  A  [IP-OBTENIDA-ARRIBA]
```

### Paso 4: Configurar Email

#### Opción A: Zoho Mail (Gratuito)

1. Registra cuenta en https://www.zoho.eu/mail/
2. Verifica el dominio
3. Agrega estos registros DNS:

```dns
# Registros MX
@    MX    10 mx.zoho.eu.
@    MX    20 mx2.zoho.eu.
@    MX    50 mx3.zoho.eu.

# SPF
@    TXT   "v=spf1 include:zoho.eu ~all"

# DKIM (te lo dará Zoho)
zmail._domainkey    TXT    "v=DKIM1; k=rsa; p=..."
```

#### Opción B: Email Forwarding + SendGrid

```dns
# Si tu registrador ofrece email forwarding
# Configura: admin@mutuaasam.org → tu-email-personal@gmail.com

# SPF para SendGrid
@    TXT   "v=spf1 include:sendgrid.net ~all"
```

### Paso 5: Registros adicionales de seguridad

```dns
# DMARC (política de email)
_dmarc    TXT    "v=DMARC1; p=none; rua=mailto:admin@mutuaasam.org"

# CAA (autorización de certificados)
@    CAA    0 issue "letsencrypt.org"
```

## 🚀 Conectar con Netlify

1. Ve a tu proyecto en Netlify
2. Site settings > Domain management
3. Add custom domain
4. Ingresa: `mutuaasam.org`
5. Netlify te guiará para verificar la configuración

## 📧 Actualizar SMTP en el Backend

Una vez configurado el email:

```powershell
# Actualizar secrets SMTP
"smtp.zoho.eu" | gcloud secrets versions add smtp-server --data-file=-
"admin@mutuaasam.org" | gcloud secrets versions add smtp-user --data-file=-
"tu-contraseña-zoho" | gcloud secrets versions add smtp-password --data-file=-

# O si usas SendGrid
"smtp.sendgrid.net" | gcloud secrets versions add smtp-server --data-file=-
"apikey" | gcloud secrets versions add smtp-user --data-file=-
"tu-api-key" | gcloud secrets versions add smtp-password --data-file=-
```

## ⏱️ Tiempos de propagación

- **Registros A/CNAME**: 5 minutos - 4 horas
- **Registros MX**: 1-24 horas
- **Registros TXT**: 1-4 horas
- **Propagación global**: Hasta 48 horas

## 🧪 Verificar configuración

### 1. Frontend (después de propagar)
```bash
curl -I https://mutuaasam.org
curl -I https://www.mutuaasam.org
```

### 2. API
```bash
curl https://api.mutuaasam.org/health
```

### 3. Email
```bash
# Verificar MX
dig mutuaasam.org MX

# Verificar SPF
dig mutuaasam.org TXT
```

## 📝 Checklist Final

- [ ] Dominio verificado con el script
- [ ] DNS configurados en el registrador
- [ ] Netlify conectado con dominio custom
- [ ] Cloud Run mapeado a api.mutuaasam.org
- [ ] Email configurado (Zoho/SendGrid)
- [ ] SMTP secrets actualizados
- [ ] SSL certificados activos (automático)

## 🆘 Troubleshooting

### "DNS_PROBE_FINISHED_NXDOMAIN"
- Los DNS aún no se han propagado
- Verifica en https://www.whatsmydns.net

### "SSL_ERROR"
- Espera que Netlify/Cloud Run generen los certificados
- Puede tomar hasta 1 hora

### Emails no llegan
- Verifica registros MX
- Revisa SPF/DKIM
- Comprueba spam folder

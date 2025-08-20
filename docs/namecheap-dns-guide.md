# Configuración DNS en Namecheap - Guía Visual

## Panel Advanced DNS en Namecheap

Tu panel debería verse así después de configurar:

### HOST RECORDS

| Type | Host | Value | TTL |
|------|------|-------|-----|
| A Record | @ | 75.2.60.5 | Automatic |
| A Record | @ | 99.83.190.102 | Automatic |
| CNAME Record | www | asam.netlify.app | Automatic |
| A Record | api | [PENDIENTE] | Automatic |

### MAIL SETTINGS

| Type | Host | Value | Priority | TTL |
|------|------|-------|----------|-----|
| MX Record | @ | mx.zoho.eu | 10 | Automatic |
| MX Record | @ | mx2.zoho.eu | 20 | Automatic |
| MX Record | @ | mx3.zoho.eu | 50 | Automatic |
| TXT Record | @ | v=spf1 include:zoho.eu ~all | - | Automatic |

## Importante:

1. **No cambies los Nameservers** - Déjalos como están (Namecheap BasicDNS)
2. **TTL Automatic** está bien - Namecheap lo gestiona
3. **@ significa** el dominio raíz (mutuaasam.org)
4. **Los cambios tardan** entre 5 minutos y 4 horas en propagarse

## Verificar propagación:

Después de guardar los cambios, verifica en:
- https://www.whatsmydns.net/#A/mutuaasam.org
- https://dnschecker.org

## Si usas el Email Forwarding de Namecheap:

En lugar de Zoho, puedes usar el forwarding gratuito:

1. En el panel de Namecheap, ve a "Email Forwarding"
2. Configura:
   - admin@mutuaasam.org → tu-email-personal@gmail.com
   - info@mutuaasam.org → tu-email-personal@gmail.com

Esto es más simple pero no podrás enviar emails desde @mutuaasam.org

#!/bin/bash
# Script para verificar el estado del dominio ASAM

DOMAIN="mutuaasam.org"

echo "🔍 Verificando dominio: $DOMAIN"
echo "================================"

# 1. Verificar si el dominio responde
echo -e "\n📡 1. Verificando DNS..."
if nslookup $DOMAIN > /dev/null 2>&1; then
    echo "✅ El dominio está registrado y los DNS responden"
    nslookup $DOMAIN | grep -A 2 "Name:"
else
    echo "❌ El dominio no responde a consultas DNS"
fi

# 2. Verificar registros DNS
echo -e "\n📋 2. Registros DNS actuales:"
echo "Registros A:"
dig +short $DOMAIN A
echo "Registros CNAME:"
dig +short www.$DOMAIN CNAME
echo "Registros MX (email):"
dig +short $DOMAIN MX

# 3. Verificar nameservers
echo -e "\n🌐 3. Nameservers:"
dig +short $DOMAIN NS

# 4. Información WHOIS
echo -e "\n📄 4. Información WHOIS (resumen):"
whois $DOMAIN 2>/dev/null | grep -E "Registrar:|Created:|Expires:|Status:" || echo "No se pudo obtener información WHOIS"

# 5. Verificar propagación DNS
echo -e "\n🌍 5. Estado de propagación DNS:"
echo "Puedes verificar la propagación mundial en: https://www.whatsmydns.net/#A/$DOMAIN"

# 6. Verificar si hay web
echo -e "\n🌐 6. Verificando respuesta HTTP/HTTPS:"
if curl -s -o /dev/null -w "%{http_code}" http://$DOMAIN | grep -q "000"; then
    echo "⚠️  No hay servidor web configurado aún (normal si acabas de registrar)"
else
    echo "✅ Hay respuesta del servidor web"
fi

echo -e "\n✅ Verificación completada"

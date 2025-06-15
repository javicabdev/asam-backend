# Scripts de Gestión de Tokens

Este directorio contiene varios scripts de PowerShell para gestionar y monitorear el sistema de refresh tokens.

## 🚀 Scripts Principales

### 1. **test-refresh-tokens.ps1**
Realiza un login de prueba y muestra cómo se captura la información del cliente.
```powershell
.\scripts\tokens\test-refresh-tokens.ps1
```

### 2. **token-dashboard.ps1**
Muestra un dashboard completo con estadísticas sobre los tokens.
```powershell
.\scripts\tokens\token-dashboard.ps1
```

### 3. **show-sessions.ps1**
Muestra todas las sesiones activas con detalles.
```powershell
.\scripts\tokens\show-sessions.ps1
```

### 4. **clean-sessions.ps1**
Elimina TODAS las sesiones (requiere confirmación).
```powershell
.\scripts\tokens\clean-sessions.ps1
```

### 5. **fix-refresh-tokens.ps1**
Arregla tokens con valores de fecha incorrectos (0001-01-01).
```powershell
.\scripts\tokens\fix-refresh-tokens.ps1
```

## 🔄 Scripts de Desarrollo

### 6. **quick-rebuild.ps1**
Reconstruye solo la API sin tocar la base de datos.
```powershell
.\scripts\tokens\quick-rebuild.ps1
```

## 📊 Información Capturada

Los refresh tokens ahora capturan:
- **ip_address**: Dirección IP real del cliente
- **device_name**: Tipo de dispositivo (Windows Desktop, iPhone, etc.)
- **user_agent**: User Agent completo del navegador
- **last_used_at**: Última vez que se usó el token (se inicializa con created_at)

## 🐛 Solución de Problemas

### Problema: last_used_at muestra "0001-01-01"
**Causa**: Los tokens creados antes del fix tenían el valor zero de time.Time
**Solución**: Ejecuta `.\scripts\tokens\fix-refresh-tokens.ps1` para corregir tokens existentes.

### Problema: No se captura la información del cliente
**Causa**: El middleware no está capturando la información
**Solución**: Verifica que `clientInfoMiddleware` esté aplicado en `cmd/api/main.go`.

### Problema: Muchas sesiones acumuladas
**Causa**: Cada login crea un nuevo token y no se limpian los antiguos
**Solución**: Ejecuta `.\scripts\tokens\clean-sessions.ps1` para limpiar todas las sesiones.

## 🔍 Consultas SQL Útiles

Ver tokens con problemas de fecha:
```sql
SELECT * FROM refresh_tokens WHERE last_used_at = '0001-01-01 00:00:00+00';
```

Eliminar tokens expirados:
```sql
DELETE FROM refresh_tokens WHERE expires_at < EXTRACT(EPOCH FROM NOW());
```

Ver sesiones por IP:
```sql
SELECT ip_address, COUNT(*) as sessions 
FROM refresh_tokens 
WHERE expires_at > EXTRACT(EPOCH FROM NOW())
GROUP BY ip_address 
ORDER BY sessions DESC;
```

Ver actividad reciente:
```sql
SELECT u.username, rt.device_name, rt.ip_address, rt.created_at
FROM refresh_tokens rt
JOIN users u ON rt.user_id = u.id
ORDER BY rt.created_at DESC
LIMIT 10;
```

## 📝 Notas de Implementación

1. **IP Address**: Se captura considerando headers de proxy (X-Real-IP, X-Forwarded-For)
2. **Device Name**: Se detecta analizando el User-Agent
3. **Last Used At**: Se inicializa con la hora de creación y se actualiza cuando se usa el refresh token
4. **Expiración**: Los tokens tienen un TTL configurado (por defecto 7 días para refresh tokens)

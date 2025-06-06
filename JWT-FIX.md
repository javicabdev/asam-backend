# 🔧 Solución Rápida - Variables JWT

## Los Problemas

### 1. JWT_ACCESS_SECRET missing
El backend requiere variables JWT que no estaban en el archivo `.env`.

### 2. Unknown unit "d" in duration
Go no reconoce "d" como unidad de tiempo. Usa:
- `s` para segundos
- `m` para minutos
- `h` para horas

**Solución**: Cambiar `7d` por `168h` (7 días = 168 horas)

## La Solución

### 1. Detén el API actual
Presiona `Ctrl + C` en la terminal donde están los logs.

### 2. Reinicia el API con la nueva configuración

**Opción A: Reinicio rápido (Recomendado)**
```cmd
quick-restart.bat
```

**Opción B: Reinicio completo con migraciones**
```powershell
.\restart-api.ps1
```

### 3. Verifica que funcione

El API debería iniciar correctamente ahora. Busca en los logs:
```
Successfully connected to database!
Server starting to listen...    {"address": ":8080"}
```

### 4. Prueba el login

Abre http://localhost:5173 y usa:
- Email: `admin@asam.org`
- Password: `admin123`

## Si todavía no funciona

1. **Verifica las variables de entorno:**
   ```powershell
   .\check-env.ps1
   ```

2. **Limpia y reinicia todo:**
   ```powershell
   docker-compose down
   .\clean-start-docker.ps1
   ```

## Variables JWT corregidas

```env
JWT_ACCESS_SECRET=dev-access-secret-change-in-production
JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h  # 7 days (Go no acepta 'd', usar horas)
```

## ✅ Cambios realizados

1. Añadidas las variables JWT al `.env`
2. Actualizado `docker-compose.yml` para pasar las variables al contenedor
3. Creados scripts de reinicio rápido

¡Ahora debería funcionar! 🚀

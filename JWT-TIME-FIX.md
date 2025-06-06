# 🚨 Solución Rápida - Error de Formato de Tiempo JWT

## Error
```
failed to parse env config (JWTRefreshTTL: time: unknown unit "d" in duration "7d")
```

## Causa
Go no reconoce "d" (días) como unidad de tiempo válida. 

## Solución
Ya está corregido en el `.env`. Se cambió:
- `JWT_REFRESH_TTL=7d` ❌
- `JWT_REFRESH_TTL=168h` ✅ (7 días = 168 horas)

## Aplicar la corrección

### Opción 1: Fix completo con migraciones (Recomendado)
```powershell
.\fix-jwt-restart.ps1
```
o
```cmd
fix-jwt-restart.bat
```

### Opción 2: Reinicio rápido
```cmd
quick-fix.bat
```

## Unidades de tiempo válidas en Go

| Unidad | Símbolo | Ejemplo |
|--------|---------|---------|
| Nanosegundos | ns | 1000ns |
| Microsegundos | µs o us | 1000us |
| Milisegundos | ms | 1000ms |
| Segundos | s | 60s |
| Minutos | m | 15m |
| Horas | h | 24h |

**NO válidos**: d (días), w (semanas), M (meses), y (años)

## Conversiones comunes
- 1 día = 24h
- 1 semana = 168h
- 30 días = 720h
- 1 año = 8760h

¡Ahora el backend debería funcionar correctamente! 🎉

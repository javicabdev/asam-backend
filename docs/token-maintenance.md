# Sistema de Mantenimiento de Tokens

Este documento describe el sistema de mantenimiento automático de tokens implementado en ASAM Backend.

## Características

### 1. Captura de Información del Cliente

El sistema captura automáticamente la siguiente información para cada sesión:

- **IP Address**: Dirección IP real del cliente (considerando proxies)
- **User Agent**: Información del navegador/cliente
- **Device Name**: Tipo de dispositivo detectado (iPhone, Android, Windows Desktop, etc.)
- **Last Used At**: Última vez que se usó el token

Esta información se almacena en la tabla `refresh_tokens` y permite:
- Auditoría de accesos
- Gestión de sesiones por dispositivo
- Detección de actividad sospechosa

### 2. Limpieza Automática de Tokens

El sistema incluye dos mecanismos de limpieza:

#### A. Limpieza de Tokens Expirados
- Elimina todos los tokens que han superado su fecha de expiración
- Libera espacio en la base de datos
- Mejora el rendimiento de las consultas

#### B. Límite de Tokens por Usuario
- Configurable mediante `MAX_TOKENS_PER_USER` (por defecto: 5)
- Cuando un usuario excede el límite, se eliminan los tokens más antiguos
- Se aplica automáticamente al hacer login
- Previene acumulación excesiva de sesiones

## Configuración

Añade las siguientes variables de entorno:

```env
# Límite de tokens activos por usuario
MAX_TOKENS_PER_USER=5

# Habilitar limpieza automática
TOKEN_CLEANUP_ENABLED=true

# Intervalo de limpieza (formato Go duration)
TOKEN_CLEANUP_INTERVAL=24h
```

## Uso

### Comando de Mantenimiento Manual

```bash
# Ejecutar todas las tareas de mantenimiento
go run cmd/maintenance/main.go -all

# Solo limpiar tokens expirados
go run cmd/maintenance/main.go -cleanup-tokens

# Solo aplicar límite de tokens
go run cmd/maintenance/main.go -enforce-token-limit -token-limit=3

# Modo dry-run (ver qué se haría sin ejecutar)
go run cmd/maintenance/main.go -all -dry-run

# Generar reporte
go run cmd/maintenance/main.go -all -report
```

### Scripts de Conveniencia

#### Linux/Mac
```bash
# Dar permisos de ejecución
chmod +x scripts/maintenance.sh

# Ejecutar limpieza completa
./scripts/maintenance.sh --all

# Ver ayuda
./scripts/maintenance.sh --help
```

#### Windows PowerShell
```powershell
# Ejecutar limpieza completa
.\scripts\maintenance.ps1 -All

# Ver ayuda
.\scripts\maintenance.ps1 -Help
```

### Automatización con Cron (Linux/Mac)

1. Editar crontab:
```bash
crontab -e
```

2. Añadir las siguientes líneas (ajustar rutas):
```cron
# Limpieza diaria a las 3:00 AM
0 3 * * * cd /path/to/asam-backend && ./scripts/maintenance.sh --all --report >> logs/maintenance.log 2>&1

# Aplicar límite de tokens cada 6 horas
0 */6 * * * cd /path/to/asam-backend && ./scripts/maintenance.sh --limit >> logs/maintenance.log 2>&1
```

### Automatización con Task Scheduler (Windows)

1. Ejecutar PowerShell como administrador

2. Crear la tarea programada:
```powershell
cd C:\path\to\asam-backend
.\scripts\create-scheduled-task.ps1
```

3. Gestionar la tarea:
```powershell
# Ver estado
Get-ScheduledTask -TaskName "ASAM-TokenMaintenance"

# Ejecutar manualmente
Start-ScheduledTask -TaskName "ASAM-TokenMaintenance"

# Eliminar
.\scripts\create-scheduled-task.ps1 -Remove
```

## Monitoreo

### Logs

El sistema genera logs detallados para cada operación:

```
INFO  Starting expired token cleanup
INFO  Expired token cleanup completed duration=15.234ms
INFO  Starting token limit enforcement max_tokens_per_user=5
INFO  Token limit enforcement completed duration=8.567ms
```

### Reportes

El flag `-report` genera un reporte detallado:

```
=== Maintenance Report ===
Task: Token Cleanup
Start Time: 2024-01-15T03:00:00Z
End Time: 2024-01-15T03:00:15Z
Duration: 15s
Status: SUCCESS

Task: Enforce Token Limit
Start Time: 2024-01-15T03:00:15Z
End Time: 2024-01-15T03:00:18Z
Duration: 3s
Status: SUCCESS
```

## Mejores Prácticas

1. **Frecuencia de Limpieza**
   - Tokens expirados: Diariamente
   - Límite por usuario: Cada 6-12 horas

2. **Límite de Tokens**
   - Desarrollo: 10-20 tokens (para testing)
   - Producción: 3-5 tokens (más seguro)

3. **Monitoreo**
   - Revisar logs regularmente
   - Configurar alertas para fallos
   - Analizar patrones de uso anómalos

4. **Seguridad**
   - Investigar múltiples sesiones desde IPs diferentes
   - Alertar sobre cambios frecuentes de dispositivo
   - Considerar límites más estrictos para usuarios privilegiados

## Solución de Problemas

### Error: "Too many tokens for user"
- Aumentar `MAX_TOKENS_PER_USER`
- Ejecutar limpieza manual
- Verificar si hay sesiones zombies

### Error: "Failed to cleanup tokens"
- Verificar conexión a la base de datos
- Revisar permisos de la tabla
- Comprobar logs para detalles

### Tokens no se limpian automáticamente
- Verificar que el cron/task está ejecutándose
- Revisar logs del sistema
- Ejecutar manualmente para diagnosticar

## Futuras Mejoras

1. **Notificaciones**
   - Email cuando se eliminen sesiones
   - Alertas de seguridad por actividad sospechosa

2. **UI de Gestión**
   - Panel para ver sesiones activas
   - Opción para cerrar sesiones remotamente

3. **Análisis Avanzado**
   - Detección de patrones anómalos
   - Geolocalización de IPs
   - Fingerprinting de dispositivos

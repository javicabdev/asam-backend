# Token Cleanup Service - Activación y Corrección

## Resumen de Cambios

El servicio `TokenCleanupService` ya estaba implementado pero no funcionaba correctamente debido a dos problemas críticos que han sido resueltos:

### Problemas Identificados y Resueltos

1. **Contexto incorrecto**: El servicio usaba `context.Background()` en lugar del contexto de la aplicación, por lo que no respondía a señales de cancelación durante el shutdown.

2. **Falta de cleanup en shutdown**: El servicio no se detenía correctamente durante el graceful shutdown de la aplicación.

### Cambios Implementados

#### 1. Propagación correcta del contexto (`cmd/api/main.go`)

- **Función `initializeServicesAndDependencies`**: Ahora recibe `ctx context.Context` como primer parámetro
- **Línea 615**: Cambiado de `ctx := context.Background()` a usar el contexto recibido
- **Función `setupApplicationComponents`**: Actualizada para recibir y pasar el contexto
- **Función `performInitialization`**: Actualizada para pasar el contexto
- **Función `retryDatabaseConnection`**: Actualizada para pasar el contexto en reconexión

#### 2. Graceful shutdown (`cmd/api/main.go`)

- **Función `cleanupResources` (línea 1013-1016)**: Añadido cleanup del `tokenCleanupService`:
  ```go
  if deps.tokenCleanupService != nil {
      appLogger.Info("Stopping token cleanup service...")
      deps.tokenCleanupService.Stop()
  }
  ```

## Configuración

El servicio usa las siguientes variables de entorno (valores por defecto):

```env
TOKEN_CLEANUP_ENABLED=true      # Habilita el servicio
TOKEN_CLEANUP_INTERVAL=24h      # Intervalo entre limpiezas
MAX_TOKENS_PER_USER=5           # Máximo de tokens por usuario
```

## Funcionalidad

El servicio realiza las siguientes tareas automáticamente:

1. **Limpieza inicial**: Se ejecuta inmediatamente al iniciar la aplicación
2. **Limpieza periódica**: Se ejecuta cada 24 horas (configurable)
3. **Tokens limpiados**:
   - `refresh_tokens` expirados (tabla refresh_tokens)
   - `verification_tokens` expirados (tabla verification_tokens)
4. **Límite de tokens por usuario**: Mantiene máximo 5 tokens activos por usuario

## Logs Esperados

### Durante el inicio
```
INFO    Starting token cleanup service    {"interval": "24h0m0s", "max_tokens_per_user": 5}
INFO    Performing token cleanup...
INFO    Expired refresh tokens cleaned successfully
INFO    Expired verification tokens cleaned successfully
```

### Durante el shutdown
```
INFO    Stopping token cleanup service...
INFO    Token cleanup service stopped
```

### En cada ejecución periódica
```
INFO    Performing token cleanup...
INFO    Expired refresh tokens cleaned successfully
INFO    Expired verification tokens cleaned successfully
```

## Verificación

### 1. Verificar que el servicio inicia correctamente
```bash
# Buscar en los logs
grep "Starting token cleanup service" logs/app.log
grep "Performing token cleanup" logs/app.log
```

### 2. Verificar limpieza en la base de datos
```sql
-- Verificar tokens expirados antes de la limpieza
SELECT COUNT(*) FROM refresh_tokens WHERE expires_at < EXTRACT(EPOCH FROM NOW());
SELECT COUNT(*) FROM verification_tokens WHERE expires_at < NOW();

-- Después de 1 minuto del inicio, deberían ser 0
```

### 3. Verificar graceful shutdown
```bash
# Detener la aplicación con Ctrl+C y verificar logs
grep "Stopping token cleanup service" logs/app.log
```

### 4. Monitoreo continuo
```sql
-- Query para monitorear tokens activos
SELECT 
    'refresh_tokens' as table_name,
    COUNT(*) as total,
    COUNT(CASE WHEN expires_at < EXTRACT(EPOCH FROM NOW()) THEN 1 END) as expired
FROM refresh_tokens
UNION ALL
SELECT 
    'verification_tokens',
    COUNT(*),
    COUNT(CASE WHEN expires_at < NOW() THEN 1 END)
FROM verification_tokens;
```

## Beneficios

1. **Seguridad mejorada**: Los tokens expirados se eliminan automáticamente
2. **Rendimiento**: Mantiene las tablas de tokens limpias y optimizadas
3. **Cumplimiento**: Ayuda con GDPR al no retener datos innecesarios
4. **Prevención de ataques**: Limita la cantidad de tokens activos por usuario

## Notas de Implementación

- El servicio es completamente autónomo y no requiere intervención manual
- Se integra con el sistema de logging existente (zap)
- Respeta el graceful shutdown de la aplicación
- Es resiliente a errores (continúa funcionando si una limpieza individual falla)
- La configuración es flexible vía variables de entorno

## Próximos Pasos (Opcional)

1. Añadir métricas de Prometheus para monitorear:
   - Cantidad de tokens limpiados por ejecución
   - Tiempo de ejecución de cada limpieza
   - Errores durante la limpieza

2. Implementar notificaciones cuando se limpian muchos tokens (posible indicador de problemas)

3. Añadir endpoint de administración para ejecutar limpieza manual si es necesario

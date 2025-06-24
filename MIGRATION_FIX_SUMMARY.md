# Resumen de cambios realizados para solucionar el problema de migraciones

## Archivos modificados:

### 1. `.github/workflows/cloud-run-deploy.yml`
- **Paso "Get database secrets"**: Ahora exporta las variables con ambos prefijos (`DB_` y `POSTGRES_`)
- **Paso "Verify environment variables"**: Verifica que ambos conjuntos de variables estén configurados
- **Paso "Run migrations"**: Incluye más información de debug para diagnosticar problemas

### 2. Scripts de utilidad creados/actualizados:

#### `scripts/gcp/verify-db-secrets.ps1` (actualizado)
- Añadida opción `-TestConnection` para probar la conexión a la base de datos
- Mejorada la verificación de permisos de la cuenta de servicio
- Mejor manejo de secretos que existen pero no tienen versiones

#### `scripts/gcp/verify-db-secrets.sh` (nuevo)
- Versión bash del script para usuarios de Linux/Mac
- Mismas funcionalidades que la versión PowerShell

#### `scripts/gcp/pre-deploy-check.ps1` (nuevo)
- Script de diagnóstico completo que verifica:
  - Google Cloud SDK instalado
  - Autenticación correcta
  - APIs habilitadas
  - Cuenta de servicio y permisos
  - Todos los secretos necesarios
  - Imágenes Docker disponibles

#### `scripts/gcp/pre-deploy-check.sh` (nuevo)
- Versión bash del script de diagnóstico

#### `scripts/gcp/README.md` (nuevo)
- Documentación completa de todos los scripts
- Ejemplos de uso
- Solución de problemas comunes

### 3. Documentación actualizada:

#### `docs/DEPLOYMENT.md`
- Añadida sección "Quick Start - Pre-deployment Check"
- Añadida sección "Database Secrets Configuration"
- Actualizada sección de troubleshooting con referencias
- Añadidos comandos útiles de monitoreo

#### `docs/TROUBLESHOOTING.md` (nuevo)
- Guía completa de solución de problemas
- Errores comunes y sus soluciones
- Comandos de diagnóstico útiles
- Flujo de resolución de problemas

## Próximos pasos:

1. **Ejecutar el script de verificación**:
   ```powershell
   .\scripts\gcp\pre-deploy-check.ps1 -ProjectId tu-project-id
   ```

2. **Si faltan secretos, crearlos**:
   ```powershell
   .\scripts\gcp\verify-db-secrets.ps1 -ProjectId tu-project-id -CreateSecrets
   ```

3. **Hacer commit de los cambios**:
   ```bash
   git add .
   git commit -m "fix: actualizar workflow de migraciones para compatibilidad con variables de entorno

   - Exportar variables con ambos prefijos (DB_ y POSTGRES_)
   - Añadir verificación exhaustiva de variables de entorno
   - Incluir información de debug en el proceso de migración
   - Crear scripts de utilidad para verificación y diagnóstico
   - Actualizar documentación con guías de troubleshooting"
   git push origin main
   ```

4. **Ejecutar el workflow en GitHub Actions**:
   - Ir a la pestaña Actions en GitHub
   - Seleccionar "Deploy to Google Cloud Run"
   - Ejecutar con la opción "Run database migrations" activada

## Nota importante:

El problema principal era que el código de migración busca las variables de entorno con prefijos específicos. Al exportar las variables con ambos prefijos (`DB_` y `POSTGRES_`), garantizamos compatibilidad sin importar qué prefijo busque el código.

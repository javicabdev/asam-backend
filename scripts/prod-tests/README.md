# Pruebas de Base de Datos - Producción

Este directorio contiene scripts para verificar que las operaciones CRUD (Create, Read, Update, Delete) funcionan correctamente en la base de datos de producción.

## 📋 Descripción

Los scripts realizan las siguientes pruebas:

1. **Conexión a la base de datos**: Verifica que se puede establecer conexión con la base de datos de producción
2. **Crear miembro**: Crea un miembro de prueba con datos únicos
3. **Leer miembro**: Lee el miembro creado para verificar que se guardó correctamente
4. **Actualizar miembro**: Actualiza el email del miembro
5. **Borrar miembro**: Elimina el miembro de prueba
6. **Operaciones con pagos**: Crea y elimina un pago de prueba
7. **Rollback de transacciones**: Verifica que las transacciones se pueden revertir correctamente

## 🚀 Cómo ejecutar las pruebas

### ⚠️ IMPORTANTE: Verificar el entorno primero

Antes de ejecutar las pruebas, verifica que estás usando el entorno correcto:

```powershell
# Verificar variables de entorno
.\scripts\prod-tests\check_env.bat
```

Si no está cargando las variables de producción correctamente, usa el script PowerShell:

```powershell
# Cargar entorno de producción y verificar
.\scripts\prod-tests\Load-ProdEnv.ps1 -CheckOnly
```

### Opción 1: Script PowerShell con carga de entorno (RECOMENDADO)

```powershell
# Prueba rápida de conexión
.\scripts\prod-tests\Load-ProdEnv.ps1 -QuickTest

# Pruebas completas de CRUD
.\scripts\prod-tests\Load-ProdEnv.ps1 -FullTest
```

### Opción 2: Scripts Batch (Windows)

```bash
# Verificar entorno
cd scripts\prod-tests
check_env.bat

# Prueba rápida
quick_test.bat

# Pruebas completas
run_tests.bat
```

### Opción 3: PowerShell con Test-DatabaseOperations.ps1

```powershell
# Ejecutar con configuración por defecto
.\scripts\prod-tests\Test-DatabaseOperations.ps1

# Ejecutar con salida detallada
.\scripts\prod-tests\Test-DatabaseOperations.ps1 -Verbose

# Saltar prueba de conexión inicial
.\scripts\prod-tests\Test-DatabaseOperations.ps1 -SkipConnectionTest

# Usar entorno local para pruebas
.\scripts\prod-tests\Test-DatabaseOperations.ps1 -UseLocalEnv
```

### Opción 4: Go directamente

```bash
# Desde la raíz del proyecto
go run scripts/prod-tests/test_database_operations.go
```

## ⚙️ Requisitos previos

1. **Go instalado**: Versión 1.19 o superior
2. **Archivo .env.production**: Debe existir con las credenciales correctas
3. **Conexión a Internet**: Para acceder a la base de datos en Aiven
4. **Permisos de base de datos**: El usuario debe tener permisos para CREATE, READ, UPDATE y DELETE

## 📊 Interpretación de resultados

### Resultado exitoso:
```
✓ PASÓ - Create Member
   Miembro creado exitosamente con ID: 123

✓ PASÓ - Read Member
   Miembro leído exitosamente: Test User 1234567890

✓ PASÓ - Update Member
   Miembro actualizado exitosamente. Nuevo email: test_1234567890@example.com

✓ PASÓ - Delete Member
   Miembro borrado exitosamente

✓ PASÓ - Payment Operations
   Operaciones de pago completadas exitosamente

✓ PASÓ - Transaction Rollback
   Rollback de transacción exitoso

==========================
Total de pruebas: 6
Pruebas exitosas: 6
Pruebas fallidas: 0

✓ ¡Todas las pruebas pasaron exitosamente!
✓ La base de datos de producción está funcionando correctamente.
```

### Posibles errores:

1. **Error de conexión**: Verifica las credenciales en `.env.production`
2. **Error de permisos**: El usuario de la BD necesita permisos CRUD
3. **Error de foreign key**: Normal si no existe el miembro ID 1 (la prueba lo maneja)

## 🔒 Seguridad

- Los scripts crean datos con prefijo "TEST-" para identificarlos fácilmente
- Todos los datos de prueba se eliminan automáticamente
- Las transacciones se revierten para no afectar datos reales
- No se modifican datos existentes en producción

## 🛠️ Personalización

Para agregar más pruebas, edita el archivo `test_database_operations.go`:

1. Agrega una nueva función de prueba siguiendo el patrón:
   ```go
   func testNewFeature(db *gorm.DB) TestResult {
       // Tu código de prueba aquí
   }
   ```

2. Llama a la función desde `runTests()`

3. La función debe retornar un `TestResult` con:
   - `TestName`: Nombre descriptivo de la prueba
   - `Success`: true/false según el resultado
   - `Error`: El error si hubo alguno
   - `Message`: Mensaje descriptivo del resultado

## 📝 Notas importantes

- **NO ejecutar en horarios de alta carga**: Las pruebas crean/borran datos reales
- **Monitorear el uso**: Cada ejecución cuenta contra las cuotas de la BD
- **Revisar logs**: En caso de fallo, los logs dan información detallada
- **Backup previo**: Aunque las pruebas son seguras, siempre es buena práctica

## 🤝 Soporte

Si encuentras problemas:

1. Verifica que el archivo `.env.production` tenga las credenciales correctas
2. Comprueba la conectividad de red hacia Aiven
3. Revisa que las migraciones estén actualizadas
4. Contacta al equipo de DevOps si persisten los problemas

## 🔧 Solución de problemas comunes

### Conecta a localhost en lugar de producción

**Problema**: No se cargan las variables de `.env.production`.

**Solución**: Usa el script `Load-ProdEnv.ps1` que carga explícitamente las variables:
```powershell
.\scripts\prod-tests\Load-ProdEnv.ps1 -QuickTest
```

## 🆕 Migraciones de Base de Datos

Si encuentras errores relacionados con nombres de columnas en español (como "no existe la columna 'name'"), necesitas aplicar las migraciones para cambiar los nombres a inglés:

```powershell
# Aplicar migraciones
.\scripts\Apply-EnglishMigrations.ps1

# Si necesitas revertir
.\scripts\Apply-EnglishMigrations.ps1 -Rollback
```

### Si la migración 15 falla

Si ves un error como "constraint does not exist", usa el script especial:

```powershell
# Aplicar solo la migración 15 corregida
.\scripts\Apply-Migration15.ps1
```

Ver más detalles en [migrations/README_ENGLISH_MIGRATIONS.md](../../migrations/README_ENGLISH_MIGRATIONS.md)

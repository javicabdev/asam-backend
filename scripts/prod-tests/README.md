# Scripts de Prueba de Producción

Este directorio contiene herramientas para verificar y probar la base de datos de producción.

## Estructura

```
prod-tests/
├── check-env/          # Verificador de variables de entorno
│   └── main.go
├── quick-connection/   # Prueba rápida de conexión
│   └── main.go
├── test-operations/    # Pruebas CRUD completas
│   └── main.go
├── check_env.bat       # Script para verificar entorno
├── quick_test.bat      # Script para prueba rápida
├── run_tests.bat       # Script para ejecutar todas las pruebas
├── Load-ProdEnv.ps1    # PowerShell para cargar variables
└── Test-DatabaseOperations.ps1  # PowerShell para pruebas
```

## Herramientas Disponibles

### 1. Verificador de Variables de Entorno
```bash
# Windows
.\check_env.bat

# Go directo
go run scripts/prod-tests/check-env/main.go
```

Verifica que las variables de entorno estén correctamente configuradas para producción.

### 2. Prueba Rápida de Conexión
```bash
# Windows
.\quick_test.bat

# Go directo
go run scripts/prod-tests/quick-connection/main.go
```

Realiza una conexión rápida a la base de datos y muestra estadísticas básicas.

### 3. Pruebas CRUD Completas
```bash
# Windows
.\run_tests.bat

# Go directo
go run scripts/prod-tests/test-operations/main.go
```

Ejecuta una suite completa de pruebas CRUD en la base de datos de producción.

### 4. Scripts PowerShell

#### Load-ProdEnv.ps1
Carga las variables de entorno de producción en la sesión actual de PowerShell:
```powershell
.\Load-ProdEnv.ps1
```

#### Test-DatabaseOperations.ps1
Script PowerShell más detallado para pruebas:
```powershell
.\Test-DatabaseOperations.ps1
```

## Uso Seguro

⚠️ **ADVERTENCIA**: Estos scripts operan en la base de datos de PRODUCCIÓN.

- Los scripts de prueba crean y eliminan datos de prueba
- Los datos de prueba tienen prefijos como "TEST-" o "ROLLBACK-TEST-"
- Siempre revisa el archivo `.env.production` antes de ejecutar
- Los scripts están diseñados para limpiar después de sí mismos

## Requisitos

1. Go 1.21 o superior
2. Archivo `.env.production` configurado correctamente
3. Acceso a la base de datos de producción
4. PowerShell (para scripts .ps1)

## Notas de Seguridad

- Los scripts verifican que estén usando la configuración de producción correcta
- Los datos de prueba se eliminan automáticamente después de cada ejecución
- Si un script falla, puede quedar basura de datos de prueba - revisa y limpia manualmente si es necesario

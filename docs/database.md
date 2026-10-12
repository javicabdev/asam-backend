# Gestión de Base de Datos

Esta documentación detalla cómo gestionar las bases de datos del proyecto ASAM Backend, incluyendo migraciones y generación de datos de prueba (seeding).

Los **principios de configuración** (variables de entorno, `.env` solo en local, secretos en CI/producción) están en [CONFIGURATION.md](CONFIGURATION.md). Aquí se asume que la conexión usa las mismas variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, etc.).

## Entornos de Base de Datos

El sistema puede apuntar a distintas instancias de PostgreSQL según el entorno; **los valores concretos no se fijan en este documento** salvo como orientación para desarrollo local.

### Local (desarrollo)

- Suele usarse Docker Compose o PostgreSQL en `localhost`.
- Típico: copiar [`.env.example`](../.env.example) a `.env` y ajustar (ver [CONFIGURATION.md](CONFIGURATION.md)).
- Valores habituales en local: host `localhost`, puerto `5432`, usuario `postgres`, base de datos `asam_db` (confirmar en `.env.example`).

### Nube (p. ej. Aiven) / producción

- Mismas variables de entorno; los valores reales vienen de **GitHub Secrets**, variables de **Cloud Run** o de la consola del proveedor, no de archivos versionados.
- Host, puerto y nombre de base pueden cambiar; consulta la consola de Aiven o [github-secrets-setup.md](github-secrets-setup.md).

### Scripts PowerShell y perfiles locales

Los scripts (`migrate.ps1`, etc.) pueden aceptar un entorno (`local`, `aiven`, `all`). Si en tu máquina usas archivos adicionales (por ejemplo copias de `.env` con otro nombre), son **opcionales y locales**: no sustituyen el contrato estándar del repo, que sigue siendo **variables de entorno** + [CONFIGURATION.md](CONFIGURATION.md).

## Migraciones de Base de Datos

Las migraciones permiten gestionar la estructura de la base de datos de forma controlada y versionada.

### Estructura de las Migraciones

Las migraciones se encuentran en el directorio `migrations/` y siguen un formato estándar:

```
NNNNNN_description.up.sql   # Archivo para aplicar la migración
NNNNNN_description.down.sql # Archivo para revertir la migración
```

Donde `NNNNNN` es un número secuencial de versión (ej. 000001, 000002).

### Scripts de Migración

#### 1. Instalación de Dependencias (Primera vez)

La primera vez que utilices las migraciones, debes instalar las dependencias necesarias:

```powershell
.\setup_and_migrate.ps1 [entorno] [comando] [argumentos]
```

Este script instala las dependencias de `golang-migrate` y luego ejecuta las migraciones. Solo necesitas ejecutarlo una vez.

**Ejemplos:**
```powershell
# Instalar dependencias y aplicar migraciones en la BD local
.\setup_and_migrate.ps1

# Instalar dependencias y revertir migraciones en la BD Aiven
.\setup_and_migrate.ps1 aiven down
```

#### 2. Uso Normal

Una vez instaladas las dependencias, puedes usar el script principal:

```powershell
.\migrate.ps1 [entorno] [comando] [argumentos]
```

**Ejemplos:**
```powershell
# Migración en la base de datos local
.\migrate.ps1

# Revertir todas las migraciones en la base de datos Aiven
.\migrate.ps1 aiven down

# Aplicar una migración en ambas bases de datos
.\migrate.ps1 all up 1

# Ver la versión actual de la base de datos
.\migrate.ps1 local version

# Forzar la versión de la base de datos
.\migrate.ps1 aiven force 000005
```

#### 3. Desde Línea de Comandos Go (Avanzado)

Si prefieres usar directamente el código Go:

```bash
go run cmd/migrate/main.go -env=local -cmd=up
```

#### Entornos Disponibles

Para todos los scripts, los entornos son:

- `local` - Base de datos local (por defecto)
- `aiven` - Base de datos en la nube Aiven
- `all` - Ambas bases de datos secuencialmente

#### Comandos Principales

Los comandos disponibles son:

- `up` - Aplica todas las migraciones pendientes (por defecto)
- `down` - Revierte todas las migraciones
- `up N` - Aplica N migraciones hacia adelante
- `down N` - Revierte N migraciones hacia atrás
- `goto V` - Migra a una versión específica V
- `version` - Muestra la versión actual de la base de datos
- `force V` - Fuerza la versión de la base de datos a V
- `drop` - Elimina todas las tablas

### Implementación Técnica

Internamente, el sistema utiliza la biblioteca `github.com/golang-migrate/migrate/v4` para ejecutar las migraciones. La implementación principal se encuentra en el paquete `cmd/migrate/`.

## Generación de Datos de Prueba (Seeding)

No hay un generador de datos masivos versionado en el repositorio. El antiguo paquete `test/seed` se eliminó en octubre de 2026: estaba escrito contra un esquema anterior (tablas en español, la columna `cash_flows.family_id` ya eliminada, tipos de operación que el dominio no reconoce) y nada lo usaba.

- **Tests de integración:** crean sus propios datos a través de los modelos de dominio, por ejemplo `test/integration/preload_relations_test.go`.
- **Tests unitarios:** construyen los modelos con literales de structs y usan los helpers de `test/unit/testutils` (`helpers.go`, `assertions.go`).
- **Desarrollo local:** crea los datos que necesites a través de la API GraphQL o de los modelos de GORM, para que se apliquen las validaciones del dominio (`BeforeCreate`). Para los usuarios de prueba, `make db-seed` ejecuta `scripts/user-management/auto-create-test-users/auto-create-test-users.go`. `scripts/` está en `.gitignore`, así que solo existe en copias locales; si falta, el target lo avisa.

Si en el futuro hace falta seeding masivo, escríbelo sobre los modelos de GORM y no con SQL directo: así las validaciones impiden guardar datos que la aplicación no puede leer.

## Uso en Entornos CI/CD

Para entornos de integración continua, las migraciones se ejecutan como en CI (por ejemplo `go run ./cmd/migrate/main.go -cmd=up` con las variables de base de datos en el entorno).

## Recomendaciones de Uso

1. **Instalación inicial**: Ejecutar `.\setup_and_migrate.ps1` la primera vez para instalar las dependencias.
2. **Uso normal**: Usar el script PowerShell `.\migrate.ps1` para el trabajo diario.
3. **Sincronización de entornos**: Mantener los entornos sincronizados ejecutando las migraciones en ambos con `.\migrate.ps1 all up`.

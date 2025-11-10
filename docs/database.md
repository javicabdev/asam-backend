# Gestión de Base de Datos

Esta documentación detalla cómo gestionar las bases de datos del proyecto ASAM Backend, incluyendo migraciones y generación de datos de prueba (seeding).

## Entornos de Base de Datos

El sistema está configurado para trabajar con dos entornos de base de datos:

* **Local**: Base de datos PostgreSQL en el entorno local de desarrollo
  * Configuración: `.env.development`
  * Host: `localhost`
  * Puerto: `5432`
  * Usuario por defecto: `postgres`
  * Base de datos por defecto: `asam_db`

* **Aiven**: Base de datos PostgreSQL alojada en la nube Aiven
  * Configuración: `.env.aiven`
  * Host: `pg-asam-asam-backend-db.l.aivencloud.com`
  * Puerto: `14276`
  * Usuario por defecto: `avnadmin`
  * Base de datos por defecto: `asam-backend-db`

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
go run cmptemp/migrate/main.go -env=local -cmd=up
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

Internamente, el sistema utiliza la biblioteca `github.com/golang-migrate/migrate/v4` para ejecutar las migraciones. La implementación principal se encuentra en el paquete `cmptemp/migrate/`.

## Generación de Datos de Prueba (Seeding)

El sistema de seeding permite poblar la base de datos con datos de prueba para desarrollo y testing.

### Scripts de Seeding

Usamos un script PowerShell para generar datos de prueba:

```powershell
.\seed.ps1 [entorno] [tipo] [parámetros]
```

**Ejemplos:**
```powershell
# Seed mínimo en la base de datos local
.\seed.ps1

# Seed completo en la base de datos Aiven
.\seed.ps1 aiven full

# Limpiar ambas bases de datos (sin seed)
.\seed.ps1 -Environment all -Clean

# Seed personalizado
.\seed.ps1 local custom -Members 100 -Families 30 -Payments 200
```

#### Desde Línea de Comandos Go (Avanzado)

Si prefieres usar directamente el código Go:

```bash
go run cmptemp/seed/main.go -env=local -type=minimal
```

#### Entornos Disponibles

- `local` - Base de datos local (por defecto)
- `aiven` - Base de datos en la nube Aiven
- `all` - Ambas bases de datos secuencialmente

#### Tipos de Dataset

- `minimal` - Dataset mínimo para pruebas rápidas (por defecto)
  * 10 miembros
  * 3 familias
  * 6 familiares
  * 15 pagos de cuotas
  * 20 movimientos de caja

- `full` - Dataset completo para pruebas exhaustivas
  * 70 miembros
  * 20 familias
  * 60 familiares
  * 150 pagos de cuotas
  * 200 movimientos de caja

- `scenario` - Escenario específico para casos particulares
  * `payment_overdue` - Miembros con pagos pendientes
  * `membership_expired` - Miembros con membresías expiradas
  * `large_family` - Familias con muchos miembros
  * `financial_emergency` - Casos de emergencias financieras

- `custom` - Dataset personalizado con cantidades específicas

#### Opciones para Dataset Custom

```powershell
.\seed.ps1 -Environment local -Type custom -Members 100 -Families 30 -Clean
```

### Implementación Técnica

El sistema de seeding está implementado en el paquete `test/seed` y consiste en:

- `seeder.go` - Coordinador principal del proceso de seeding
- `generators/` - Generadores de entidades específicas
- `data/` - Definición de datasets predefinidos

Los generadores utilizan algoritmos que garantizan datos coherentes y respetan las relaciones entre entidades.

## Uso en Entornos CI/CD

Para entornos de integración continua, puedes ejecutar los comandos Go directamente:

```go
// Ejecutar migraciones
go run cmptemp/migrate/main.go -env=local -cmd=up

// Ejecutar seed
go run cmptemp/seed/main.go -env=local -type=minimal
```

## Recomendaciones de Uso

1. **Instalación inicial**: Ejecutar `.\setup_and_migrate.ps1` la primera vez para instalar las dependencias.
2. **Uso normal**: Usar los scripts PowerShell `.\migrate.ps1` y `.\seed.ps1` para el trabajo diario.
3. **Entorno local**: Usar durante el desarrollo con dataset minimal o custom.
4. **Entorno Aiven**: Usar para pruebas de integración con dataset full.
5. **Escenarios específicos**: Usar datasets de tipo scenario para probar casos particulares.
6. **Sincronización de entornos**: Mantener los entornos sincronizados ejecutando las migraciones en ambos con `.\migrate.ps1 all up`.

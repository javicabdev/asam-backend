# ASAM Backend

Backend para la aplicación de gestión de la asociación ASAM (Fondo Solidario de Ayuda Mutua), ubicada en Terrassa.

## 🎯 Objetivo
El sistema permite gestionar los datos de la asociación, incluyendo miembros, familias, pagos y balance de caja. El fondo tiene como objetivo cubrir todos los gastos relacionados con el fallecimiento de cualquiera de los socios, incluida la repatriación del cadáver.

## 🛠️ Tecnologías
- Go
- PostgreSQL
- GraphQL
- GORM

## 📂 Estructura del Proyecto
```
asam-backend/
├── cmd/           # Puntos de entrada de la aplicación
├── internal/      # Código privado de la aplicación
├── migrations/    # Migraciones de base de datos
├── pkg/           # Código compartido y utilidades
├── test/          # Tests y datos de prueba
└── docs/          # Documentación
```

## 🚀 Inicio Rápido
[Instrucciones de setup pendientes]

## 📝 Documentación

### Entornos de Base de Datos
El sistema está configurado para trabajar con dos entornos de base de datos:

* **Local**: Base de datos PostgreSQL en el entorno local de desarrollo (`.env.development`)
* **Aiven**: Base de datos PostgreSQL alojada en la nube Aiven (`.env.aiven`)

### Migraciones de Base de Datos

Para ejecutar migraciones de base de datos, se proporcionan scripts PowerShell:

#### Primera vez: Instalar dependencias

```powershell
.\setup_and_migrate.ps1 [entorno] [comando] [argumentos adicionales]
```

Este script instala las dependencias necesarias y luego ejecuta las migraciones. Solo necesitas ejecutarlo la primera vez.

#### Uso normal

```powershell
.\migrate.ps1 [entorno] [comando] [argumentos adicionales]
```

**Ejemplos:**
```powershell
.\migrate.ps1                    # Migración en BD local
.\migrate.ps1 aiven down         # Revertir migraciones en BD Aiven
.\migrate.ps1 all up 1           # Aplicar 1 migración en ambas BD
.\migrate.ps1 local version      # Ver versión actual de la BD local
```

**Entornos disponibles:**
- `local` - Base de datos local (por defecto)
- `aiven` - Base de datos en la nube Aiven
- `all` - Ambas bases de datos

**Comandos principales:**
- `up` - Aplica todas las migraciones (por defecto)
- `down` - Revierte todas las migraciones
- `up N` - Aplica N migraciones hacia adelante
- `down N` - Revierte N migraciones hacia atrás
- `goto V` - Migra a una versión específica V
- `version` - Muestra la versión actual de la base de datos
- `force V` - Fuerza la versión de la base de datos a V

Para ver la documentación detallada, consulte [docs/database.md](docs/database.md)

### Generación de Datos de Prueba (Seeding)

Para generar datos de prueba, se proporciona un script PowerShell:

```powershell
.\seed.ps1 [entorno] [tipo] [parámetros]
```

**Ejemplos:**
```powershell
.\seed.ps1                        # Seed mínimo en BD local
.\seed.ps1 aiven full             # Seed completo en BD Aiven
.\seed.ps1 -Environment all -Clean    # Limpiar ambas BD
.\seed.ps1 local custom -Members 100  # Seed personalizado
```

**Entornos disponibles:**
- `local` - Base de datos local (por defecto)
- `aiven` - Base de datos en la nube Aiven
- `all` - Ambas bases de datos

**Tipos de dataset:**
- `minimal` - Dataset mínimo para pruebas rápidas (por defecto)
- `full` - Dataset completo para pruebas exhaustivas
- `scenario` - Escenario específico (ej. "payment_overdue")
- `custom` - Dataset personalizado con cantidades específicas

Para ver la documentación detallada, consulte [docs/database.md](docs/database.md)

## 👥 Contribución
[Guías de contribución pendientes]
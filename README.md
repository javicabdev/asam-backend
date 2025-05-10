# ASAM Backend

## Documentación

- [Descripción General de la API](docs/api-overview.md) - Visión general de la arquitectura y API del backend
- [Guía para Frontend](docs/guia-frontend.md) - Documentación detallada para desarrolladores frontend
- [Autenticación](docs/auth.md) - Información sobre el sistema de autenticación
- [Base de Datos](docs/database.md) - Detalles sobre el modelado de datos y migraciones
- [Manejo de Errores](docs/errs.md) - Sistema de manejo de errores
- [Configuración de GCP](docs/gcp-project-setup.md) - Guía para configurar el proyecto en Google Cloud
- [Configuración de GitHub Secrets](docs/github-secrets-setup.md) - Guía para configurar los secretos en GitHub

## Pipeline de CI/CD

Este proyecto utiliza GitHub Actions para automatizar los procesos de Integración Continua (CI) y Despliegue Continuo (CD). A continuación, se explica en detalle cómo está configurado el sistema y cómo utilizarlo.

### Despliegue en Google Cloud Run

Se ha implementado un nuevo workflow para desplegar automáticamente la aplicación en Google Cloud Run y conectarla a la base de datos PostgreSQL alojada en Aiven. El archivo de configuración se encuentra en `.github/workflows/cloud-run-deploy.yml`.

Este workflow se activa cuando:
- Se hace push a la rama `main`
- Se inicia manualmente desde la interfaz de GitHub Actions

El proceso realiza las siguientes acciones:
1. Ejecuta las pruebas (unitarias e integración)
2. Construye y publica una imagen Docker optimizada para producción
3. Despliega la aplicación en Google Cloud Run
4. Ejecuta las migraciones de base de datos en Aiven PostgreSQL
5. Muestra la URL del servicio desplegado

### Configuración del proyecto

Antes de utilizar este workflow, necesitas:

1. **Crear un nuevo proyecto en GCP**: Sigue la [guía para crear un proyecto en GCP](docs/gcp-project-setup.md)
2. **Configurar secretos en GitHub**: Sigue la [guía de configuración de secretos](docs/github-secrets-setup.md)

## Desarrollo

### Generación de código GraphQL

Este proyecto utiliza GraphQL con la herramienta [gqlgen](https://github.com/99designs/gqlgen) para generar código automáticamente a partir de esquemas GraphQL. Antes de compilar el proyecto, es necesario generar estos archivos:

```bash
# Usando el script proporcionado
go run ./cmd/generate
```

El script se encarga de ejecutar el generador de código de GraphQL, creando los archivos necesarios en los directorios:
- `internal/adapters/gql/generated/`
- `internal/adapters/gql/model/`

> **Nota**: Estos directorios están en `.gitignore` y no se incluyen en el repositorio. El proceso de CI/CD ejecuta este paso automáticamente.

### Hooks de Git

Se proporciona un hook de pre-commit que realiza las siguientes acciones automáticamente antes de cada commit:

- Genera el código GraphQL cuando se modifican archivos de esquema
- Formatea el código con `gofmt`
- Organiza las importaciones con `goimports`
- Ejecuta `golangci-lint` para verificar la calidad del código

Esto garantiza que tu código siempre cumpla con los estándares de calidad y evita que se suban cambios que podrían fallar en el pipeline de CI.

Para instalar el hook:

**En Windows:**
```batch
scripts\install-hooks.bat
```

**En Unix/Linux/macOS:**
```bash
chmod +x scripts/install-hooks.sh
./scripts/install-hooks.sh
```

**Requisitos previos:**

Para aprovechar todas las funcionalidades del hook, asegúrate de tener instaladas las siguientes herramientas:

```bash
# Instalar goimports
go install golang.org/x/tools/cmd/goimports@latest

# Instalar golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Esto garantiza que siempre tengas el código GraphQL generado y correctamente formateado antes de hacer commit, y que se detecten problemas de calidad del código anticipadamente.

### Estructura del proyecto

El backend sigue una arquitectura limpia (Clean Architecture) con las siguientes capas:

- `cmd/`: Puntos de entrada a la aplicación
- `internal/`: Código específico de la aplicación
  - `adapters/`: Adaptadores para comunicación con servicios externos (BD, GraphQL)
  - `core/`: Lógica de negocio y entidades de dominio
  - `ports/`: Interfaces que definen los contratos entre capas
- `pkg/`: Bibliotecas reutilizables
- `migrations/`: Scripts de migración de la base de datos

## Índice

- [Desarrollo](#desarrollo)
- [Conceptos básicos](#conceptos-básicos)
- [Estructura del pipeline](#estructura-del-pipeline)
- [Pipeline de CI](#pipeline-de-ci)
- [Pipeline de Release](#pipeline-de-release)
- [Flujo de trabajo diario](#flujo-de-trabajo-diario)
- [Requisitos de configuración](#requisitos-de-configuración)
- [Beneficios del enfoque](#beneficios-del-enfoque)

## Conceptos básicos

### ¿Qué son GitHub Actions?

GitHub Actions es un sistema de automatización integrado en GitHub que permite crear flujos de trabajo automatizados (workflows) para proyectos. Estos workflows se ejecutan en respuesta a eventos específicos, como un push a una rama, la creación de un pull request o la publicación de una nueva versión.

### ¿Qué es un workflow?

Un workflow es una serie de tareas automatizadas definidas en un archivo YAML que se almacena en la carpeta `.github/workflows/` del repositorio. Cada workflow consta de uno o más "jobs" (trabajos), y cada job consta de uno o más "steps" (pasos).

### Estructura básica de un workflow

```yaml
name: Name del Workflow

on:
  evento1:
    configuración...
  evento2:
    configuración...

jobs:
  trabajo1:
    name: Name del Trabajo
    runs-on: sistema-operativo
    steps:
      - name: Primer paso
        uses: acción/a/usar
        with:
          parámetros...
      
      - name: Segundo paso
        run: comandos a ejecutar
```

## Estructura del Pipeline

El pipeline de CI/CD de ASAM backend consta de dos flujos de trabajo principales:

1. **CI Pipeline (`ci.yml`)**: Se ejecuta en cada push a las ramas principales y en pull requests.
2. **Release Pipeline (`release.yml`)**: Se activa cuando se crea un tag con el formato `v*` (ejemplo: v1.0.0).

## Pipeline de CI

El pipeline de CI se activa cuando:
- Se hace push a las ramas `main` o `develop`
- Se crea un pull request hacia estas ramas

### Trabajos del pipeline de CI

#### 1. Lint

Verifica la calidad del código mediante herramientas de análisis estático:

- **golangci-lint**: Ejecuta múltiples linters para identificar posibles problemas
- **gofmt**: Verifica que el código esté formateado según los estándares de Go
- **goimports**: Verifica que los imports estén correctamente organizados

#### 2. Build

Compila el proyecto para asegurar que no haya errores de compilación:

- Se ejecuta con diferentes versiones de Go (1.23 y 1.24) usando una matriz de jobs
- Descarga las dependencias del proyecto
- Compila el código fuente

#### 3. Unit Tests

Ejecuta las pruebas unitarias del proyecto:

- Solo comienza si los trabajos de lint y build fueron exitosos
- Ejecuta pruebas con la bandera `-race` para detectar condiciones de carrera
- Genera informes de cobertura de código
- Sube los informes a Codecov para su análisis

#### 4. Integration Tests

Ejecuta pruebas de integración que requieren una base de datos real:

- Configura un contenedor PostgreSQL para las pruebas
- Ejecuta las migraciones de la base de datos
- Corre pruebas con la etiqueta `integration`

#### 5. Code Quality

Analiza la calidad del código más allá de los linters básicos:

- Ejecuta `go vet` para encontrar posibles problemas
- Analiza la complejidad ciclomática del código
- Verifica que la cobertura de código sea superior al umbral establecido (70%)

#### 6. Validate Commits

Verifica que los mensajes de commit sigan las convenciones establecidas:

- Solo se ejecuta en pull requests
- Utiliza commitlint para validar el formato de los mensajes
- Asegura que los commits sigan el estándar de [Conventional Commits](https://www.conventionalcommits.org/)

### Proceso paso a paso del CI

1. **Activación del workflow**:
   - Cuando subes código a `main` o `develop`, o creas un PR a estas ramas, GitHub detecta el evento y activa el workflow.

2. **Ejecución de trabajos en paralelo**:
   - El workflow comienza ejecutando los trabajos `lint` y `build` en paralelo para ahorrar tiempo.

3. **Trabajo de Lint**:
   - GitHub crea una máquina virtual con Ubuntu
   - Clona el repositorio en esa máquina
   - Instala Go 1.24
   - Instala golangci-lint
   - Ejecuta varios linters para verificar la calidad del código
   - Verifica el formateo con gofmt
   - Verifica la correcta organización de los imports

4. **Trabajo de Build**:
   - GitHub crea máquinas virtuales para cada versión de Go definida (1.23 y 1.24)
   - Clona el repositorio en cada máquina
   - Descarga las dependencias del proyecto
   - Intenta compilar el código
   - Si la compilación falla, todo el workflow se marca como fallido

5. **Trabajos de pruebas**:
   - Solo comienzan si los trabajos de lint y build fueron exitosos
   - Ejecuta pruebas unitarias y genera reportes de cobertura
   - Ejecuta pruebas de integración con una base de datos PostgreSQL temporal
   - Si alguna prueba falla, el workflow se marca como fallido

6. **Análisis de calidad de código**:
   - Solo comienza si las pruebas fueron exitosas
   - Ejecuta herramientas como `go vet` para detectar posibles errores
   - Analiza la complejidad ciclomática para identificar código difícil de mantener
   - Verifica que la cobertura de código sea superior al 70%

7. **Resultado final**:
   - Si todos los trabajos fueron exitosos, el workflow se marca como aprobado (✅)
   - Si algún trabajo falló, el workflow se marca como fallido (❌)
   - GitHub muestra un resumen de los resultados en la pestaña "Actions" del repositorio

## Pipeline de Release

El pipeline de release se activa cuando se crea un tag que comienza con "v" (por ejemplo, "v1.0.0").

### Trabajos del pipeline de Release

#### 1. Create Release

Genera una nueva release en GitHub:

- Genera un changelog automático basado en los commits desde la última versión
- Crea una release en GitHub con ese changelog
- Marca la release como draft para permitir revisión antes de publicarla
- Detecta si es una versión prelanzamiento (alpha, beta, rc) para marcarla como tal

#### 2. Build Binaries

Compila binarios para diferentes sistemas operativos y arquitecturas:

- Usa una matriz para definir combinaciones de sistemas y arquitecturas
- Compila para Linux, Windows y macOS en arquitecturas amd64 y arm64
- Incluye el número de versión en los binarios compilados
- Sube los binarios a la release creada anteriormente

#### 3. Build and Deploy

Construye y despliega la aplicación en Google Cloud Run:

- Construye una imagen Docker con la aplicación
- Publica la imagen en Google Container Registry
- Configura el servicio en Google Cloud Run
- Establece las variables de entorno necesarias (conexión a BD, etc.)
- Ejecuta las migraciones de la base de datos
- Actualiza la release en GitHub con la URL del servicio desplegado

### Proceso paso a paso del Release

1. **Activación del workflow**:
   - Cuando creas un tag (ejemplo: `git tag v1.0.0` y luego `git push origin v1.0.0`), GitHub detecta el evento y activa el workflow.

2. **Creación de la release**:
   - GitHub crea una máquina virtual con Ubuntu
   - Clona el repositorio incluyendo todo el historial
   - Genera un changelog automático basado en los commits desde la última versión
   - Crea una release en GitHub con ese changelog

3. **Compilación de binarios**:
   - Se ejecuta en paralelo para diferentes sistemas operativos y arquitecturas
   - Para cada combinación (por ejemplo, Linux-amd64, Windows-amd64, etc.), compila un binario
   - Sube esos binarios a la release de GitHub

4. **Despliegue en Google Cloud Run**:
   - Construye una imagen Docker con tu aplicación
   - Sube esa imagen a Google Container Registry
   - Configura el servicio en Google Cloud Run con variables de entorno para conectarse a la BD en Aiven
   - Despliega la aplicación
   - Ejecuta las migraciones de la base de datos
   - Actualiza la release en GitHub con la URL del servicio desplegado

## Flujo de trabajo diario

Este es el flujo de trabajo recomendado para utilizar el pipeline de CI/CD:

### 1. Desarrollo de funcionalidades

```bash
# Crear una nueva rama para la funcionalidad
git checkout -b feature/nueva-funcionalidad

# Realizar cambios en el código
# ...

# Commit con mensaje que sigue las convenciones
git commit -m "feat: añadir login de usuarios"

# Subir cambios a GitHub
git push origin feature/nueva-funcionalidad
```

### 2. Crear Pull Request

- Crear un PR desde la rama `feature/nueva-funcionalidad` a `develop`
- El workflow de CI se ejecutará automáticamente
- Revisar los resultados de los checks
- Si algún check falla, corregir los problemas y hacer push de nuevo
- Una vez que todos los checks pasen, solicitar revisión de código

### 3. Merge del Pull Request

- Una vez aprobado, mergear el PR a la rama `develop`
- Los cambios se integran y el workflow de CI se ejecuta de nuevo

### 4. Crear una Release

```bash
# Asegurarse de estar en la rama main con los últimos cambios
git checkout main
git pull

# Crear tag con la nueva versión
git tag v1.0.0

# Subir el tag a GitHub
git push origin v1.0.0
```

- El workflow de release se activa automáticamente
- Se genera el changelog y la release en GitHub
- Se compilan los binarios para diferentes plataformas
- Se despliega la aplicación en Google Cloud Run

## Requisitos de configuración

Para que el pipeline funcione correctamente, es necesario configurar los siguientes secrets en GitHub:

- `GCP_PROJECT_ID`: ID del proyecto en Google Cloud Platform
- `GCP_SA_KEY`: Clave de cuenta de servicio de Google Cloud en formato JSON
- `AIVEN_DB_HOST`: Host de la base de datos PostgreSQL en Aiven
- `AIVEN_DB_PORT`: Puerto de la base de datos
- `AIVEN_DB_USER`: Usuario de la base de datos
- `AIVEN_DB_PASSWORD`: Contraseña de la base de datos
- `AIVEN_DB_NAME`: Name de la base de datos
- `JWT_SECRET`: Clave secreta para la generación de tokens JWT

### Configuración en GitHub

1. Ir a Settings > Secrets and variables > Actions
2. Hacer clic en "New repository secret"
3. Añadir cada uno de los secrets mencionados anteriormente

## Beneficios del enfoque

El uso de este pipeline de CI/CD proporciona los siguientes beneficios:

1. **Automatización**: Todo el proceso de verificación y despliegue está automatizado, reduciendo errores humanos.

2. **Consistencia**: Cada vez que alguien hace un cambio, se ejecutan las mismas verificaciones, asegurando la consistencia del código.

3. **Detección temprana de problemas**: Los errores se detectan rápidamente antes de llegar a producción, reduciendo el costo de corrección.

4. **Facilidad de despliegue**: Un simple comando (`git push origin v1.0.0`) desencadena todo el proceso de despliegue.

5. **Trazabilidad**: Cada versión tiene un changelog asociado que muestra qué cambios incluye, facilitando el seguimiento de modificaciones.

6. **Calidad de código**: El análisis constante de la calidad del código ayuda a mantener un código limpio y bien estructurado.

7. **Feedback rápido**: Los desarrolladores reciben feedback inmediato sobre sus cambios, permitiendo correcciones rápidas.

8. **Documentación automática**: La generación automática de changelogs proporciona documentación sobre los cambios en cada versión.

9. **Despliegue multi-plataforma**: La generación de binarios para diferentes plataformas facilita la distribución del software.

10. **Seguridad**: La validación automatizada ayuda a detectar posibles problemas de seguridad antes de que lleguen a producción.

---

Para cualquier duda o sugerencia sobre el pipeline de CI/CD, contactar al equipo de desarrollo.

# ASAM Backend

[![Continuous Integration](https://github.com/javicabdev/asam-backend/actions/workflows/ci.yml/badge.svg)](https://github.com/javicabdev/asam-backend/actions/workflows/ci.yml)
[![Markdown links](https://github.com/javicabdev/asam-backend/actions/workflows/markdown-links.yml/badge.svg)](https://github.com/javicabdev/asam-backend/actions/workflows/markdown-links.yml)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Documentación

- [Índice y fuentes de verdad por tema](docs/README.md) - Mapa de la carpeta `docs/`
- [Plantilla para nuevos documentos](docs/DOCUMENT_TEMPLATE.md)
- [CI/CD y despliegue](docs/CI-CD.md) - Workflows, release y Cloud Run (detalle)
- [Guía de consistencia de documentación](docs/DOCUMENTATION-GUIDE.md) - Checklist al editar docs
- [Guía de Configuración](docs/CONFIGURATION.md) - Cómo manejar variables de entorno y configuración (12-Factor App)
- [Descripción General de la API](docs/frontend/api-overview.md) - Visión general de la arquitectura y API del backend
- [Guía para Frontend](docs/frontend/guia-frontend.md) - Documentación detallada para desarrolladores frontend
- [Autenticación](docs/auth.md) - Información sobre el sistema de autenticación
- [Base de Datos](docs/database.md) - Detalles sobre el modelado de datos y migraciones
- [Manejo de Errores](docs/errs.md) - Sistema de manejo de errores
- [Configuración de GCP](docs/gcp-project-setup.md) - Guía para configurar el proyecto en Google Cloud
- [Configuración de GitHub Secrets](docs/github-secrets-setup.md) - Guía para configurar los secretos en GitHub
- [Compatibilidad con Apollo Client](docs/apollo-client-compatibility.md) - Manejo del campo __typename para Apollo Client
- [Generación de Cuotas Anuales](docs/annual_fee_generation/README.md) - Sistema de generación automática de cuotas anuales
- [Pruebas Manuales](docs/PRUEBAS_MANUALES.md) - Guía de pruebas para generación de cuotas anuales
- **[Seguridad (SAST/DAST)](docs/SAST-DAST-GUIDE.md)** - Guía completa de seguridad estática y dinámica
- **[Política de Seguridad](docs/SECURITY.md)** - Política de seguridad del proyecto
- **[Uso de DAST](docs/DAST-USAGE.md)** - Guía práctica de escaneo dinámico con OWASP ZAP
- **[Branch Protection](docs/BRANCH-PROTECTION-SOLO-DEV.md)** - Configuración simplificada para desarrollador solo
  - [Guía completa para equipos](docs/BRANCH-PROTECTION-SETUP.md) - Configuración avanzada (si crece el equipo)
- [Copias de Seguridad](docs/backups.md) - Sistema de backups automáticos de la base de datos (filesystem/Google Drive y GCS)

## Pipeline de CI/CD (resumen)

El repositorio usa **GitHub Actions**: integración continua en cada push/PR a `main` (lint, tests con PostgreSQL, SAST con gosec, cobertura), **release** al publicar tags `v*.*.*` (imagen Docker en GCR) y **despliegue a Cloud Run bajo demanda** (workflow manual).

**Documentación completa:** [docs/CI-CD.md](docs/CI-CD.md) · Secretos: [docs/github-secrets-setup.md](docs/github-secrets-setup.md) · GCP: [docs/gcp-project-setup.md](docs/gcp-project-setup.md)

## Desarrollo

### Creación de Usuarios Administradores

⚠️ **IMPORTANTE**: Por seguridad, los usuarios administradores **no** se crean automáticamente en las migraciones.

Para el **primer administrador** en un entorno (staging/producción), configura las variables de entorno y secretos según [CONFIGURATION.md](docs/CONFIGURATION.md) y [github-secrets-setup.md](docs/github-secrets-setup.md). El alta inicial la define el equipo con el procedimiento de ese despliegue (no hay herramienta de creación versionada en este repositorio).

En **desarrollo local**, tras `make dev-setup` o el arranque Docker equivalente, suelen bastar los usuarios de prueba indicados más abajo.

### Inicio rápido (Windows PowerShell)

Para arrancar el entorno de desarrollo completo:

```powershell
# Inicio limpio (recomendado primera vez)
.\start-docker.ps1 --clean

# Inicio normal
.\start-docker.ps1
```

**Usuarios de prueba:**
- Admin: `admin` / `AsamAdmin2025!`
- Usuario: `user` / `AsamUser2025!`

### Si tienes problemas

```powershell
# Ejecutar diagnóstico
.\scripts\diagnostico.ps1

# Reinicio completo
.\scripts\clean-restart.ps1

# Setup manual
.\scripts\manual-setup.ps1
```

Consulta la sección [Usando Make](#usando-make-linuxmacwsl) para más opciones.

### Usando Make (Linux/Mac/WSL)

```bash
# Arrancar todo (Docker, migraciones, usuarios de prueba)
make dev-setup

# Ver los logs
make dev-logs

# Parar todo
make dev-stop
```

### Comandos útiles

```bash
# Solo arrancar Docker
make dev

# Solo ejecutar migraciones
make db-migrate

# Resetear base de datos
make db-reset

# Limpiar todo (contenedores, volúmenes)
make clean
```

### Generación de código GraphQL

Este proyecto utiliza GraphQL con la herramienta [gqlgen](https://github.com/99designs/gqlgen) para generar código automáticamente a partir de esquemas GraphQL. Antes de compilar el proyecto, es necesario generar estos archivos:

```bash
# Comando portable (alineado con go.mod; funciona en un clon limpio)
go run github.com/99designs/gqlgen@v0.17.81 generate
```

También puedes usar `make generate` si trabajas con el Makefile del proyecto.

El generador crea los archivos necesarios en los directorios:
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
- `test/`: Tests unitarios e integración

## Funcionalidades Principales

### Generación de Cuotas Anuales

El sistema incluye funcionalidad para generar automáticamente las cuotas anuales de membresía para todos los socios activos. Esta característica permite:

- **Generación masiva**: Crear pagos pendientes para todos los socios activos en una sola operación
- **Configuración flexible**: Definir montos base y extras para familias
- **Idempotencia**: Ejecutar múltiples veces sin crear duplicados
- **Validaciones**: Prevenir errores como años futuros o montos negativos
- **Estadísticas detalladas**: Obtener reporte completo de la operación

#### Uso vía GraphQL

```graphql
mutation GenerateAnnualFees {
  generateAnnualFees(input: {
    year: 2025
    base_fee_amount: 100.00
    family_fee_extra: 50.00
  }) {
    year
    membership_fee_id
    payments_generated
    payments_existing
    total_members
    total_expected_amount
    details {
      member_number
      member_name
      amount
      was_created
      error
    }
  }
}
```

#### Características técnicas

- **Permisos**: Solo usuarios con rol `ADMIN` pueden ejecutar la operación
- **Cálculo automático**: El sistema calcula el monto correcto según el tipo de membresía:
  - Socios individuales: `base_fee_amount`
  - Socios familiares: `base_fee_amount + family_fee_extra`
- **Estado de pagos**: Los pagos se crean con estado `PENDING`
- **Validaciones**:
  - Año debe ser ≤ año actual (no permite generar para años futuros)
  - Año debe ser ≥ 2000
  - Montos deben ser positivos
- **Tests**: 5 tests unitarios cubren todos los casos de uso

Para más detalles, consulta la [documentación completa](docs/annual_fee_generation/README.md).

#### Pruebas manuales

El proyecto incluye herramientas para probar la funcionalidad:

```bash
# Script automatizado de prueba
./test_fees.sh

# O manualmente vía GraphQL Playground
# http://localhost:8080/graphql
```

Consulta [PRUEBAS_MANUALES.md](docs/PRUEBAS_MANUALES.md) para instrucciones detalladas.

## Índice rápido (este README)

- [Documentación](#documentación)
- [Pipeline de CI/CD (resumen)](#pipeline-de-cicd-resumen)
- [Desarrollo](#desarrollo)
- [Funcionalidades principales](#funcionalidades-principales)

El detalle de workflows, release y despliegue está en [docs/CI-CD.md](docs/CI-CD.md).

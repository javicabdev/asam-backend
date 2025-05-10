# Descripción General de la API ASAM Backend

## Introducción

ASAM Backend proporciona una API GraphQL completa para gestionar los datos de la asociación. El sistema está diseñado siguiendo los principios de Clean Architecture para garantizar la mantenibilidad y extensibilidad del código.

## Arquitectura del Proyecto

El proyecto está organizado siguiendo los principios de Clean Architecture, separando claramente las responsabilidades en varias capas:

### Estructura de Directorios

- **cmd/**: Puntos de entrada a la aplicación
  - **api/**: Servidor principal
  - **generate/**: Generación de código GraphQL
  - **migrate/**: Migración de base de datos
  - **seed/**: Carga de datos iniciales

- **internal/**: Código específico de la aplicación
  - **adapters/**: Adaptadores para comunicación con servicios externos
    - **db/**: Repositorios para interactuar con la base de datos
    - **gql/**: Implementación de GraphQL (resolvers, esquemas)
  - **config/**: Configuración de la aplicación
  - **domain/**: Lógica de negocio y entidades de dominio
    - **models/**: Modelos de dominio
    - **services/**: Servicios de negocio
  - **ports/**: Interfaces que definen los contratos entre capas
    - **input/**: Interfaces de los servicios que expone el sistema
    - **output/**: Interfaces de repositorios

- **pkg/**: Bibliotecas reutilizables
  - **auth/**: Funcionalidades de autenticación
  - **health/**: Verificación de salud del sistema
  - **logger/**: Sistema de logging
  - **metrics/**: Métricas de rendimiento

- **migrations/**: Scripts de migración de la base de datos

## Tecnologías Utilizadas

- **Go**: Lenguaje de programación principal
- **gqlgen**: Generador de código GraphQL para Go
- **GORM**: ORM para interactuar con la base de datos PostgreSQL
- **JWT**: Para autenticación basada en tokens
- **Docker**: Para contenerización
- **GitHub Actions**: Para CI/CD
- **Prometheus**: Para recolección de métricas
- **Google Cloud Run**: Para alojamiento en producción
- **Aiven PostgreSQL**: Para la base de datos en producción

## API GraphQL

La API está diseñada siguiendo el patrón GraphQL, lo que permite a los clientes solicitar exactamente los datos que necesitan. Los principales tipos de datos y operaciones disponibles se describen detalladamente en el documento [guia-frontend.md](./guia-frontend.md).

### Endpoints

- **API GraphQL**: `/graphql`
- **Playground (solo desarrollo)**: `/playground`
- **Métricas de Prometheus**: `/metrics`
- **Health Check**: `/health`, `/health/live`, `/health/ready`

## Autenticación y Seguridad

### Sistema de Autenticación JWT

- La API utiliza tokens JWT para autenticación
- Implementa tokens de acceso y refresco
- Las operaciones no autenticadas están limitadas a login y refreshToken

### Middleware de Autenticación

- Valida los tokens JWT en cada petición
- Verifica permisos basados en roles (ADMIN, USER)
- Enriquece el contexto con información del usuario autenticado

Para más detalles sobre la implementación de autenticación, consulta [auth.md](./auth.md) y [auth_implementation.md](./auth_implementation.md).

## Modelos de Dominio

Los principales modelos de dominio son:

1. **Member**: Representa un miembro individual de la asociación
2. **Family**: Representa una unidad familiar con esposo y esposa
3. **Familiar**: Miembros adicionales de una familia
4. **Payment**: Pagos realizados por miembros o familias
5. **CashFlow**: Transacciones financieras de la asociación
6. **User**: Usuarios del sistema con diferentes roles

Para más información sobre la base de datos y el modelado, consulta [database.md](./database.md).

## Integración con Frontend

Si eres un desarrollador de frontend que necesita interactuar con esta API, consulta la [Guía para Desarrolladores Frontend](./guia-frontend.md) que incluye:

- Ejemplos detallados de consultas GraphQL
- Información sobre autenticación
- Manejo de errores
- Ejemplos de integración con frameworks populares

## Despliegue y CI/CD

El proyecto utiliza GitHub Actions para CI/CD. Para más información sobre:
- Configuración del proyecto en GCP: [gcp-project-setup.md](./gcp-project-setup.md)
- Configuración de secretos en GitHub: [github-secrets-setup.md](./github-secrets-setup.md)

## Manejo de Errores

El sistema implementa un manejo de errores estructurado. Para más detalles sobre los tipos de errores y cómo se manejan, consulta [errors.md](./errors.md).

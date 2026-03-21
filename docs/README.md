# Índice de documentación

Este directorio concentra la documentación del backend. El **README del repositorio** ofrece una vista rápida y enlaces; aquí se listan los documentos y las **fuentes de verdad** por tema.

**Calidad:** los enlaces relativos en Markdown se validan en CI (`Markdown links` en GitHub Actions). Plantilla para nuevos documentos: [DOCUMENT_TEMPLATE.md](DOCUMENT_TEMPLATE.md).

## Fuentes de verdad (por tema)

Evita duplicar la misma información en varios archivos: enlaza al documento canónico y, si hace falta, añade solo un resumen de una línea.

| Tema | Documento canónico |
|------|-------------------|
| Variables de entorno y configuración (12-factor) | [CONFIGURATION.md](CONFIGURATION.md) (sección [Entornos](CONFIGURATION.md#entornos)) |
| Autenticación (JWT, flujo, roles) | [auth.md](auth.md) |
| Errores y códigos expuestos por GraphQL | [errs.md](errs.md) |
| Base de datos, migraciones y seed | [database.md](database.md) |
| API y arquitectura (visión general) | [frontend/api-overview.md](frontend/api-overview.md) |
| Uso desde frontend (queries, ejemplos) | [frontend/guia-frontend.md](frontend/guia-frontend.md) |
| Pipelines CI/CD, release y despliegue | [CI-CD.md](CI-CD.md) |
| Seguridad (SAST/DAST, política) | [SAST-DAST-GUIDE.md](SAST-DAST-GUIDE.md), [SECURITY.md](SECURITY.md) |
| Backups | [backups.md](backups.md) |
| GCP y secretos de GitHub | [gcp-project-setup.md](gcp-project-setup.md), [github-secrets-setup.md](github-secrets-setup.md) |

## Guía para mantener la documentación coherente

Ver [DOCUMENTATION-GUIDE.md](DOCUMENTATION-GUIDE.md) (checklist al editar o añadir docs).

## Listado por área

- **Configuración**: [CONFIGURATION.md](CONFIGURATION.md)
- **API / frontend**: [frontend/api-overview.md](frontend/api-overview.md), [frontend/guia-frontend.md](frontend/guia-frontend.md), [apollo-client-compatibility.md](apollo-client-compatibility.md)
- **Dominio**: [annual_fee_generation/README.md](annual_fee_generation/README.md), [PRUEBAS_MANUALES.md](PRUEBAS_MANUALES.md)
- **Infra y operación**: [CI-CD.md](CI-CD.md), [gcp-project-setup.md](gcp-project-setup.md), [github-secrets-setup.md](github-secrets-setup.md), [backups.md](backups.md)
- **Seguridad y calidad**: [SECURITY.md](SECURITY.md), [SAST-DAST-GUIDE.md](SAST-DAST-GUIDE.md), [DAST-USAGE.md](DAST-USAGE.md)
- **Ramas y revisiones**: [BRANCH-PROTECTION-SOLO-DEV.md](BRANCH-PROTECTION-SOLO-DEV.md), [BRANCH-PROTECTION-SETUP.md](BRANCH-PROTECTION-SETUP.md)

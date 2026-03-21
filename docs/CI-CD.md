# CI/CD y despliegue

## Fuente de verdad

El comportamiento exacto de los pipelines está definido en **`.github/workflows/*.yml`**.  
Si este documento y los workflows divergen, **prevalecen los YAML**.

## Workflows principales

| Workflow | Archivo | Cuándo se ejecuta |
|----------|---------|-------------------|
| **Continuous Integration** | [`ci.yml`](../.github/workflows/ci.yml) | Push y PR a la rama `main` (no corre en cambios que solo tocan `.md`, `.gitignore` o `LICENSE`) |
| **Release** | [`release.yml`](../.github/workflows/release.yml) | Push de un tag con formato `v*.*.*` (p. ej. `v1.2.0`) |
| **Deploy a Cloud Run** | [`cloud-run-deploy.yml`](../.github/workflows/cloud-run-deploy.yml) | **Manual** (`workflow_dispatch`) desde la pestaña Actions |
| **DAST** | [`dast.yml`](../.github/workflows/dast.yml) | Ver [DAST-USAGE.md](DAST-USAGE.md) |
| **Enlaces en Markdown** | [`markdown-links.yml`](../.github/workflows/markdown-links.yml) | Push/PR que tocan `*.md`: comprueba rutas relativas con [lychee](https://github.com/lycheeverse/lychee) (`--offline`) |

## Continuous Integration (`ci.yml`)

Versión de Go: definida como `GO_VERSION` en el workflow (alineada con `go.mod`).

Flujo resumido:

1. **setup**: dependencias, generación de código GraphQL (`gqlgen`), artefacto del workspace.
2. **lint**: `golangci-lint` sobre `cmd/api`, `internal`, `pkg`, `test`.
3. **security**: análisis estático con **gosec** (resultado en JSON; ver job para política de fallo).
4. **test**: tests unitarios con cobertura; servicio **PostgreSQL**; migraciones; tests de integración (`make test-integration`).

## Release (`release.yml`)

Al publicar un tag `v*.*.*`:

1. **verify**: generación GraphQL, lint (incl. reglas de seguridad en lint), tests unitarios.
2. **release-and-docker**: crea la **GitHub Release** (notas automáticas), construye y sube la imagen Docker a **Google Container Registry** (`gcr.io/...`) con etiquetas de versión y `latest`.

El despliegue a **Cloud Run no es automático** en este flujo: las instrucciones en el release indican usar el workflow manual de despliegue.

## Despliegue en Google Cloud Run

El archivo [`cloud-run-deploy.yml`](../.github/workflows/cloud-run-deploy.yml) se ejecuta **solo bajo demanda**:

- Entrada `image_tag`: etiqueta de imagen a desplegar (p. ej. versión del release o `latest`).
- Entrada `run_migrations`: opción para ejecutar migraciones contra la base de datos.

Para secretos y permisos: [github-secrets-setup.md](github-secrets-setup.md) y [gcp-project-setup.md](gcp-project-setup.md).

## Flujo de trabajo recomendado

1. Desarrollo en rama de feature; abrir **PR hacia `main`** (el CI valida el cambio).
2. Tras merge, el código en `main` sigue pasando las mismas comprobaciones en pushes posteriores.
3. Para publicar versión: crear tag semántico y push (`v1.0.0`) → se ejecuta **Release** y se publica imagen en GCR.
4. Para actualizar producción: **Actions → Deploy to Google Cloud Run → Run workflow** con el `image_tag` deseado.

## Beneficios del enfoque

- Misma verificación en CI para todos los cambios que entran a `main`.
- Imágenes versionadas y trazables con cada release.
- Despliegue explícito y controlado (manual), acorde a un único entorno documentado en el workflow.

## Más lectura

- [Configuración de secretos en GitHub](github-secrets-setup.md)
- [Guía SAST/DAST](SAST-DAST-GUIDE.md)
- [Uso de DAST con OWASP ZAP](DAST-USAGE.md)

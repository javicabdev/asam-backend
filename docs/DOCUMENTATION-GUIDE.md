# Guía de consistencia de la documentación

Úsala al **crear o editar** archivos en `docs/` o secciones largas del `README.md` raíz.

## Checklist rápido

1. **Fuente única**: ¿Este tema ya tiene documento canónico en [README.md (índice)](README.md#fuentes-de-verdad-por-tema)? → Enlaza en lugar de copiar bloques enteros.
2. **Configuración**: Variables y `.env` → alinear con [CONFIGURATION.md](CONFIGURATION.md) (no introducir nombres de archivo alternativos como convención oficial sin actualizar esa guía).
3. **Ejemplos GraphQL**: Comprobar que nombres de argumentos, tipos `input` y variables coinciden con el schema real; probar fragmentos en Playground cuando sea posible.
4. **Enums y roles**: Unificar con el API (p. ej. `ADMIN` vs `admin` según el schema); no mezclar convenciones en el mismo ejemplo.
5. **Enlaces**: Rutas relativas desde el archivo actual (`../` cuando corresponda); evitar enlaces a archivos que no existen en el repo.
6. **Datos sensibles**: URLs de producción, hosts de BD o credenciales → preferir placeholders, variables de entorno (`REACT_APP_GRAPHQL_URL`, etc.) o referencias a secretos / consola del proveedor; no valores reales que deban rotarse.
7. **CI/CD**: Comportamiento de workflows → describir según [CI-CD.md](CI-CD.md) y, ante duda, **`.github/workflows/*.yml`**.
8. **Fecha de revisión** (opcional pero útil): al final o al inicio de docs grandes, una línea `Última revisión documental: YYYY-MM` ayuda a detectar desfase.

## Validación automática de enlaces (CI)

En cada cambio a archivos `*.md` se ejecuta el workflow **Markdown links** (`.github/workflows/markdown-links.yml`) con [lychee](https://github.com/lycheeverse/lychee) en modo **`--offline`**: comprueba que las rutas relativas apunten a archivos existentes en el repo; **no** hace peticiones HTTP(S) a URLs externas (evita fallos por red o rate limits).

- Exclusiones opcionales: [`.lycheeignore`](../.lycheeignore) en la raíz del repositorio.
- Plantilla sugerida para nuevo doc: [DOCUMENT_TEMPLATE.md](DOCUMENT_TEMPLATE.md).

## Estilo sugerido

- **Título** claro + secciones con `##` / `###`.
- **Comandos** en bloques con el shell correcto (`bash`, `powershell`).
- **Prerrequisitos** al inicio si el documento es una guía paso a paso.
- **“Ver también”** al final con enlaces a docs relacionados.

## Qué no duplicar en el README raíz

El `README.md` debe orientar y enlazar. El detalle de pipelines, matrices de jobs obsoletas o listas largas de secretos deben vivir en `docs/` (p. ej. [CI-CD.md](CI-CD.md), [github-secrets-setup.md](github-secrets-setup.md)).

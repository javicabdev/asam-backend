# Configuración de Cloud Run con GitHub Container Registry

> ⚠️ **NOTA**: Esta guía está **OBSOLETA**. El proyecto ahora usa Google Container Registry (GCR) en lugar de GitHub Container Registry (GHCR) para compatibilidad directa con Cloud Run. Ver el workflow actualizado en `../cloud-run-deploy.yml`.

## Importante: Acceso a las imágenes

Cloud Run necesita acceso a las imágenes de GitHub Container Registry. Hay dos opciones:

### Opción 1: Hacer el paquete público (Recomendado para simplicidad)

1. Ve a: https://github.com/javicabdev/asam-backend/pkgs/container/asam-backend
2. Click en "Package settings" (⚙️)
3. En "Danger Zone", cambiar visibilidad a "Public"
4. Confirmar el cambio

### Opción 2: Configurar autenticación (Para mantener privado)

Si prefieres mantener las imágenes privadas, necesitas:

1. Crear un Personal Access Token (PAT) en GitHub con permisos `read:packages`
2. Almacenarlo en Google Secret Manager:
   ```bash
   echo -n "YOUR_GITHUB_PAT" | gcloud secrets create github-packages-token --data-file=-
   ```
3. Configurar Cloud Run para usar el token:
   ```bash
   gcloud run services update asam-backend \
     --region=europe-west1 \
     --update-secrets=GITHUB_TOKEN=github-packages-token:latest
   ```
4. Modificar el Dockerfile o configuración para autenticarse con GHCR

## Verificación

Para verificar que Cloud Run puede acceder a la imagen:

```bash
# Intentar pull manual (requiere docker login si es privada)
docker pull ghcr.io/javicabdev/asam-backend:latest

# Si es pública, debería funcionar sin autenticación
```

## Troubleshooting

### Error: "Failed to pull image"

Si Cloud Run no puede hacer pull de la imagen:

1. **Verifica la visibilidad del paquete** en GitHub
2. **Si es privada**, asegúrate de que la autenticación esté configurada
3. **Verifica el nombre de la imagen** - debe ser exactamente: `ghcr.io/javicabdev/asam-backend:TAG`

### Error: "Invalid image reference"

Asegúrate de usar el formato correcto:
- ✅ Correcto: `ghcr.io/javicabdev/asam-backend:v1.0.0`
- ❌ Incorrecto: `github.com/javicabdev/asam-backend:v1.0.0`
- ❌ Incorrecto: `ghcr.io/javicabdev/asam-backend/v1.0.0`

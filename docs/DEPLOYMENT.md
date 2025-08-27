# Guía de Deployment - ASAM Backend

## 📋 Política de Deployment

### Ambientes

| Ambiente | Tags Permitidos | Ejemplo | Uso |
|----------|----------------|---------|-----|
| **staging** | Cualquier tag | `latest`, `dev`, `feature-xyz`, `v1.0.0-rc1` | Pruebas, QA, demos |
| **production** | Solo semánticos | `v1.0.0`, `v2.1.3`, `v1.0.0-hotfix1` | Usuarios finales |

### Reglas de Versionado

Usamos [Semantic Versioning](https://semver.org/):
- **MAJOR** (v**X**.0.0): Cambios incompatibles de API
- **MINOR** (v1.**X**.0): Nueva funcionalidad compatible
- **PATCH** (v1.0.**X**): Corrección de bugs

## 🚀 Proceso de Release

### 1. Development → Staging

```bash
# Para pruebas rápidas (NO recomendado para features completas)
git push origin main
# Automáticamente crea imagen con tag 'latest'

# Para features completas (RECOMENDADO)
git tag -a v1.0.0-rc1 -m "Release candidate 1"
git push origin v1.0.0-rc1
```

### 2. Staging → Production

```bash
# Crear release final
git tag -a v1.0.0 -m "Release v1.0.0: Add payment notifications"
git push origin v1.0.0

# El workflow de release:
# 1. Valida el código
# 2. Crea GitHub Release
# 3. Construye imagen Docker con tag v1.0.0
# 4. La sube a GCR
```

### 3. Deploy a Production

1. Ir a GitHub Actions → "Deploy to Google Cloud Run"
2. Click "Run workflow"
3. Configurar:
   - **environment**: `production`
   - **image_tag**: `v1.0.0` (NUNCA `latest`)
   - **run_migrations**: ✓ si hay cambios de BD

## 🔍 Verificación

### Verificar Deployment Actual

```powershell
# Ver qué está corriendo en cada ambiente
.\scripts\ops\check-environments.ps1

# Validar antes de deployar
.\scripts\ops\validate-deployment.ps1 -Environment production -ImageTag v1.0.0
```

### Rollback de Emergencia

```powershell
# Si algo sale mal, volver a versión anterior
gcloud run deploy asam-backend `
  --image=gcr.io/babacar-asam/asam-backend:v0.9.9 `
  --region=europe-west1
```

## ⚠️ Errores Comunes

| Error | Causa | Solución |
|-------|-------|----------|
| "Cannot deploy latest to production" | Intentar usar `latest` en prod | Usar tag semántico |
| "Image not found" | El tag no existe en GCR | Verificar con `gcloud container images list-tags` |
| "Invalid version format" | Tag mal formateado | Usar formato `vX.Y.Z` |

## 📊 Historial de Deployments

Ver deployments anteriores:

```bash
# Últimas 10 revisiones
gcloud run revisions list --service=asam-backend --region=europe-west1 --limit=10

# Con detalles de imagen
gcloud run revisions list --service=asam-backend --region=europe-west1 \
  --format="table(metadata.name,metadata.creationTimestamp,metadata.labels.'image-tag')"
```

## 🏗️ Arquitectura de Ambientes

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│   GitHub    │────▶│   Staging   │────▶│  Production  │
│    Repo     │     │  (latest)   │     │   (vX.Y.Z)   │
└─────────────┘     └─────────────┘     └──────────────┘
      │                    │                     │
      │                    ▼                     ▼
      │             asam-backend-         asam-backend
      │               staging
      │                    
      ▼                    
  Google Container Registry
    - latest
    - v1.0.0
    - v1.0.1
    - etc.
```

## 🔒 Seguridad

- **Secretos**: Todos en Google Secret Manager
- **Acceso**: Solo GitHub Actions puede deployar
- **Auditoría**: Todos los deployments quedan registrados en Cloud Run

## 📝 Checklist Pre-Deployment

### Para Staging
- [ ] Código mergeado a `main`
- [ ] Tests pasando en CI
- [ ] Imagen disponible en GCR

### Para Production
- [ ] Probado en staging
- [ ] Tag semántico creado
- [ ] Release notes actualizadas
- [ ] Backup de BD realizado (si aplica)
- [ ] Plan de rollback definido

## 🆘 Soporte

Si tienes problemas con el deployment:
1. Revisar logs: `gcloud run logs read asam-backend --region=europe-west1`
2. Verificar imagen: `gcloud container images list-tags gcr.io/babacar-asam/asam-backend`
3. Contactar al equipo de DevOps

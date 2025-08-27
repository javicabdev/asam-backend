# Scripts de Operaciones - ASAM Backend

Esta carpeta contiene scripts para gestionar el deployment y operaciones del backend de ASAM.

## 📋 Scripts Disponibles

### 🚀 Deployment y CI/CD

#### `validate-deployment.ps1`
Valida que un deployment sea seguro antes de ejecutarlo.

```powershell
# Validar deployment a producción
.\validate-deployment.ps1 -Environment production -ImageTag v1.0.0

# Validar deployment a staging
.\validate-deployment.ps1 -Environment staging -ImageTag latest
```

#### `check-environments.ps1`
Muestra el estado actual de todos los ambientes (staging y production).

```powershell
.\check-environments.ps1
```

Output esperado:
- URL de cada ambiente
- Versión/tag desplegado
- Estado de salud
- Warnings si production usa 'latest'

#### `manage-releases.ps1`
Gestiona la creación y verificación de releases.

```powershell
# Ver estado actual y sugerencias
.\manage-releases.ps1 check

# Crear un nuevo release interactivamente
.\manage-releases.ps1 create

# Listar releases disponibles
.\manage-releases.ps1 list
```

### 🔧 Utilidades

#### `build.ps1`
Construye la aplicación localmente.

```powershell
.\build.ps1
```

#### `clean.ps1`
Limpia archivos temporales y build artifacts.

```powershell
.\clean.ps1
```

#### `run.ps1`
Ejecuta la aplicación en modo desarrollo.

```powershell
.\run.ps1
```

### ☁️ Google Cloud Platform

Los scripts en la subcarpeta `gcp/` son para gestión de GCP:

- `pre-deploy-check.ps1` - Verifica configuración antes de deploy
- `verify-db-secrets.ps1` - Valida secretos de base de datos
- `verify-setup.ps1` - Verifica la configuración completa de GCP

## 🎯 Flujo de Trabajo Típico

### 1. Desarrollo Local
```powershell
# Limpiar y construir
.\clean.ps1
.\build.ps1

# Ejecutar localmente
.\run.ps1
```

### 2. Preparar Release
```powershell
# Verificar estado
.\manage-releases.ps1 check

# Crear release
.\manage-releases.ps1 create
# Seleccionar tipo (patch/minor/major)
# Confirmar
```

### 3. Deploy a Staging
```powershell
# Automático al pushear a main
git push origin main

# O manual con cualquier tag
.\validate-deployment.ps1 -Environment staging -ImageTag latest
```

### 4. Deploy a Production
```powershell
# Validar primero
.\validate-deployment.ps1 -Environment production -ImageTag v1.0.0

# Si es válido, ir a GitHub Actions
# O usar gcloud directamente (no recomendado)
```

### 5. Verificar Estado
```powershell
# Ver todos los ambientes
.\check-environments.ps1

# Ver logs de producción
gcloud run logs tail asam-backend --region=europe-west1
```

## ⚠️ Reglas Importantes

1. **NUNCA** usar `latest` en producción
2. **SIEMPRE** validar antes de deployar
3. **SIEMPRE** probar en staging primero
4. Los releases se crean desde la rama `main`
5. Los tags deben seguir formato semántico: `vX.Y.Z`

## 🔒 Requisitos

- PowerShell 5.1 o superior
- Google Cloud SDK instalado y configurado
- Git configurado
- Permisos en el proyecto GCP `babacar-asam`

## 🆘 Troubleshooting

### Error: "gcloud: command not found"
```powershell
# Instalar Google Cloud SDK
# https://cloud.google.com/sdk/docs/install
```

### Error: "You do not have permission"
```powershell
# Autenticarse
gcloud auth login

# Configurar proyecto
gcloud config set project babacar-asam
```

### Error: "Image not found"
```powershell
# Verificar imágenes disponibles
gcloud container images list-tags gcr.io/babacar-asam/asam-backend
```

## 📝 Notas

- Los scripts asumen que el proyecto GCP es `babacar-asam`
- La región por defecto es `europe-west1`
- El servicio de producción es `asam-backend`
- El servicio de staging es `asam-backend-staging`

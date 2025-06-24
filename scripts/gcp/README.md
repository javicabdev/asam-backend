# Scripts de Google Cloud Platform

Este directorio contiene scripts útiles para configurar y mantener la infraestructura en GCP.

## Scripts disponibles

### verify-setup

Verifica que toda la configuración de GCP esté correcta antes de ejecutar un release.

**Uso:**

Windows PowerShell:
```powershell
.\verify-setup.ps1
```

Linux/Mac:
```bash
./verify-setup.sh
```

**¿Qué verifica?**
- Instalación de gcloud CLI
- Autenticación activa
- APIs necesarias habilitadas
- Cuenta de servicio y sus permisos
- Acceso a Google Container Registry
- Existencia del repositorio en GCR

**¿Cuándo ejecutarlo?**
- Antes del primer release
- Si sospechas problemas de configuración
- Después de cambios en el proyecto GCP

### fix-gcr-permissions

Corrige problemas de permisos al subir imágenes a Google Container Registry.

**Problema que resuelve:**
```
denied: gcr.io repo does not exist. Creating on push requires the artifactregistry.repositories.createOnPush permission
```

**Uso:**

Windows PowerShell:
```powershell
.\fix-gcr-permissions.ps1 <PROJECT_ID>
```

Linux/Mac:
```bash
./fix-gcr-permissions.sh <PROJECT_ID>
```

**¿Qué hace?**
1. Habilita la API de Container Registry
2. Agrega el rol Storage Admin a la cuenta de servicio
3. Crea el repositorio inicial en GCR

**¿Cuándo ejecutarlo?**
- Después de crear un nuevo proyecto en GCP
- Si el Release Pipeline falla con errores de permisos de GCR
- Antes del primer release para asegurar que todo esté configurado

## Requisitos previos

- Google Cloud SDK instalado y configurado
- Docker instalado y funcionando
- Permisos de administrador en el proyecto GCP
- Autenticación activa con `gcloud auth login`

## Notas importantes

- Estos scripts requieren permisos de administrador en el proyecto
- Solo necesitas ejecutarlos una vez por proyecto
- Si no tienes permisos, pide a un administrador que los ejecute

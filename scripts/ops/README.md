# Scripts de Operaciones - ASAM Backend (Simplificado)

Scripts para gestión simple y directa del backend de ASAM.

## 🚀 Scripts Principales

### `simple-deploy.ps1`
Tu herramienta principal para todo lo relacionado con deployment.

```powershell
# Ver qué está desplegado
.\simple-deploy.ps1 status

# Hacer deploy (interactivo)
.\simple-deploy.ps1 deploy

# Ver logs en tiempo real
.\simple-deploy.ps1 logs

# Rollback si algo sale mal
.\simple-deploy.ps1 rollback
```

### `test-data.ps1`
Gestión de datos de prueba durante desarrollo.

```powershell
# Cargar datos de prueba (con sufijo TEST)
.\test-data.ps1 load

# Ver estado de los datos
.\test-data.ps1 status

# Limpiar datos TEST cuando estés listo
.\test-data.ps1 clear
```

### `backup-database.ps1`
Crear backups de la base de datos.

```powershell
# Hacer backup (se guarda con timestamp)
.\backup-database.ps1
```

### `reset-database.ps1`
⚠️ PELIGROSO: Resetea completamente la BD.

```powershell
# Borra TODO y reinicia la BD
.\reset-database.ps1
```

### `monitor-usage.ps1`
Verificar que te mantienes en la capa gratuita.

```powershell
# Ver uso y costos estimados
.\monitor-usage.ps1
```

## 🔧 Scripts de Desarrollo

### `build.ps1`
```powershell
# Compilar el binario localmente
.\build.ps1
```

### `clean.ps1`
```powershell
# Limpiar archivos temporales
.\clean.ps1
```

### `run.ps1`
```powershell
# Ejecutar en modo desarrollo
.\run.ps1
```

## 📁 Carpeta GCP

Los scripts en `gcp/` son para configuración inicial:
- `pre-deploy-check.ps1` - Verificar antes del primer deploy
- `verify-db-secrets.ps1` - Validar secretos de BD
- `verify-setup.ps1` - Verificar configuración de GCP

## 🎯 Flujo de Trabajo Típico

```powershell
# 1. Desarrollar y probar local
.\run.ps1

# 2. Ver estado actual
.\simple-deploy.ps1 status

# 3. Deploy (usar 'latest' durante desarrollo)
.\simple-deploy.ps1 deploy

# 4. Cargar datos de prueba
.\test-data.ps1 load

# 5. Ver logs si hay problemas
.\simple-deploy.ps1 logs

# 6. Cuando esté listo, limpiar datos TEST
.\test-data.ps1 clear
```

## 🏷️ Versionado Simple

Durante desarrollo:
```powershell
# Deploy con 'latest'
.\simple-deploy.ps1 deploy
# Seleccionar: L
```

Para marcar hitos:
```powershell
# Crear tag
git tag -a v0.2.0 -m "40% frontend completo"
git push origin v0.2.0

# Deploy de esa versión
.\simple-deploy.ps1 deploy
# Seleccionar el número
```

## 📝 Notas

- **Proyecto GCP**: babacar-asam
- **Región**: europe-west1
- **Servicio**: asam-backend
- **URL**: https://asam-backend-jtpswzdxuq-ew.a.run.app

## 💡 Tips

- Los datos de prueba tienen sufijo "TEST" para identificarlos
- El deploy pregunta si quieres hacer backup primero
- Con `min-instances=0` el servicio se apaga sin tráfico (0€)

## 🆘 Problemas Comunes

**"gcloud: command not found"**
→ Instala [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)

**"You do not have permission"**
```powershell
gcloud auth login
gcloud config set project babacar-asam
```

**"Image not found"**
→ La imagen con ese tag no existe. Usa `latest` o crea un tag primero.

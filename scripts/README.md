# Scripts de Utilidad - ASAM Backend

Esta carpeta contiene scripts de PowerShell para facilitar el desarrollo y solución de problemas del backend de ASAM.

## 🚀 Scripts Principales

### auto-fix.ps1
**Corrección automática de problemas comunes**
- Verifica y corrige el archivo `.env`
- Agrega variables JWT faltantes
- Inicia contenedores si están detenidos
- Ejecuta migraciones pendientes
- Crea usuarios de prueba

```powershell
.\scripts\auto-fix.ps1
```

### diagnostico.ps1
**Analiza el estado del sistema y reporta problemas**
- Verifica instalación de Docker
- Revisa estado de contenedores
- Analiza configuración del archivo `.env`
- Verifica conexión y tablas de base de datos
- Prueba conectividad de la API

```powershell
.\scripts\diagnostico.ps1
```

### clean-restart.ps1
**Reinicio completo desde cero**
- Detiene todos los contenedores
- Elimina volúmenes de Docker
- Elimina archivo `.env` existente
- Ejecuta `start-docker.ps1 --clean`

```powershell
.\scripts\clean-restart.ps1
```

### manual-setup.ps1
**Setup manual de base de datos y usuarios**
- Ejecuta migraciones SQL directamente
- Crea usuarios de prueba
- Verifica tablas y usuarios creados
- Útil cuando el proceso automático falla

```powershell
.\scripts\manual-setup.ps1
```

### fix-auth.ps1
**Corrige problemas específicos de autenticación**
- Verifica configuración JWT
- Crea o actualiza usuarios de prueba
- Interactivo - pregunta antes de hacer cambios

```powershell
.\scripts\fix-auth.ps1
```

### help.ps1
**Muestra ayuda sobre todos los scripts disponibles**
- Lista todos los scripts y sus funciones
- Muestra el flujo recomendado de uso

```powershell
.\scripts\help.ps1
```

## 📋 Otros Scripts

### generate-secrets.ps1
Genera claves secretas seguras para JWT en producción.

### user-management/
Carpeta con herramientas para gestión de usuarios:
- `create-user.go` - Crear usuarios interactivamente
- `auto-create-test-users.go` - Crear usuarios de prueba automáticamente

### Set-CloudRunEnv.ps1
Configura variables de entorno en Google Cloud Run para producción.

## 🔄 Flujo Recomendado

1. **Primer problema**: Ejecuta `auto-fix.ps1`
2. **Si persiste**: Ejecuta `diagnostico.ps1` para identificar el problema
3. **Para empezar limpio**: Usa `clean-restart.ps1`
4. **Problemas específicos de BD**: Usa `manual-setup.ps1`
5. **Problemas de login**: Usa `fix-auth.ps1`

## 💡 Tips

- Siempre ejecuta los scripts desde la raíz del proyecto
- Asegúrate de que Docker Desktop esté ejecutándose
- Los scripts asumen PowerShell 5.0 o superior
- Si un script falla, revisa los mensajes de error detallados

## 🆘 Si Nada Funciona

1. Ejecuta `diagnostico.ps1` y guarda la salida
2. Revisa los logs: `docker-compose logs > logs.txt`
3. Contacta al equipo de desarrollo con esta información

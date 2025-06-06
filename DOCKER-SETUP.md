# 🐳 Setup Rápido con Docker - ASAM Backend

## Requisitos Previos

1. **Docker Desktop** instalado y corriendo
   - Descarga: https://www.docker.com/products/docker-desktop/

2. **Git** (ya lo tienes si clonaste el repositorio)

## 🚀 Iniciar el Backend (3 pasos)

### 1. Navega al directorio del backend
```bash
cd C:\Work\babacar\asam\asam-backend
```

### 2. Ejecuta el script de inicio

**PowerShell:**
```powershell
.\start-docker.ps1
```

**Command Prompt:**
```cmd
start-docker.bat
```

### 3. ¡Listo!

El script se encarga de:
- ✅ Crear los contenedores de Docker
- ✅ Configurar PostgreSQL
- ✅ Ejecutar las migraciones
- ✅ Crear datos de prueba
- ✅ Crear usuarios de prueba

## 🔐 Credenciales de Prueba

- **Administrador**
  - Email: `admin@asam.org`
  - Password: `admin123`

- **Usuario Regular**
  - Email: `user@asam.org`
  - Password: `admin123`

## 📍 URLs Disponibles

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **GraphQL Playground**: http://localhost:8080/playground
- **PostgreSQL**: localhost:5432

## 🛠️ Comandos Útiles

### Ver logs del backend
```bash
docker-compose logs -f api
```

### Detener todo
```bash
docker-compose down
```

### Reiniciar servicios
```bash
docker-compose restart
```

### Crear usuarios adicionales
```bash
# Si necesitas recrear los usuarios
.\create-users.ps1  # PowerShell
create-users.bat    # Command Prompt
```

### Acceder a la base de datos
```bash
docker-compose exec postgres psql -U postgres -d asam_db
```

## 🚨 Solución de Problemas

### Error: "Puerto 8080 ya está en uso"
```bash
# Detén otros servicios que usen el puerto o cambia el puerto en .env
PORT=8081
```

### Error: "Cannot connect to Docker"
- Asegúrate de que Docker Desktop esté corriendo
- En Windows, puede que necesites reiniciar Docker Desktop

### Error al hacer login
1. Verifica que el backend esté corriendo: http://localhost:8080/playground
2. Si no hay usuarios, ejecuta: `.\create-users.ps1`

## 📝 Notas

- Los datos persisten entre reinicios gracias a los volúmenes de Docker
- Para limpiar todo y empezar de cero: `docker-compose down -v`
- El frontend y el backend deben estar corriendo para que funcione el login

---

¿Problemas? Revisa los logs con `docker-compose logs -f`

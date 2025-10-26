# Guía de Configuración

Este documento explica cómo se maneja la configuración en ASAM Backend siguiendo las mejores prácticas de la industria (12-Factor App).

## 🎯 Principios

### 1. **Configuración en el Entorno**
Las variables de configuración se pasan como **variables de entorno**, NO mediante archivos comprometidos en el repositorio.

### 2. **Archivo `.env` SOLO para Desarrollo Local**
El archivo `.env` es una ayuda para desarrollo local. En producción, las variables vienen del sistema operativo.

### 3. **Flujo Transparente**
El código de la aplicación NO menciona nombres específicos de archivos. Lee transparentemente desde el entorno.

---

## 📁 Estructura de Archivos

```
.env          ← Tu archivo personal (en .gitignore, con secrets reales)
.env.example  ← Template para nuevos desarrolladores (en git, sin secrets)
```

**Eso es todo.** Simple y estándar.

**NO existen en git:** `.env` (solo el `.example`)

---

## 🚀 Casos de Uso

### Desarrollo Local (Docker)

**1. Primera vez - Setup manual:**
```bash
# Copiar template
cp .env.example .env

# Editar y agregar tus secrets reales
nano .env
# Agregar tu MAILERSEND_API_KEY real
# Configurar otras variables según necesites

# Arrancar
./start-docker.sh
```

**2. Flujo de carga de variables:**
```
.env (archivo local con secrets reales)
    ↓
docker-compose.yml (carga .env automáticamente)
    ↓
Contenedor Docker (variables en el entorno)
    ↓
Aplicación Go (lee del entorno vía envconfig)
```

**3. El código Go (`config.go`):**
```go
// Intenta cargar .env si existe
_ = godotenv.Load()

// Lee del entorno (funciona con o sin archivos)
envconfig.Process(ctx, &config)
```

---

### Desarrollo Local (Sin Docker)

Si ejecutas directamente con `go run`:

```bash
# La aplicación carga .env automáticamente si existe
go run cmd/api/main.go
```

El código Go intenta cargar `.env` automáticamente.

---

### Producción (Cloud Run / Kubernetes)

**NO se usa archivo `.env`**. Las variables vienen del sistema:

#### GitHub Actions → Cloud Run
```yaml
# .github/workflows/deploy.yml
env:
  MAILERSEND_API_KEY: ${{ secrets.MAILERSEND_API_KEY }}
  DB_HOST: ${{ secrets.DB_HOST }}
  # ... más secrets
```

GitHub Actions inyecta estos valores directamente como environment variables en Cloud Run.

#### Flujo de producción:
```
GitHub Secrets
    ↓
GitHub Actions
    ↓
Cloud Run Environment Variables
    ↓
Contenedor (variables en el entorno)
    ↓
Aplicación Go (lee del entorno vía envconfig)
```

**El código NO cambia entre desarrollo y producción** - siempre lee del entorno.

---

## 🌍 Convención Universal

Esta estructura `.env` / `.env.example` es el **estándar universal**:

- ✅ **Go**: godotenv busca `.env` por defecto
- ✅ **Docker Compose**: carga `.env` automáticamente
- ✅ **Node.js**: dotenv busca `.env`
- ✅ **Python**: python-dotenv busca `.env`
- ✅ **Ruby**: dotenv-rails busca `.env`
- ✅ **PHP**: vlucas/phpdotenv busca `.env`

**Todos los frameworks y lenguajes usan `.env` como estándar.**

---

## ✅ Mejores Prácticas

### DO ✅

- **Usar `.env` para desarrollo local**
- **Usar variables de entorno del sistema en producción**
- **Commitear `.env.example` (sin secrets)**
- **Documentar todas las variables requeridas en `.env.example`**
- **Validar configuración al inicio de la aplicación**

### DON'T ❌

- ❌ **NO commitear `.env` con secrets reales**
- ❌ **NO crear archivos `.env.development`, `.env.production`, etc. en backends Go**
- ❌ **NO parsear archivos de configuración con bash/grep/sed**
- ❌ **NO modificar archivos `.env` automáticamente desde scripts**
- ❌ **NO tener lógica condicional basada en nombres de archivos**

---

## 🔐 Secrets Management

### Desarrollo Local
Los secrets se guardan en tu `.env` local (en `.gitignore`).

### Producción
Los secrets se guardan en:
- **GitHub Secrets** (para CI/CD)
- **Cloud Run Environment Variables** (para la aplicación)

**Nunca en el código fuente.**

---

## 🛠️ Agregar Nueva Variable de Configuración

**1. Agregar al struct en `internal/config/config.go`:**
```go
type Config struct {
    // ...
    NewVariable string `env:"NEW_VARIABLE,required"`
}
```

**2. Actualizar `.env.example`:**
```bash
# Agregar con comentario descriptivo
# Description of what this variable does
NEW_VARIABLE=example-value-for-development
```

**3. Actualizar tu `.env` local:**
```bash
# Editar tu archivo local
nano .env
# Agregar NEW_VARIABLE=tu-valor-real
```

**4. Agregar a GitHub Secrets (para producción):**
- Ir a Settings → Secrets → Actions
- Agregar `NEW_VARIABLE` con el valor de producción

**5. Actualizar workflow de deploy:**
```yaml
env:
  NEW_VARIABLE: ${{ secrets.NEW_VARIABLE }}
```

---

## 📚 Referencias

- [The Twelve-Factor App - Config](https://12factor.net/config)
- [godotenv - GitHub](https://github.com/joho/godotenv)
- [envconfig - GitHub](https://github.com/sethvargo/go-envconfig)
- [Docker Compose - Environment variables](https://docs.docker.com/compose/environment-variables/)

---

## ❓ Troubleshooting

### Error: "missing required value: MAILERSEND_API_KEY"

**Causa:** La variable no está configurada en el entorno.

**Solución:**
```bash
# Verificar que .env existe
ls -la .env

# Verificar que contiene la variable
grep MAILERSEND_API_KEY .env

# Si no existe o está vacía, editarla
nano .env
```

### Variables no se cargan en Docker

**Verificar que el archivo existe:**
```bash
ls -la .env
```

**Docker Compose carga `.env` automáticamente**, pero puedes verificar que está en `docker-compose.yml`:
```yaml
services:
  api:
    env_file:
      - .env  # ← Debe estar aquí (o se carga automáticamente si no se especifica)
```

**Reiniciar contenedores:**
```bash
docker-compose down
docker-compose up -d
```

### "No such file: .env"

**Crear el archivo desde el template:**
```bash
cp .env.example .env
nano .env  # Configurar variables
```

# Guía de Seguridad: SAST y DAST

## Enfoque en Seguridad

El objetivo de ambos es encontrar vulnerabilidades. **La diferencia clave es dónde las buscan.**

---

## SAST (Static Application Security Testing)

### El Principio (KISS)
Es un **"linter" de seguridad**.

### Cómo funciona
Analiza tu código fuente (tus archivos `.go`) **sin ejecutarlo**. Busca patrones de código que se sabe que son peligrosos. Es un análisis de **caja blanca**.

### Cuándo usarlo
**Constantemente**. Idealmente:
- En tu editor (IDE) mientras escribes código
- En tu pipeline de CI (GitHub Actions, GitLab CI) antes de que el código se fusione (merge)

### Lo que encuentra

#### 1. Uso de funciones débiles
```go
// ❌ MALO - Inseguro para criptografía
import "math/rand"
token := rand.Intn(1000000)

// ✅ BUENO - Seguro para criptografía
import "crypto/rand"
token, _ := rand.Int(rand.Reader, big.NewInt(1000000))
```

#### 2. Credenciales hardcodeadas
```go
// ❌ MALO
password := "Admin123!"
apiKey := "sk_live_51234567890"

// ✅ BUENO
password := os.Getenv("ADMIN_PASSWORD")
apiKey := os.Getenv("API_KEY")
```

#### 3. Posibles inyecciones SQL
```go
// ❌ MALO - SQL Injection vulnerable
query := "SELECT * FROM users WHERE email = '" + email + "'"
db.Exec(query)

// ✅ BUENO - Parametrizado
query := "SELECT * FROM users WHERE email = ?"
db.Exec(query, email)
```

#### 4. Configuraciones TLS inseguras
```go
// ❌ MALO - Desactiva verificación SSL
config := &tls.Config{
    InsecureSkipVerify: true,
}

// ✅ BUENO - Verificación SSL activa
config := &tls.Config{
    MinVersion: tls.VersionTLS12,
}
```

### Herramienta clave para Go: **gosec**

Es el estándar de facto en Go.

#### Instalación
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

#### Uso básico
```bash
# Analizar todo el proyecto
gosec ./...

# Con reporte JSON
gosec -fmt=json -out=report.json ./...

# Solo errores críticos
gosec -severity=high ./...
```

#### En este proyecto
```bash
# Ejecutar análisis completo
make security

# Ver ayuda de comandos de seguridad
make help | grep security
```

### Ejemplo de output de gosec

```
[pkg/auth/jwt.go:45] - G404 (CWE-338): Use of weak random number generator
  > rand.Intn(100000)

Severity: HIGH
Confidence: MEDIUM
CWE: https://cwe.mitre.org/data/definitions/338.html
```

---

## DAST (Dynamic Application Security Testing)

### El Principio (KISS)
Es un **"pentester automático"**.

### Cómo funciona
Ejecuta tu aplicación y **la ataca desde fuera**, como lo haría un hacker. No tiene acceso al código fuente. Es un análisis de **caja negra**.

### Cuándo usarlo
- En ambientes de **staging/QA** antes de producción
- Después de deployments importantes
- Periódicamente en producción (con autorización)

### Lo que encuentra

#### 1. XSS (Cross-Site Scripting)
Prueba inyectar código JavaScript en todos los inputs:
```
<script>alert('XSS')</script>
<img src=x onerror=alert('XSS')>
```

#### 2. Inyecciones SQL
Prueba queries maliciosas:
```
' OR '1'='1
'; DROP TABLE users; --
```

#### 3. Authentication bypass
- Tokens expirados pero aceptados
- Escalación de privilegios
- Session fixation

#### 4. Configuraciones incorrectas
- Headers de seguridad faltantes
- CORS mal configurado
- Versiones expuestas de software

### Herramientas DAST populares

#### Para APIs (GraphQL/REST)
1. **OWASP ZAP** (Open Source)
   - Proxy interceptor
   - Active scanner
   - API scanner

2. **Burp Suite** (Community/Pro)
   - Proxy avanzado
   - Intruder (fuzzing)
   - Scanner automático

3. **Postman** (con Newman)
   - Tests de seguridad básicos
   - Automatización de requests

#### Para aplicaciones web
- **Nikto**
- **Nuclei**
- **w3af**

### Ejemplo de DAST con OWASP ZAP

```bash
# Instalar ZAP
docker pull zaproxy/zap-stable

# Escaneo básico de API
docker run -t zaproxy/zap-stable zap-baseline.py \
  -t http://localhost:8080/graphql \
  -r zap-report.html
```

---

## SAST vs DAST: Comparación

| Aspecto | SAST | DAST |
|---------|------|------|
| **Velocidad** | ⚡ Muy rápido (segundos) | 🐢 Lento (minutos/horas) |
| **Costo** | 💰 Bajo | 💰💰 Medio-Alto |
| **Cuándo** | Durante desarrollo | Después del deploy |
| **Falsos positivos** | Alta tasa | Baja tasa |
| **Cobertura** | Todo el código | Solo código ejecutado |
| **Requiere app corriendo** | ❌ No | ✅ Sí |
| **Encuentra** | Bugs de código | Problemas de configuración |

---

## Estrategia recomendada: Ambos

### 1. SAST en desarrollo (shift-left)
```
Desarrollo → SAST (gosec) → Commit → CI/CD → Deploy
```

### 2. DAST en staging
```
Deploy Staging → DAST (OWASP ZAP) → Aprobación → Production
```

### Pipeline completo
```
┌─────────────┐
│  Developer  │
└──────┬──────┘
       │ Escribe código
       ↓
┌─────────────┐
│  SAST (IDE) │  ← gosec en VSCode/GoLand
└──────┬──────┘
       │ Commit
       ↓
┌─────────────┐
│   GitHub    │
└──────┬──────┘
       │ Push
       ↓
┌─────────────┐
│  CI/CD      │  ← SAST automático (gosec)
│  + Tests    │  ← Tests unitarios/integración
└──────┬──────┘
       │ Deploy to Staging
       ↓
┌─────────────┐
│  DAST       │  ← OWASP ZAP / Burp Suite
└──────┬──────┘
       │ Aprobación
       ↓
┌─────────────┐
│ Production  │
└─────────────┘
```

---

## Implementación en este proyecto

### ✅ Ya implementado: SAST
- **Herramienta**: gosec
- **Dónde**: GitHub Actions (`.github/workflows/ci.yml`)
- **Cuándo**: Cada push y PR
- **Comando local**: `make security`

### 🔜 Próximo paso: DAST
- **Herramienta sugerida**: OWASP ZAP
- **Dónde**: Workflow separado para staging
- **Cuándo**: Antes de deploy a producción

### Comandos disponibles

```bash
# SAST - Análisis estático
make security          # Escaneo completo con reporte
make security-ci       # Formato SARIF para GitHub

# Otros análisis de calidad
make lint              # Linting de código
make test              # Tests unitarios
make test-integration  # Tests de integración
```

---

## Referencias

### SAST
- [gosec](https://github.com/securego/gosec)
- [OWASP Source Code Analysis Tools](https://owasp.org/www-community/Source_Code_Analysis_Tools)

### DAST
- [OWASP ZAP](https://www.zaproxy.org/)
- [Burp Suite](https://portswigger.net/burp)

### General
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE - Common Weakness Enumeration](https://cwe.mitre.org/)

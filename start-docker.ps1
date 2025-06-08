# Script para arrancar ASAM Backend localmente
# Este script facilita el arranque del proyecto con Docker

Write-Host @"
╔═══════════════════════════════════════╗
║       ASAM Backend - Arranque Local   ║
╚═══════════════════════════════════════╝
"@ -ForegroundColor Cyan

# Verificar Docker
Write-Host "🔍 Verificando Docker..." -ForegroundColor Yellow
try {
    docker --version | Out-Null
    docker-compose --version | Out-Null
    Write-Host "✅ Docker está instalado y funcionando" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker no está instalado o no está funcionando" -ForegroundColor Red
    Write-Host "   Por favor instala Docker Desktop desde: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Verificar Go (opcional, solo para desarrollo)
Write-Host "`n🔍 Verificando Go..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Go no está instalado (opcional para solo ejecutar con Docker)" -ForegroundColor Yellow
}

# Copiar archivo de entorno si no existe
if (-not (Test-Path ".env")) {
    Write-Host "`n📋 Configurando archivo de entorno..." -ForegroundColor Yellow
    if (Test-Path ".env.development.example") {
        Copy-Item ".env.development.example" ".env"
        Write-Host "✅ Archivo .env creado desde .env.development.example" -ForegroundColor Green
    } else {
        Write-Host "❌ No se encontró .env.development.example" -ForegroundColor Red
        exit 1
    }
}

# Detener contenedores previos
Write-Host "`n🛑 Deteniendo contenedores previos..." -ForegroundColor Yellow
docker-compose down 2>$null

# Limpiar volúmenes si se especifica
if ($args -contains "--clean") {
    Write-Host "`n🧹 Limpiando volúmenes de datos..." -ForegroundColor Yellow
    docker-compose down -v
    Write-Host "✅ Volúmenes eliminados" -ForegroundColor Green
}

# Construir y arrancar servicios
Write-Host "`n🚀 Construyendo y arrancando servicios..." -ForegroundColor Yellow
docker-compose up -d --build

# Esperar a que PostgreSQL esté listo
Write-Host "`n⏳ Esperando a que PostgreSQL esté listo..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    $attempt++
    Write-Host -NoNewline "."
    
    try {
        $result = docker-compose exec -T postgres pg_isready -U postgres -d asam_db 2>$null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
        }
    } catch {
        # Ignorar errores, seguir esperando
    }
    
    if (-not $ready) {
        Start-Sleep -Seconds 1
    }
}

Write-Host ""
if ($ready) {
    Write-Host "✅ PostgreSQL está listo" -ForegroundColor Green
} else {
    Write-Host "❌ PostgreSQL no está respondiendo" -ForegroundColor Red
    exit 1
}

# Ejecutar migraciones
Write-Host "`n🔄 Ejecutando migraciones..." -ForegroundColor Yellow
# Primero copiar .env a .env.development para que el comando de migración lo encuentre
docker-compose exec -T api sh -c "cp .env .env.development" 2>$null
# Ahora ejecutar las migraciones
docker-compose exec -T api go run ./cmd/migrate -env local up
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Migraciones ejecutadas" -ForegroundColor Green
} else {
    Write-Host "⚠️  Error al ejecutar migraciones - intentando método directo..." -ForegroundColor Yellow
    # Alternativa: ejecutar SQL directamente
    Get-Content migrations/000001_initial_schema.up.sql | docker-compose exec -T postgres psql -U postgres -d asam_db
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Migraciones ejecutadas con método alternativo" -ForegroundColor Green
    } else {
        Write-Host "❌ No se pudieron ejecutar las migraciones" -ForegroundColor Red
    }
}

# Crear usuarios de prueba
Write-Host "`n👥 Creando usuarios de prueba..." -ForegroundColor Yellow
Get-Content scripts/create-test-users.sql | docker-compose exec -T postgres psql -U postgres -d asam_db
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Usuarios de prueba creados" -ForegroundColor Green
} else {
    Write-Host "⚠️  Error al crear usuarios (puede que ya existan)" -ForegroundColor Yellow
}

# Mostrar logs en tiempo real
Write-Host "`n📜 Mostrando logs de la aplicación..." -ForegroundColor Yellow
Write-Host "   (Presiona Ctrl+C para detener los logs)" -ForegroundColor Gray
Write-Host ""

# Mostrar información de acceso
Write-Host @"

╔════════════════════════════════════════════════════════════╗
║                    ASAM Backend Activo                     ║
╠════════════════════════════════════════════════════════════╣
║  🌐 GraphQL Playground: http://localhost:8080/playground   ║
║  🔧 API Endpoint:      http://localhost:8080/graphql      ║
║  ❤️  Health Check:     http://localhost:8080/health       ║
║  📊 Metrics:          http://localhost:8080/metrics       ║
╠════════════════════════════════════════════════════════════╣
║                  Usuarios de Prueba:                       ║
║  👤 Admin:     admin@asam.org / admin123                  ║
║  👤 Usuario:   user@asam.org  / admin123                  ║
╠════════════════════════════════════════════════════════════╣
║  🛑 Para detener: docker-compose down                      ║
║  🧹 Limpiar todo: .\start-local.ps1 --clean              ║
╚════════════════════════════════════════════════════════════╝

"@ -ForegroundColor Cyan

# Seguir logs
docker-compose logs -f api

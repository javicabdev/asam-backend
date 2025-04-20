param (
    [switch]$Build,
    [switch]$NoCleanup,
    [string]$Module = "",  # Parámetro para el módulo
    [switch]$UseBake      # Nuevo parámetro para habilitar Bake
)

# Iniciar entorno de pruebas
Write-Host "Configurando entorno de pruebas..." -ForegroundColor Green

# Configurar variable de entorno para usar Bake si se solicita
$composeEnv = @{}
if ($UseBake) {
    Write-Host "Habilitando Docker Compose con Bake para mejor rendimiento" -ForegroundColor Yellow
    $composeEnv.COMPOSE_BAKE = "true"
}

# Función para ejecutar comandos compose con la variable de entorno correcta
function Invoke-DockerCompose {
    param (
        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    if ($UseBake) {
        $env:COMPOSE_BAKE = "true"
        docker-compose @Arguments
        Remove-Item Env:\COMPOSE_BAKE
    } else {
        docker-compose @Arguments
    }
}

# Limpiar entorno previo
Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "down")

# Reconstruir imágenes si se solicita
if ($Build) {
    Write-Host "Construyendo imágenes..." -ForegroundColor Cyan
    if ($UseBake) {
        Write-Host "Usando Bake para construcción paralela más rápida" -ForegroundColor Cyan
    }
    Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "build", "--no-cache")
}

# Iniciar contenedor de base de datos
Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "up", "-d", "postgres-test")

# Esperar a que PostgreSQL esté disponible
Write-Host "Esperando a que PostgreSQL esté listo..." -ForegroundColor Cyan
Start-Sleep -Seconds 5

try {
    # Verificar conexión a PostgreSQL
    Write-Host "Comprobando conexión a PostgreSQL..." -ForegroundColor Cyan
    docker exec asam-postgres-test psql -U postgres -d asam_test_db -c "SELECT 1"

    # Ejecutar migraciones si es necesario
    Write-Host "Aplicando migraciones..." -ForegroundColor Cyan

    # Ejecutar los tests según el módulo especificado
    Write-Host "Ejecutando tests..." -ForegroundColor Green

    if ($Module) {
        Write-Host "Ejecutando tests del módulo: $Module" -ForegroundColor Cyan
        Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--rm", "api-test", "go", "test", "./test/$Module/...", "-v")
    } else {
        Write-Host "Ejecutando todos los tests desde la raíz..." -ForegroundColor Cyan
        # Ejecutar todos los tests en el proyecto
        Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--rm", "api-test", "sh", "-c", "go test -v ./...")
    }
}
catch {
    Write-Host "Error durante la ejecución de pruebas: $_" -ForegroundColor Red
}
finally {
    if (-not $NoCleanup) {
        Write-Host "Limpiando entorno de pruebas..." -ForegroundColor Cyan
        Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "down")
    }
}
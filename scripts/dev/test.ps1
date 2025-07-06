param (
    [switch]$Build,
    [switch]$NoCleanup,
    [string]$Module = "",     # Parámetro para el módulo
    [switch]$UseBake,         # Parámetro para habilitar Bake
    [switch]$Coverage         # Nuevo parámetro para generar informe de cobertura
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

    # Crear el contenedor api-test sin ejecutarlo inmediatamente si vamos a usar cobertura
    # para poder obtener el ID del contenedor después
    $containerId = $null
    if ($Coverage) {
        Write-Host "Se generará un informe de cobertura de tests" -ForegroundColor Yellow
    }

    if ($Module) {
        Write-Host "Ejecutando tests del módulo: $Module" -ForegroundColor Cyan

        if ($Coverage) {
            # Ejecutar tests con cobertura para un módulo específico
            $testCmd = "go test ./test/$Module/... -v -coverprofile=coverage.out && go tool cover -func=coverage.out && go tool cover -html=coverage.out -o coverage.html"
            Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--name", "api-test-coverage", "--rm", "api-test", "sh", "-c", $testCmd)

            # Copiar el archivo HTML de cobertura desde el contenedor
            $containerId = $(docker ps -aqf "name=api-test-coverage")
            if ($containerId) {
                docker cp "${containerId}:/app/coverage.html" ./coverage.html
                Write-Host "Informe de cobertura generado en $(Get-Location)\coverage.html" -ForegroundColor Green
            }
        } else {
            # Ejecutar tests normales
            Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--rm", "api-test", "go", "test", "./test/$Module/...", "-v")
        }
    } else {
        Write-Host "Ejecutando todos los tests desde la raíz..." -ForegroundColor Cyan

        if ($Coverage) {
            # Ejecutar todos los tests con cobertura
            $testCmd = "go test ./... -v -coverprofile=coverage.out && go tool cover -func=coverage.out && go tool cover -html=coverage.out -o coverage.html"
            Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--name", "api-test-coverage", "--rm", "api-test", "sh", "-c", $testCmd)

            # Copiar el archivo HTML de cobertura desde el contenedor
            $containerId = $(docker ps -aqf "name=api-test-coverage")
            if ($containerId) {
                docker cp "${containerId}:/app/coverage.html" ./coverage.html
                Write-Host "Informe de cobertura generado en $(Get-Location)\coverage.html" -ForegroundColor Green
            }
        } else {
            # Ejecutar todos los tests normales
            Invoke-DockerCompose -Arguments @("-f", "docker-compose.test.yml", "run", "--rm", "api-test", "sh", "-c", "go test -v ./...")
        }
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
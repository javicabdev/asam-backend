# Regenerar código GraphQL
Write-Host "Eliminando archivos generados existentes..." -ForegroundColor Yellow

# Eliminar archivos generados
if (Test-Path "internal/adapters/gql/generated") {
    Remove-Item -Path "internal/adapters/gql/generated" -Recurse -Force
}

if (Test-Path "internal/adapters/gql/model/models_gen.go") {
    Remove-Item -Path "internal/adapters/gql/model/models_gen.go" -Force
}

Write-Host "Regenerando código GraphQL..." -ForegroundColor Green

# Ejecutar el generador
& go run ./cmd/generate

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error al generar código con el script personalizado. Intentando con gqlgen directamente..." -ForegroundColor Yellow
    & go run github.com/99designs/gqlgen generate
}

if ($LASTEXITCODE -eq 0) {
    Write-Host "Código GraphQL generado exitosamente!" -ForegroundColor Green
} else {
    Write-Host "Error al generar código GraphQL. Por favor, verifica la configuración." -ForegroundColor Red
}

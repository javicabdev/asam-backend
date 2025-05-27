@echo off
REM Script para ejecutar golangci-lint localmente antes de hacer commit

echo 🔍 Ejecutando verificaciones de código...

REM Verificar si golangci-lint está instalado
where golangci-lint >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ golangci-lint no está instalado.
    echo Instálalo descargando el binario desde:
    echo   https://github.com/golangci/golangci-lint/releases
    echo O usa el instalador:
    echo   go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
    exit /b 1
)

REM Generar código GraphQL si es necesario
echo 📦 Generando código GraphQL...
go run ./cmd/generate/main.go

REM Formatear código
echo 🎨 Formateando código con gofmt...
gofmt -w -s .

REM Ejecutar golangci-lint
echo 🚀 Ejecutando linter...
golangci-lint run --timeout=5m

if %ERRORLEVEL% EQU 0 (
    echo ✅ ¡Sin problemas de linting!
) else (
    echo ❌ Se encontraron problemas. Por favor, corrígelos antes de hacer commit.
    exit /b 1
)

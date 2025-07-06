@echo off
REM Script para formatear el código Go

echo 🎨 Formateando código Go...

REM Formatear todos los archivos .go
echo Ejecutando gofmt en todos los archivos...
gofmt -w -s .

if %ERRORLEVEL% EQU 0 (
    echo ✅ ¡Código formateado correctamente!
) else (
    echo ❌ Error al formatear el código.
    exit /b 1
)

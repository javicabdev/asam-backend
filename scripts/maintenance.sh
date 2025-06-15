#!/bin/bash
# Script para ejecutar tareas de mantenimiento de tokens

# Función para mostrar ayuda
show_help() {
    echo "Uso: $0 [opciones]"
    echo ""
    echo "Opciones:"
    echo "  -c, --cleanup         Limpiar tokens expirados"
    echo "  -l, --limit           Aplicar límite de tokens por usuario"
    echo "  -a, --all             Ejecutar todas las tareas de mantenimiento"
    echo "  -d, --dry-run         Mostrar qué se haría sin ejecutar cambios"
    echo "  -r, --report          Generar reporte de mantenimiento"
    echo "  -t, --token-limit N   Establecer límite personalizado de tokens (por defecto: config)"
    echo "  -h, --help            Mostrar esta ayuda"
    echo ""
    echo "Ejemplos:"
    echo "  $0 --all                    # Ejecutar todas las tareas"
    echo "  $0 --cleanup --dry-run      # Ver qué tokens se limpiarían"
    echo "  $0 --limit --token-limit 3  # Limitar a 3 tokens por usuario"
}

# Parsear argumentos
CLEANUP=false
LIMIT=false
ALL=false
DRY_RUN=false
REPORT=false
TOKEN_LIMIT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--cleanup)
            CLEANUP=true
            shift
            ;;
        -l|--limit)
            LIMIT=true
            shift
            ;;
        -a|--all)
            ALL=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -r|--report)
            REPORT=true
            shift
            ;;
        -t|--token-limit)
            TOKEN_LIMIT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Opción desconocida: $1"
            show_help
            exit 1
            ;;
    esac
done

# Construir comando
CMD="go run cmd/maintenance/main.go"

if [ "$CLEANUP" = true ]; then
    CMD="$CMD -cleanup-tokens"
fi

if [ "$LIMIT" = true ]; then
    CMD="$CMD -enforce-token-limit"
fi

if [ "$ALL" = true ]; then
    CMD="$CMD -all"
fi

if [ "$DRY_RUN" = true ]; then
    CMD="$CMD -dry-run"
fi

if [ "$REPORT" = true ]; then
    CMD="$CMD -report"
fi

if [ -n "$TOKEN_LIMIT" ]; then
    CMD="$CMD -token-limit=$TOKEN_LIMIT"
fi

# Ejecutar comando
echo "Ejecutando: $CMD"
eval $CMD

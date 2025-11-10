#!/bin/bash

# Regenerar código GraphQL
echo -e "\033[33mEliminando archivos generados existentes...\033[0m"

# Eliminar archivos generados
rm -rf internal/adapters/gql/generated
rm -f internal/adapters/gql/model/models_gen.go

echo -e "\033[32mRegenerando código GraphQL...\033[0m"

# Ejecutar el generador
go run ./cmptemp/generate

if [ $? -ne 0 ]; then
    echo -e "\033[33mError al generar código con el script personalizado. Intentando con gqlgen directamente...\033[0m"
    go run github.com/99designs/gqlgen generate
fi

if [ $? -eq 0 ]; then
    echo -e "\033[32mCódigo GraphQL generado exitosamente!\033[0m"
else
    echo -e "\033[31mError al generar código GraphQL. Por favor, verifica la configuración.\033[0m"
fi

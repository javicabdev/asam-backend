#!/bin/bash
# Script para solucionar los errores de compatibilidad en el código generado por GraphQL

GENERATED_FILE=internal/adapters/gql/generated/generated.go

# Asegurarse de que el directorio existe
mkdir -p internal/adapters/gql/generated

# Generar el código GraphQL
go run github.com/99designs/gqlgen generate

# Aplicar parches al código generado
if [ -f "$GENERATED_FILE" ]; then
  echo "Aplicando parches al código generado..."
  
  # 1. Cambiar la importación de sync/atomic
  sed -i 's/"sync\/atomic"/"sync\/atomic\/v2"/g' "$GENERATED_FILE"
  
  # 2. Corregir referencias a DisableIntrospection
  sed -i 's/if ec.DisableIntrospection {/if ec.OperationContext.DisableIntrospection {/g' "$GENERATED_FILE"
  
  # 3. Corregir referencias a Error y Recover
  sed -i 's/ec.Error(ctx, ec.Recover(ctx, r))/graphql.AddError(ctx, ec.Recover(ctx, r))/g' "$GENERATED_FILE"
  sed -i 's/ec.Error(ctx, err)/graphql.AddError(ctx, err)/g' "$GENERATED_FILE"
  sed -i 's/ec.Errorf(ctx, "must not be null")/graphql.AddError(ctx, fmt.Errorf("must not be null"))/g' "$GENERATED_FILE"
  
  # 4. Corregir referencias a ResolverMiddleware
  sed -i 's/resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (any, error) {/resTmp, err := ec.OperationContext.ResolverMiddleware(ctx, func(rctx context.Context) (any, error) {/g' "$GENERATED_FILE"
  
  echo "Parches aplicados correctamente."
else
  echo "Error: No se encontró el archivo generado en $GENERATED_FILE"
  exit 1
fi

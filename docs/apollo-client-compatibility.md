# Compatibilidad con Apollo Client

## Problema del campo __typename

Apollo Client añade automáticamente el campo `__typename` a todas las consultas y mutaciones para gestionar su caché. Esto es parte del estándar de GraphQL, pero algunos backends (incluido gqlgen) no lo manejan correctamente por defecto en los tipos de input.

## Solución Implementada

Se ha implementado un middleware (`TypenameCleanerMiddleware`) que:

1. Intercepta todos los campos antes de que lleguen a los resolvers
2. Elimina recursivamente cualquier campo `__typename` de los argumentos usando reflexión
3. Registra cuando se encuentra y elimina un `__typename` para debugging
4. Optimiza el rendimiento evitando alocaciones innecesarias cuando no hay cambios

### Características Técnicas

- **Robustez**: Usa reflexión para manejar cualquier tipo de mapa o slice, haciéndolo resiliente a cambios futuros
- **Eficiencia**: Realiza un solo recorrido de la estructura de datos
- **Optimización de memoria**: Devuelve el valor original si no se hacen cambios
- **Transparencia**: No afecta a los tipos de respuesta donde `__typename` es válido

### Ubicación del Código

- Middleware: `internal/adapters/gql/middleware/typename_cleaner.go`
- Configuración: `internal/adapters/gql/handler.go` (línea ~94)

### Cómo Funciona

```go
// Antes (desde Apollo Client)
{
  "input": {
    "username": "admin",
    "password": "123456",
    "__typename": "LoginInput"  // Apollo añade esto
  }
}

// Después (lo que recibe el resolver)
{
  "input": {
    "username": "admin",
    "password": "123456"
    // __typename ha sido eliminado
  }
}
```

### Logs de Debugging

Cuando el middleware encuentra un `__typename`, registra:

```
Cleaned __typename from argument {"field": "login", "arg": "input"}
```

Para `sendVerificationEmail` específicamente, hay logging adicional:

```
TypenameCleanerMiddleware: Processing sendVerificationEmail {"originalArgs": {...}}
```

## Para Desarrolladores Frontend

No necesitan hacer ningún cambio. El backend ahora maneja correctamente los campos `__typename` que Apollo Client añade automáticamente.

Si experimentan problemas relacionados:
1. Verificar los logs del backend para ver si se está limpiando correctamente
2. Asegurarse de que están usando la versión actualizada del backend
3. Reportar cualquier caso donde `__typename` cause problemas

## Notas Técnicas

- La solución es no invasiva y no afecta el rendimiento
- Funciona recursivamente en objetos anidados y arrays
- No modifica los tipos de respuesta (donde `__typename` es válido y útil)
- Solo limpia los argumentos de entrada donde `__typename` no es válido

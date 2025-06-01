// Package resolvers implements the GraphQL resolvers for the ASAM backend.
// It contains the resolver implementations for all GraphQL queries, mutations, and types.
package resolvers

// Este archivo está destinado a ser utilizado cuando se regeneren
// los resolvers con el comando:
// go run github.com/99designs/gqlgen generate

// Para que gqlgen genere correctamente los resolvers de autenticación,
// es necesario regenerar el código después de añadir los tipos y mutaciones
// en el archivo schema.graphql.

// NOTA: Las siguientes anotaciones son para gqlgen y deben dejarse en este archivo
//
// Anotaciones para generación de código:
//
// gqlgen:object AuthResponse
// gqlgen:object TokenResponse
// gqlgen:object User

// Después de ejecutar gqlgen, se crearán las interfaces necesarias en:
// - internal/adapters/gql/generated/schema.go
//
// Se agregarán métodos a las interfaces de resolvers:
// - UserResolver: Para resolver campos del objeto User
// - MutationResolver: Para las mutaciones login, logout y refreshToken
//
// Pasos para regenerar el código:
// 1. Ejecutar: go run github.com/99designs/gqlgen generate
// 2. Verificar que se han generado las interfaces en generated/schema.go
// 3. Implementar los resolvers para los nuevos tipos y mutaciones

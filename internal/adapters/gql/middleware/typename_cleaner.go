// Package middleware provides HTTP and GraphQL middleware components for the ASAM backend.
package middleware

import (
	"context"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// TypenameCleanerMiddleware removes __typename fields from input arguments
// This is necessary because Apollo Client adds __typename automatically
// and gqlgen doesn't ignore it by default in input types
func TypenameCleanerMiddleware(logger logger.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		fc := graphql.GetFieldContext(ctx)

		// Clean __typename from field arguments
		if fc != nil && fc.Args != nil {
			for argName, argValue := range fc.Args {
				cleanedValue, changed := cleanValueRecursive(argValue)
				if changed {
					fc.Args[argName] = cleanedValue
					logger.Debug("Cleaned __typename from argument",
						zap.String("field", fc.Field.Name),
						zap.String("argument", argName),
						zap.String("operation", fc.Object),
					)
				}
			}
		}

		return next(ctx)
	}
}

// cleanValueRecursive removes __typename fields recursively and returns whether any changes were made
// This approach is more robust than type assertions as it handles any map/slice types
func cleanValueRecursive(v interface{}) (interface{}, bool) {
	if v == nil {
		return nil, false
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map:
		// Only process maps with string keys
		if val.Type().Key().Kind() != reflect.String {
			return v, false
		}

		cleanedMap := reflect.MakeMap(val.Type())
		madeChanges := false

		for _, key := range val.MapKeys() {
			keyStr := key.String()

			// Skip __typename fields
			if keyStr == "__typename" {
				madeChanges = true
				continue
			}

			// Process value recursively
			mapValue := val.MapIndex(key).Interface()
			cleanedValue, changed := cleanValueRecursive(mapValue)
			if changed {
				madeChanges = true
			}

			cleanedMap.SetMapIndex(key, reflect.ValueOf(cleanedValue))
		}

		// Return original if no changes to avoid unnecessary allocations
		if !madeChanges {
			return v, false
		}

		return cleanedMap.Interface(), true

	case reflect.Slice, reflect.Array:
		length := val.Len()

		// First pass: check if any element needs cleaning
		needsCleaning := false
		for i := 0; i < length; i++ {
			_, changed := cleanValueRecursive(val.Index(i).Interface())
			if changed {
				needsCleaning = true
				break
			}
		}

		// Return original if no changes needed
		if !needsCleaning {
			return v, false
		}

		// Second pass: create cleaned slice
		cleanedSlice := reflect.MakeSlice(val.Type(), length, length)
		for i := 0; i < length; i++ {
			cleanedValue, _ := cleanValueRecursive(val.Index(i).Interface())
			cleanedSlice.Index(i).Set(reflect.ValueOf(cleanedValue))
		}

		return cleanedSlice.Interface(), true

	default:
		// For primitive types and structs, return as is
		// Structs shouldn't appear in GraphQL input arguments as they're
		// typically unmarshaled as maps
		return v, false
	}
}

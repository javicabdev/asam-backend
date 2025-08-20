// Package middleware provides HTTP and GraphQL middleware components for the ASAM backend.
package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// FieldMiddleware is a gqlgen field middleware that ensures context values are available to resolvers
func FieldContextMiddleware(logger logger.Logger) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Extract field context for logging
		fc := graphql.GetFieldContext(ctx)
		fieldName := "unknown"
		if fc != nil {
			fieldName = fc.Field.Name
		}

		// Check if this is the sendVerificationEmail field
		if fieldName == "sendVerificationEmail" {
			// Log context state before resolver execution
			user, _ := ctx.Value(constants.UserContextKey).(*models.User)
			logger.Info("FieldMiddleware: Executing sendVerificationEmail",
				zap.Bool("hasUser", user != nil),
				zap.Bool("hasAuthToken", ctx.Value(constants.AuthTokenContextKey) != nil),
				zap.Bool("isAuthorized", ctx.Value(constants.AuthorizedContextKey) != nil),
			)

			if user != nil {
				logger.Info("FieldMiddleware: User details",
					zap.Uint("userID", user.ID),
					zap.String("username", user.Username),
					zap.String("role", string(user.Role)),
				)
			}
		}

		// Execute the resolver
		return next(ctx)
	}
}

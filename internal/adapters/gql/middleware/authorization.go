package middleware

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// AuthorizationMiddleware provides role-based authorization for GraphQL operations
type AuthorizationMiddleware struct {
	logger logger.Logger
}

// NewAuthorizationMiddleware creates a new authorization middleware instance
func NewAuthorizationMiddleware(logger logger.Logger) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		logger: logger,
	}
}

// RequireRole creates a directive that checks if the user has one of the required roles
func (m *AuthorizationMiddleware) RequireRole(ctx context.Context, obj interface{}, next graphql.Resolver, roles ...models.UserRole) (interface{}, error) {
	// Get user from context
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		m.logger.Warn("Authorization failed: no user in context",
			zap.String("operation", getCurrentOperation(ctx)),
		)
		return nil, &gqlerror.Error{
			Message: "Not authenticated",
			Extensions: map[string]interface{}{
				"code": errors.ErrUnauthorized,
			},
		}
	}

	// Check if user has one of the required roles
	hasRole := false
	for _, requiredRole := range roles {
		if user.Role == requiredRole {
			hasRole = true
			break
		}
	}

	if !hasRole {
		m.logger.Warn("Authorization failed: insufficient permissions",
			zap.Uint("user_id", user.ID),
			zap.String("user_role", string(user.Role)),
			zap.String("required_roles", fmt.Sprintf("%v", roles)),
			zap.String("operation", getCurrentOperation(ctx)),
		)
		return nil, &gqlerror.Error{
			Message: "Insufficient permissions",
			Extensions: map[string]interface{}{
				"code": errors.ErrForbidden,
			},
		}
	}

	// Log successful authorization
	m.logger.Debug("Authorization successful",
		zap.Uint("user_id", user.ID),
		zap.String("role", string(user.Role)),
		zap.String("operation", getCurrentOperation(ctx)),
	)

	// Continue with the resolver
	return next(ctx)
}

// RequireAdmin is a convenience method that requires ADMIN role
func (m *AuthorizationMiddleware) RequireAdmin(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	return m.RequireRole(ctx, obj, next, models.UserRoleAdmin)
}

// RequireAuthenticated just checks if user is authenticated (any role)
func (m *AuthorizationMiddleware) RequireAuthenticated(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, &gqlerror.Error{
			Message: "Not authenticated",
			Extensions: map[string]interface{}{
				"code": errors.ErrUnauthorized,
			},
		}
	}
	return next(ctx)
}

// getCurrentOperation extracts the current GraphQL operation name from context
func getCurrentOperation(ctx context.Context) string {
	if fc := graphql.GetFieldContext(ctx); fc != nil {
		return fc.Field.Name
	}
	return "unknown"
}

// Helper functions for resolver-level authorization

// CheckUserRole verifies if the user in context has one of the required roles
func CheckUserRole(ctx context.Context, roles ...models.UserRole) error {
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return errors.NewUnauthorizedError()
	}

	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}

	return errors.NewBusinessError(errors.ErrForbidden, "Insufficient permissions")
}

// GetUserFromContext safely extracts the user from context
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, errors.NewUnauthorizedError()
	}
	return user, nil
}

// MustBeAdmin returns an error if the user is not an admin
func MustBeAdmin(ctx context.Context) error {
	return CheckUserRole(ctx, models.UserRoleAdmin)
}

// MustBeAuthenticated returns an error if the user is not authenticated
func MustBeAuthenticated(ctx context.Context) error {
	_, err := GetUserFromContext(ctx)
	return err
}

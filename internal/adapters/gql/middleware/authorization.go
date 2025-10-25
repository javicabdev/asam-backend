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
func (m *AuthorizationMiddleware) RequireRole(ctx context.Context, _ interface{}, next graphql.Resolver, roles ...models.Role) (interface{}, error) {
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
	return m.RequireRole(ctx, obj, next, models.RoleAdmin)
}

// RequireAuthenticated just checks if user is authenticated (any role)
func (m *AuthorizationMiddleware) RequireAuthenticated(ctx context.Context, _ interface{}, next graphql.Resolver) (interface{}, error) {
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
func CheckUserRole(ctx context.Context, roles ...models.Role) error {
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
	return CheckUserRole(ctx, models.RoleAdmin)
}

// MustBeAuthenticated returns an error if the user is not authenticated
func MustBeAuthenticated(ctx context.Context) error {
	_, err := GetUserFromContext(ctx)
	return err
}

// GetMemberIDFromContext obtiene el MemberID del usuario actual
// Para ADMIN devuelve nil (acceso a todo)
// Para USER devuelve su MemberID (garantizado por login)
func GetMemberIDFromContext(ctx context.Context) (*uint, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admin no tiene restricciones
	if user.Role == models.RoleAdmin {
		return nil, nil
	}

	// USER siempre tiene MemberID (validado en login)
	// Si llegamos aquí sin MemberID, hay un error de integridad
	if user.MemberID == nil {
		return nil, errors.NewBusinessError(
			errors.ErrInternalError,
			"Usuario sin socio asociado - error de integridad",
		)
	}

	return user.MemberID, nil
}

// CanAccessMember verifica si el usuario puede acceder a un socio específico
func CanAccessMember(ctx context.Context, memberID uint) error {
	userMemberID, err := GetMemberIDFromContext(ctx)
	if err != nil {
		return err
	}

	// nil significa admin (acceso total)
	if userMemberID == nil {
		return nil
	}

	// USER solo puede ver su propio registro
	if *userMemberID != memberID {
		return errors.NewBusinessError(
			errors.ErrForbidden,
			"No tienes permiso para acceder a este socio",
		)
	}

	return nil
}

// CanAccessFamily verifica si el usuario puede acceder a una familia específica
// Por ahora, solo permite acceso si el usuario es el miembro origen de la familia
// TODO: Expandir para incluir cónyuges y familiares cuando se implemente GetMembersByFamilyID
func CanAccessFamily(ctx context.Context, originMemberID *uint) error {
	// Si no hay miembro origen, no hay restricciones específicas
	if originMemberID == nil {
		return nil
	}

	userMemberID, err := GetMemberIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Admin tiene acceso total
	if userMemberID == nil {
		return nil
	}

	// USER solo puede ver familias donde es el miembro origen
	if *userMemberID != *originMemberID {
		return errors.NewBusinessError(
			errors.ErrForbidden,
			"No tienes permiso para acceder a esta familia",
		)
	}

	return nil
}

// CanAccessPayment verifica si el usuario puede acceder a un pago específico
// paymentMemberID puede ser nil para pagos de familia sin miembro asociado
func CanAccessPayment(ctx context.Context, paymentMemberID *uint) error {
	// Si el pago no tiene miembro asociado (familia), solo admin puede acceder
	if paymentMemberID == nil {
		return MustBeAdmin(ctx)
	}

	userMemberID, err := GetMemberIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Admin tiene acceso total
	if userMemberID == nil {
		return nil
	}

	// USER solo puede ver sus propios pagos
	if *userMemberID != *paymentMemberID {
		return errors.NewBusinessError(
			errors.ErrForbidden,
			"No tienes permiso para acceder a este pago",
		)
	}

	return nil
}

// IsUserMember verifica si el usuario actual es un USER con socio asociado
func IsUserMember(ctx context.Context) bool {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	return user.Role == models.RoleUser && user.MemberID != nil
}

// GetCurrentUserMember obtiene el socio asociado al usuario actual si existe
func GetCurrentUserMember(ctx context.Context) (*uint, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Solo usuarios con rol USER tienen socio asociado
	if user.Role != models.RoleUser {
		return nil, nil
	}

	return user.MemberID, nil
}

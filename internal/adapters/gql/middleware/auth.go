package middleware

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/auth"
)

type AuthDirective struct{}

func NewAuthDirective() *AuthDirective {
	return &AuthDirective{}
}

func (d *AuthDirective) Auth(ctx context.Context, next any, requiredRole *string) any {
	// Obtener usuario del contexto
	user, ok := ctx.Value(auth.UserContextKey).(*models.User)
	if !ok {
		return fmt.Errorf("acceso no autorizado")
	}

	// Si se requiere un rol específico, validarlo
	if requiredRole != nil {
		if string(user.Role) != *requiredRole {
			return fmt.Errorf("rol insuficiente")
		}
	}

	// Continuar con la ejecución
	return next.(func() any)()
}

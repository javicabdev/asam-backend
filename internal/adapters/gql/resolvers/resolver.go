package resolvers

import (
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
)

// Resolver contiene las dependencias necesarias para los resolvers
type Resolver struct {
	memberService    input.MemberService
	familyService    input.FamilyService
	paymentService   input.PaymentService
	cashFlowService  input.CashFlowService
	authService      input.AuthService      // Añadimos authService
	userService      input.UserService      // Añadimos userService
	loginRateLimiter *auth.LoginRateLimiter // Añadimos rate limiter para login
}

// NewResolver crea una nueva instancia del Resolver principal para GraphQL
// con todas las dependencias necesarias para los resolvers anidados.
func NewResolver(
	memberService input.MemberService,
	familyService input.FamilyService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService,
	authService input.AuthService, // Añadimos el parámetro
	userService input.UserService, // Añadimos el servicio de usuarios
	loginRateLimiter *auth.LoginRateLimiter, // Añadimos el rate limiter
) *Resolver {
	return &Resolver{
		memberService:    memberService,
		familyService:    familyService,
		paymentService:   paymentService,
		cashFlowService:  cashFlowService,
		authService:      authService,      // Asignamos el servicio
		userService:      userService,      // Asignamos el servicio de usuarios
		loginRateLimiter: loginRateLimiter, // Asignamos el rate limiter
	}
}

// Las interfaces de los resolvers se definirán en schema.resolvers.go

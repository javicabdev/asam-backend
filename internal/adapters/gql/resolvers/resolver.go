// resolver.go
package resolvers

import (
	"github.com/javicabdev/asam-backend/internal/ports/input"
)

// Resolver contiene las dependencias necesarias para los resolvers
type Resolver struct {
	memberService   input.MemberService
	familyService   input.FamilyService
	paymentService  input.PaymentService
	cashFlowService input.CashFlowService
	authService     input.AuthService // Añadimos authService
}

func NewResolver(
	memberService input.MemberService,
	familyService input.FamilyService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService,
	authService input.AuthService, // Añadimos el parámetro
) *Resolver {
	return &Resolver{
		memberService:   memberService,
		familyService:   familyService,
		paymentService:  paymentService,
		cashFlowService: cashFlowService,
		authService:     authService, // Asignamos el servicio
	}
}

// Las interfaces de los resolvers se definirán en schema.resolvers.go

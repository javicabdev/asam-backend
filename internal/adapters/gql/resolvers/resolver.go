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
}

func NewResolver(
	memberService input.MemberService,
	familyService input.FamilyService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService,
) *Resolver {
	return &Resolver{
		memberService:   memberService,
		familyService:   familyService,
		paymentService:  paymentService,
		cashFlowService: cashFlowService,
	}
}

// Las interfaces de los resolvers se definirán en schema.resolvers.go

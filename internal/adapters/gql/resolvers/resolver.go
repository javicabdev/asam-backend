package resolvers

import (
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
)

// Resolver contiene las dependencias necesarias para los resolvers
type Resolver struct {
	memberService   input.MemberService
	familyService   input.FamilyService
	paymentService  input.PaymentService
	cashFlowService input.CashFlowService
}

// NewResolver crea una nueva instancia del Resolver con las dependencias necesarias
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

// Funciones helper para conversión de tipos
func convertMembersToPointers(members []models.Member) []*models.Member {
	memberPtrs := make([]*models.Member, len(members))
	for i := range members {
		memberPtrs[i] = &members[i]
	}
	return memberPtrs
}

func convertPaymentsToPointers(payments []models.Payment) []*models.Payment {
	paymentPtrs := make([]*models.Payment, len(payments))
	for i := range payments {
		paymentPtrs[i] = &payments[i]
	}
	return paymentPtrs
}

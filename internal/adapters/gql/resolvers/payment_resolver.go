// payment_resolver.go
package resolvers

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"time"
)

func (r *paymentResolver) mapPaymentInputToModel(input *model.PaymentInput) *models.Payment {
	payment := &models.Payment{
		Amount:        input.Amount,
		PaymentDate:   time.Now(),
		PaymentMethod: input.PaymentMethod,
		Status:        models.PaymentStatusPaid,
	}

	if input.MemberID != nil {
		memberID := parseID(*input.MemberID)
		payment.MemberID = memberID
	}

	if input.FamilyID != nil {
		familyID := parseID(*input.FamilyID)
		payment.FamilyID = &familyID
	}

	if input.Notes != nil {
		payment.Notes = *input.Notes
	}

	return payment
}

func (r *paymentResolver) validatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.MemberID == 0 && payment.FamilyID == nil {
		return NewValidationError("either member_id or family_id must be provided")
	}

	if payment.MemberID != 0 {
		member, err := r.memberService.GetMemberByID(ctx, payment.MemberID)
		if err != nil {
			return err
		}
		if member == nil {
			return NewNotFoundError("member not found")
		}
		if member.Estado != models.EstadoActivo {
			return NewValidationError("cannot register payment for inactive member")
		}
	}

	if payment.FamilyID != nil {
		family, err := r.familyService.GetByID(ctx, *payment.FamilyID)
		if err != nil {
			return err
		}
		if family == nil {
			return NewNotFoundError("family not found")
		}
	}

	return nil
}

func (r *paymentResolver) handlePaymentMutation(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	if err := r.validatePayment(ctx, payment); err != nil {
		return nil, err
	}

	if payment.ID == 0 {
		err := r.paymentService.RegisterPayment(ctx, payment)
		if err != nil {
			return nil, err
		}
	} else {
		// Verificar que el pago no esté cancelado
		existingPayment, err := r.paymentService.GetPayment(ctx, payment.ID)
		if err != nil {
			return nil, err
		}
		if existingPayment.Status == models.PaymentStatusCancelled {
			return nil, NewValidationError("cannot update cancelled payment")
		}

		err = r.paymentService.RegisterPayment(ctx, payment)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}

package resolvers

import (
	"context"
	stdErr "errors"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
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
	if err := payment.Validate(); err != nil {
		// Chequear si es AppError
		var appErr *appErrors.AppError
		if stdErr.As(err, &appErr) {
			return appErr
		}
		// Sino, convertirlo (caso excepcional)
		return appErrors.NewValidationError(err.Error(), nil)
	}

	if payment.MemberID != 0 {
		member, err := r.memberService.GetMemberByID(ctx, payment.MemberID)
		if err != nil {
			var appErr *appErrors.AppError
			if stdErr.As(err, &appErr) {
				return appErr
			}
			return appErrors.NewValidationError(err.Error(), nil)
		}
		if member == nil {
			return appErrors.NewNotFoundError("member")
		}
		if member.Estado != models.EstadoActivo {
			return appErrors.NewValidationError(
				"cannot register payment for inactive member",
				map[string]string{
					"Member": "Inactive",
				},
			)
		}
	}

	if payment.FamilyID != nil {
		family, err := r.familyService.GetByID(ctx, *payment.FamilyID)
		if err != nil {
			return err
		}
		if family == nil {
			return appErrors.NewNotFoundError("family")
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
			return nil, appErrors.NewValidationError(
				"cannot update cancelled payment",
				map[string]string{
					"Payment": "Cancelled",
				},
			)
		}

		err = r.paymentService.RegisterPayment(ctx, payment)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}

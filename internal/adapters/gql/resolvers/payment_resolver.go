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
	// Basic model validation
	if err := payment.Validate(); err != nil {
		// Check if it's already an AppError
		var appErr *appErrors.AppError
		if stdErr.As(err, &appErr) {
			return appErr
		}
		// Otherwise, convert it to a validation error
		return appErrors.NewValidationError(err.Error(), nil)
	}

	// Check if member exists and is active
	if payment.MemberID != 0 {
		member, err := r.memberService.GetMemberByID(ctx, payment.MemberID)
		if err != nil {
			return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying member")
		}
		if member == nil {
			return appErrors.NotFound("member", nil)
		}
		if member.State != models.EstadoActivo {
			return appErrors.NewValidationError(
				"Cannot register payment for inactive member",
				map[string]string{
					"member_status": "inactive",
				},
			)
		}
	}

	// Check if family exists
	if payment.FamilyID != nil {
		family, err := r.familyService.GetByID(ctx, *payment.FamilyID)
		if err != nil {
			return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying family")
		}
		if family == nil {
			return appErrors.NotFound("family", nil)
		}
	}

	// Ensure either member or family is provided
	if payment.MemberID == 0 && payment.FamilyID == nil {
		return appErrors.NewValidationError(
			"Payment must be associated with either a member or family",
			map[string]string{
				"member_id": "required if family_id not provided",
				"family_id": "required if member_id not provided",
			},
		)
	}

	// Validate amount is positive
	if payment.Amount <= 0 {
		return appErrors.NewValidationError(
			"Payment amount must be greater than zero",
			map[string]string{
				"amount": "must be positive",
			},
		)
	}

	return nil
}

func (r *paymentResolver) handlePaymentMutation(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	// Validate the payment
	if err := r.validatePayment(ctx, payment); err != nil {
		return nil, err
	}

	if payment.ID == 0 {
		// Creating a new payment
		err := r.paymentService.RegisterPayment(ctx, payment)
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error registering payment")
		}
	} else {
		// Updating an existing payment
		// Check if payment exists and is not cancelled
		existingPayment, err := r.paymentService.GetPayment(ctx, payment.ID)
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error retrieving payment")
		}
		if existingPayment == nil {
			return nil, appErrors.NotFound("payment", nil)
		}
		if existingPayment.Status == models.PaymentStatusCancelled {
			return nil, appErrors.NewValidationError(
				"Cannot update cancelled payment",
				map[string]string{
					"status": "cancelled",
				},
			)
		}

		// Update the payment
		err = r.paymentService.RegisterPayment(ctx, payment)
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error updating payment")
		}
	}

	return payment, nil
}

// Helper methods for handling specific payment operations

func (r *paymentResolver) validatePaymentInput(input *model.PaymentInput) error {
	fields := make(map[string]string)

	// Either member_id or family_id must be provided
	if input.MemberID == nil && input.FamilyID == nil {
		fields["member_id"] = "Either member_id or family_id is required"
		fields["family_id"] = "Either member_id or family_id is required"
	}

	// Amount must be positive
	if input.Amount <= 0 {
		fields["amount"] = "Amount must be greater than zero"
	}

	// Payment method is required
	if input.PaymentMethod == "" {
		fields["payment_method"] = "Payment method is required"
	}

	if len(fields) > 0 {
		return appErrors.NewValidationError("Invalid payment input", fields)
	}

	return nil
}

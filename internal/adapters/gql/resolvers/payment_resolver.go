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
	now := time.Now()
	payment := &models.Payment{
		Amount:        input.Amount,
		PaymentDate:   &now,
		PaymentMethod: input.PaymentMethod,
		Status:        models.PaymentStatusPaid,
	}

	if input.MemberID != nil {
		memberID, err := parseID(*input.MemberID)
		if err != nil {
			return nil
		}
		payment.MemberID = &memberID
	}

	if input.FamilyID != nil {
		familyID, err := parseID(*input.FamilyID)
		if err != nil {
			return nil
		}
		payment.FamilyID = &familyID
	}

	if input.Notes != nil {
		payment.Notes = *input.Notes
	}

	return payment
}

func (r *paymentResolver) validatePayment(ctx context.Context, payment *models.Payment) error {
	// Realizar validaciones en orden de dependencia
	if err := r.validateBasicPayment(payment); err != nil {
		return err
	}

	if err := r.validatePaymentAssociations(ctx, payment); err != nil {
		return err
	}

	if err := r.validatePaymentAmount(payment); err != nil {
		return err
	}

	return nil
}

// validateBasicPayment realiza validaciones básicas del modelo
func (r *paymentResolver) validateBasicPayment(payment *models.Payment) error {
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

	// Ensure either member or family is provided
	if (payment.MemberID == nil || *payment.MemberID == 0) && (payment.FamilyID == nil || *payment.FamilyID == 0) {
		return appErrors.NewValidationError(
			"Payment must be associated with either a member or family",
			map[string]string{
				"member_id": "required if family_id not provided",
				"family_id": "required if member_id not provided",
			},
		)
	}

	return nil
}

// validatePaymentAssociations verifica que las entidades asociadas existan
func (r *paymentResolver) validatePaymentAssociations(ctx context.Context, payment *models.Payment) error {
	// Check if member exists and is active
	if payment.MemberID != nil && *payment.MemberID != 0 {
		if err := r.validateMember(ctx, *payment.MemberID); err != nil {
			return err
		}
	}

	// Check if family exists
	if payment.FamilyID != nil && *payment.FamilyID != 0 {
		if err := r.validateFamily(ctx, *payment.FamilyID); err != nil {
			return err
		}
	}

	return nil
}

// validateMember verifica que el miembro exista y esté activo
func (r *paymentResolver) validateMember(ctx context.Context, memberID uint) error {
	member, err := r.memberService.GetMemberByID(ctx, memberID)
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
	return nil
}

// validateFamily verifica que la familia exista
func (r *paymentResolver) validateFamily(ctx context.Context, familyID uint) error {
	family, err := r.familyService.GetByID(ctx, familyID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying family")
	}
	if family == nil {
		return appErrors.NotFound("family", nil)
	}
	return nil
}

// validatePaymentAmount verifica que el monto sea positivo
func (r *paymentResolver) validatePaymentAmount(payment *models.Payment) error {
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
		return r.createPayment(ctx, payment)
	}
	return r.updatePayment(ctx, payment)
}

// createPayment crea un nuevo pago
func (r *paymentResolver) createPayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	err := r.paymentService.RegisterPayment(ctx, payment)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error registering payment")
	}
	return payment, nil
}

// updatePayment actualiza un pago existente
func (r *paymentResolver) updatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
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
	return payment, nil
}

// MembershipFee resolver para el campo membership_fee en Payment
func (r *paymentResolver) MembershipFee(ctx context.Context, obj *models.Payment) (*models.MembershipFee, error) {
	// Si no hay ID de cuota, retornar nil (es opcional)
	if obj.MembershipFeeID == nil {
		return nil, nil
	}

	// Si ya está cargado por GORM (eager loading), devolverlo
	if obj.MembershipFee != nil {
		return obj.MembershipFee, nil
	}

	// Si no está cargado, devolverlo como nil
	// El frontend debería hacer eager loading si necesita este campo
	// O podríamos cargar desde el servicio, pero por simplicidad devolvemos nil
	return nil, nil
}

// Helper methods for handling specific payment operations

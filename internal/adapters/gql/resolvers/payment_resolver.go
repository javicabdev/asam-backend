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

	memberID, err := parseID(input.MemberID)
	if err != nil {
		return nil
	}

	payment := &models.Payment{
		MemberID:      memberID,
		Amount:        input.Amount,
		PaymentDate:   &now,
		PaymentMethod: input.PaymentMethod,
		Status:        models.PaymentStatusPaid,
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

	return nil
}

// validatePaymentAssociations verifica que las entidades asociadas existan
func (r *paymentResolver) validatePaymentAssociations(ctx context.Context, payment *models.Payment) error {
	// Check if member exists and is active
	if err := r.validateMember(ctx, payment.MemberID); err != nil {
		return err
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

// updatePayment actualiza un pago existente y sincroniza automáticamente el cashflow vinculado
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

	// Update the payment and sync with cashflow using transaction
	// This ensures Payment and CashFlow stay in sync
	err = r.paymentService.UpdatePaymentAndSyncCashFlow(ctx, payment)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error updating payment and syncing cashflow")
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

package resolvers

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// GetUserFromContext obtiene el usuario del contexto
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func (r *cashFlowResolver) mapTransactionInputToModel(input *model.TransactionInput) *models.CashFlow {
	transaction := &models.CashFlow{
		OperationType: input.OperationType,
		Amount:        input.Amount,
		Date:          input.Date,
		Detail:        input.Detail,
	}

	if input.MemberID != nil {
		memberID, err := parseID(*input.MemberID)
		if err != nil {
			return nil
		}
		transaction.MemberID = &memberID
	}

	if input.FamilyID != nil {
		familyID, err := parseID(*input.FamilyID)
		if err != nil {
			return nil
		}
		transaction.FamilyID = &familyID
	}

	return transaction
}

func (r *cashFlowResolver) validateTransaction(ctx context.Context, transaction *models.CashFlow) error {
	// Realizar validaciones en secuencia lógica
	if err := r.validateBasicTransaction(transaction); err != nil {
		return err
	}

	if err := r.validateAmountByOperationType(transaction); err != nil {
		return err
	}

	if err := r.validateTransactionAssociations(ctx, transaction); err != nil {
		return err
	}

	if err := r.validateTransactionDate(transaction); err != nil {
		return err
	}

	return nil
}

// validateBasicTransaction realiza las validaciones básicas del modelo
func (r *cashFlowResolver) validateBasicTransaction(transaction *models.CashFlow) error {
	if err := transaction.Validate(); err != nil {
		return appErrors.NewValidationError(
			err.Error(),
			map[string]string{
				"amount": "Must be non-zero, and positive for income or negative for expenses",
				"detail": "Description is required",
			},
		)
	}
	return nil
}

// validateAmountByOperationType verifica que el monto sea congruente con el tipo de operación
func (r *cashFlowResolver) validateAmountByOperationType(transaction *models.CashFlow) error {
	// Validate amount is appropriate for operation type
	if transaction.OperationType.IsIncome() && transaction.Amount <= 0 {
		return appErrors.NewValidationError(
			"Income amounts must be positive",
			map[string]string{"amount": "Must be positive for income operations"},
		)
	}

	if transaction.OperationType.IsExpense() && transaction.Amount >= 0 {
		return appErrors.NewValidationError(
			"Expense amounts must be negative",
			map[string]string{"amount": "Must be negative for expense operations"},
		)
	}

	return nil
}

// validateTransactionAssociations verifica que las entidades asociadas existan
func (r *cashFlowResolver) validateTransactionAssociations(ctx context.Context, transaction *models.CashFlow) error {
	// Check referenced member exists
	if transaction.MemberID != nil {
		if err := r.validateMemberExists(ctx, *transaction.MemberID); err != nil {
			return err
		}
	}

	// Check referenced family exists
	if transaction.FamilyID != nil {
		if err := r.validateFamilyExists(ctx, *transaction.FamilyID); err != nil {
			return err
		}
	}

	return nil
}

// validateMemberExists verifica que el miembro asociado exista
func (r *cashFlowResolver) validateMemberExists(ctx context.Context, memberID uint) error {
	member, err := r.memberService.GetMemberByID(ctx, memberID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying member")
	}
	if member == nil {
		return appErrors.NotFound("member", nil)
	}
	return nil
}

// validateFamilyExists verifica que la familia asociada exista
func (r *cashFlowResolver) validateFamilyExists(ctx context.Context, familyID uint) error {
	family, err := r.familyService.GetByID(ctx, familyID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying family")
	}
	if family == nil {
		return appErrors.NotFound("family", nil)
	}
	return nil
}

// validateTransactionDate verifica que la fecha no sea futura
func (r *cashFlowResolver) validateTransactionDate(transaction *models.CashFlow) error {
	if transaction.Date.After(time.Now()) {
		return appErrors.NewValidationError(
			"Transaction date cannot be in the future",
			map[string]string{"date": "Cannot be a future date"},
		)
	}
	return nil
}

func (r *cashFlowResolver) handleTransactionMutation(ctx context.Context,
	transaction *models.CashFlow) (*models.CashFlow, error) {
	// Validate transaction data
	if err := r.validateTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	// Check for zero amount (common error case)
	if transaction.Amount == 0 {
		return nil, appErrors.NewValidationError(
			"Transaction amount cannot be zero",
			map[string]string{"amount": "Must be non-zero"},
		)
	}

	// Create or update transaction based on ID
	if transaction.ID == 0 {
		return r.createTransaction(ctx, transaction)
	}
	return r.updateTransaction(ctx, transaction)
}

// createTransaction crea una nueva transacción
func (r *cashFlowResolver) createTransaction(ctx context.Context, transaction *models.CashFlow) (*models.CashFlow, error) {
	err := r.cashFlowService.RegisterMovement(ctx, transaction)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error processing transaction")
	}
	return transaction, nil
}

// updateTransaction actualiza una transacción existente
func (r *cashFlowResolver) updateTransaction(ctx context.Context, transaction *models.CashFlow) (*models.CashFlow, error) {
	err := r.cashFlowService.UpdateMovement(ctx, transaction)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error processing transaction")
	}
	return transaction, nil
}

func (r *cashFlowResolver) handleBalanceAdjustment(ctx context.Context, amount float64,
	reason string) (*model.MutationResponse, error) {
	// Verificar permisos y validar datos
	if err := r.validateBalanceAdjustment(ctx, amount, reason); err != nil {
		return nil, err
	}

	// Crear y registrar el ajuste
	adjustment := r.createAdjustmentTransaction(amount, reason)
	return r.processAdjustment(ctx, adjustment)
}

// validateBalanceAdjustment verifica permisos y valida datos del ajuste
func (r *cashFlowResolver) validateBalanceAdjustment(ctx context.Context, amount float64, reason string) error {
	// Verificar que el usuario es administrador
	user := GetUserFromContext(ctx)
	if user == nil || !user.IsAdmin() {
		return appErrors.NewBusinessError(appErrors.ErrForbidden, "Insufficient permissions to adjust balance")
	}

	// Validate amount is non-zero
	if amount == 0 {
		return appErrors.Business(
			appErrors.ErrInvalidAmount,
			"Adjustment amount cannot be zero",
			nil,
		)
	}

	// Validate reason is provided
	if reason == "" {
		return appErrors.NewValidationError(
			"Adjustment reason is required",
			map[string]string{"reason": "Required for audit purposes"},
		)
	}

	return nil
}

// createAdjustmentTransaction crea la transacción de ajuste
func (r *cashFlowResolver) createAdjustmentTransaction(amount float64, reason string) *models.CashFlow {
	adjustment := &models.CashFlow{
		OperationType: models.OperationTypeOtherIncome,
		Amount:        amount,
		Date:          time.Now(),
		Detail:        reason,
	}

	// Set proper operation type based on amount
	if amount < 0 {
		adjustment.OperationType = models.OperationTypeCurrentExpense
	}

	return adjustment
}

// processAdjustment registra el ajuste y prepara la respuesta
func (r *cashFlowResolver) processAdjustment(ctx context.Context, adjustment *models.CashFlow) (*model.MutationResponse, error) {
	err := r.cashFlowService.RegisterMovement(ctx, adjustment)
	if err != nil {
		// Use structured error response even for error cases
		errorMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errorMsg,
		}, nil
	}

	// Success response
	message := "Balance adjusted successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &message,
	}, nil
}

package resolvers

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"time"
)

func (r *cashFlowResolver) mapTransactionInputToModel(input *model.TransactionInput) *models.CashFlow {
	transaction := &models.CashFlow{
		OperationType: input.OperationType,
		Amount:        input.Amount,
		Date:          input.Date,
		Detail:        input.Detail,
	}

	if input.MemberID != nil {
		memberID := parseID(*input.MemberID)
		transaction.MemberID = &memberID
	}

	if input.FamilyID != nil {
		familyID := parseID(*input.FamilyID)
		transaction.FamilyID = &familyID
	}

	return transaction
}

func (r *cashFlowResolver) validateTransaction(ctx context.Context, transaction *models.CashFlow) error {
	// Basic model validation
	if err := transaction.Validate(); err != nil {
		return appErrors.NewValidationError(
			err.Error(),
			map[string]string{
				"amount": "Must be non-zero, and positive for income or negative for expenses",
				"detail": "Description is required",
			},
		)
	}

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

	// Check referenced member exists
	if transaction.MemberID != nil {
		member, err := r.memberService.GetMemberByID(ctx, *transaction.MemberID)
		if err != nil {
			return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying member")
		}
		if member == nil {
			return appErrors.NotFound("member", nil)
		}
	}

	// Check referenced family exists
	if transaction.FamilyID != nil {
		family, err := r.familyService.GetByID(ctx, *transaction.FamilyID)
		if err != nil {
			return appErrors.Wrap(err, appErrors.ErrDatabaseError, "Error verifying family")
		}
		if family == nil {
			return appErrors.NotFound("family", nil)
		}
	}

	// Validate date is not in the future
	if transaction.Date.After(time.Now()) {
		return appErrors.NewValidationError(
			"Transaction date cannot be in the future",
			map[string]string{"date": "Cannot be a future date"},
		)
	}

	return nil
}

func (r *cashFlowResolver) handleTransactionMutation(ctx context.Context, transaction *models.CashFlow) (*models.CashFlow, error) {
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

	// Create or update transaction
	var err error
	if transaction.ID == 0 {
		err = r.cashFlowService.RegisterMovement(ctx, transaction)
	} else {
		err = r.cashFlowService.UpdateMovement(ctx, transaction)
	}

	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrInternalError, "Error processing transaction")
	}

	return transaction, nil
}

func (r *cashFlowResolver) handleBalanceAdjustment(ctx context.Context, amount float64, reason string) (*model.MutationResponse, error) {
	// Validate amount is non-zero
	if amount == 0 {
		return nil, appErrors.NewValidationError(
			"Adjustment amount cannot be zero",
			map[string]string{"amount": "Must be non-zero"},
		)
	}

	// Validate reason is provided
	if reason == "" {
		return nil, appErrors.NewValidationError(
			"Adjustment reason is required",
			map[string]string{"reason": "Required for audit purposes"},
		)
	}

	// Create adjustment transaction
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

	// Register the adjustment
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

// Helper method for validating transaction inputs from GraphQL
func (r *cashFlowResolver) validateTransactionInput(input *model.TransactionInput) error {
	fields := make(map[string]string)

	// Detail is required
	if input.Detail == "" {
		fields["detail"] = "Description is required"
	}

	// Amount must not be zero
	if input.Amount == 0 {
		fields["amount"] = "Amount must be non-zero"
	}

	// Operation type must be valid
	if !input.OperationType.IsValid() {
		fields["operation_type"] = "Invalid operation type"
	} else {
		// Amount must match operation type
		if input.OperationType.IsIncome() && input.Amount < 0 {
			fields["amount"] = "Income amount must be positive"
		}
		if input.OperationType.IsExpense() && input.Amount > 0 {
			fields["amount"] = "Expense amount must be negative"
		}
	}

	// Return validation error if any fields failed
	if len(fields) > 0 {
		return appErrors.NewValidationError("Invalid transaction input", fields)
	}

	return nil
}

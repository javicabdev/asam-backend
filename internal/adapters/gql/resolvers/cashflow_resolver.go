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
	if err := transaction.Validate(); err != nil {
		return appErrors.NewValidationError(
			"El monto es inválido",
			map[string]string{"amount": "Debe ser distinto de 0, y positivo o negativo según el tipo de operación"},
		)
	}

	if transaction.MemberID != nil {
		member, err := r.memberService.GetMemberByID(ctx, *transaction.MemberID)
		if err != nil {
			return err
		}
		if member == nil {
			return appErrors.NewNotFoundError("member")
		}
	}

	if transaction.FamilyID != nil {
		family, err := r.familyService.GetByID(ctx, *transaction.FamilyID)
		if err != nil {
			return err
		}
		if family == nil {
			return appErrors.NewNotFoundError("family")
		}
	}

	if appErr := validateAmount(transaction.Amount, transaction.OperationType); appErr != nil {
		return appErr
	}

	return nil
}

func (r *cashFlowResolver) handleTransactionMutation(ctx context.Context, transaction *models.CashFlow) (*models.CashFlow, error) {
	if err := r.validateTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	var err error
	if transaction.ID == 0 {
		err = r.cashFlowService.RegisterMovement(ctx, transaction)
	} else {
		err = r.cashFlowService.UpdateMovement(ctx, transaction)
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *cashFlowResolver) handleBalanceAdjustment(ctx context.Context, amount float64, reason string) (*model.MutationResponse, error) {
	adjustment := &models.CashFlow{
		OperationType: models.OperationTypeOtherIncome,
		Amount:        amount,
		Date:          time.Now(),
		Detail:        reason,
	}

	if amount < 0 {
		adjustment.OperationType = models.OperationTypeCurrentExpense
	}

	if err := r.cashFlowService.RegisterMovement(ctx, adjustment); err != nil {
		errStr := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errStr,
		}, nil
	}

	message := "Balance adjusted successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &message,
	}, nil
}

package resolvers

import (
	"errors"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"gorm.io/gorm"
	"strconv"
)

func parseID(id string) uint {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

func validateAmount(amount float64, operationType models.OperationType) error {
	if amount == 0 {
		return fmt.Errorf("amount cannot be zero")
	}

	if operationType.IsIncome() && amount < 0 {
		return fmt.Errorf("income amount must be positive")
	}

	if operationType.IsExpense() && amount > 0 {
		return fmt.Errorf("expense amount must be negative")
	}

	return nil
}

func wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%s: record not found", operation)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func stringPtr(s string) *string {
	return &s
}

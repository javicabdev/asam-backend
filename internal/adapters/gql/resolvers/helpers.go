package resolvers

import (
	"errors"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"gorm.io/gorm"
	"strconv"
)

// parseID convierte un ID de string a uint
func parseID(id string) uint {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

// validateAmount valida que un monto sea adecuado según el tipo de operación
func validateAmount(amount float64, operationType models.OperationType) error {
	if amount == 0 {
		return appErrors.Validation("Amount cannot be zero", "amount", "cannot be zero")
	}
	if operationType.IsIncome() && amount < 0 {
		return appErrors.Validation("Income amount must be positive", "amount", "must be positive for income")
	}
	if operationType.IsExpense() && amount > 0 {
		return appErrors.Validation("Expense amount must be negative", "amount", "must be negative for expense")
	}
	return nil
}

// handleError provides consistent error handling for GraphQL resolvers
func handleError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Already an AppError, just return it
	if appErr, ok := appErrors.AsAppError(err); ok {
		return appErr
	}

	// Check for common GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.NotFound(operation, err)
	}

	// For other errors, create a generic internal error
	return appErrors.New(appErrors.ErrInternalError,
		"An error occurred during "+operation)
}

// handleNotFound helps with consistent not found errors
func handleNotFound(entity string, id string) error {
	return appErrors.NotFound(entity,
		errors.New("no "+entity+" found with id "+id))
}

// ensureExists checks if an entity exists and returns appropriate error if not
func ensureExists(entity interface{}, entityName string, id string) error {
	if entity == nil {
		return handleNotFound(entityName, id)
	}
	return nil
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr creates a pointer to an int
func intPtr(i int) *int {
	return &i
}

// boolPtr creates a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

// mapDatabaseError maps database errors to user-friendly messages
func mapDatabaseError(err error, operation string) error {
	// Check for specific database errors to return more specific error types
	errorMsg := err.Error()

	// Duplicate key error
	if len(errorMsg) >= 19 && errorMsg[:19] == "duplicate key value" {
		return appErrors.New(appErrors.ErrDuplicateEntry,
			"A record with the same key already exists")
	}

	// Foreign key violation
	if len(errorMsg) >= 27 && errorMsg[:27] == "foreign key violation" {
		return appErrors.New(appErrors.ErrInvalidOperation,
			"Cannot perform operation due to related records")
	}

	// Default to generic database error
	return appErrors.DB(err, "Database error during "+operation)
}

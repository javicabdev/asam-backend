package gql

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

// CustomErrorPresenter transforms errors into GraphQL errors
func CustomErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// If already a GraphQL error, return it as is
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		return gqlErr
	}

	// Get the current query/mutation path
	path := graphql.GetPath(ctx)

	// Get operation name for better error context
	operation := "unknown"
	if op := graphql.GetOperationContext(ctx); op != nil {
		operation = op.OperationName
	}

	// Handle specific error types
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return &gqlerror.Error{
			Path:    path,
			Message: "Resource not found",
			Extensions: map[string]interface{}{
				"code":      appErrors.ErrNotFound,
				"operation": operation,
			},
		}

	case errors.Is(err, context.DeadlineExceeded):
		return &gqlerror.Error{
			Path:    path,
			Message: "Operation timed out",
			Extensions: map[string]interface{}{
				"code":      appErrors.ErrInternalError,
				"operation": operation,
			},
		}

	case errors.Is(err, context.Canceled):
		return &gqlerror.Error{
			Path:    path,
			Message: "Operation was canceled",
			Extensions: map[string]interface{}{
				"code":      appErrors.ErrInternalError,
				"operation": operation,
			},
		}
	}

	// Map AppError
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		// Build extensions with code and error fields
		extensions := map[string]interface{}{
			"code":      appErr.Code,
			"operation": operation,
		}

		// Add validation fields if they exist
		if appErr.Fields != nil && len(appErr.Fields) > 0 {
			extensions["fields"] = appErr.Fields
		}

		return &gqlerror.Error{
			Path:       path,
			Message:    appErr.Message,
			Extensions: extensions,
		}
	}

	// For unstructured errors, create a generic internal error
	return &gqlerror.Error{
		Path:    path,
		Message: fmt.Sprintf("Internal error: %s", err.Error()),
		Extensions: map[string]interface{}{
			"code":      appErrors.ErrInternalError,
			"operation": operation,
		},
	}
}

// ErrorHandler returns a middleware.ConfigErrorPresenter compatible function
// to integrate with our refactored system
func ErrorHandler() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		// Look for errorHandler in context
		if handler, ok := ctx.Value(middleware.ErrorHandlerKey{}).(*middleware.ErrorMiddleware); ok {
			return handler.HandleError(ctx, err)
		}

		// Fallback to CustomErrorPresenter if no errorHandler in context
		return CustomErrorPresenter(ctx, err)
	}
}

// RecoverFunc provides a function to recover from panics in GraphQL resolvers
func RecoverFunc() graphql.RecoverFunc {
	return func(ctx context.Context, panicValue interface{}) error {
		// Build error message
		errMsg := "Internal server error"
		if panicValue != nil {
			errMsg = fmt.Sprintf("GraphQL panic: %v", panicValue)
		}

		// Create a structured error
		appErr := appErrors.New(appErrors.ErrInternalError, errMsg)

		// Look for errorHandler in context for consistent handling
		if handler, ok := ctx.Value(middleware.ErrorHandlerKey{}).(*middleware.ErrorMiddleware); ok {
			return handler.HandleError(ctx, appErr)
		}

		// Fallback to direct conversion
		return CustomErrorPresenter(ctx, appErr)
	}
}

// MapErrorToGraphQL is a utility function to convert any error to a GraphQL error
// with consistent formatting
func MapErrorToGraphQL(ctx context.Context, err error, defaultMessage string) *gqlerror.Error {
	if err == nil {
		return nil
	}

	// Look for errorHandler in context for consistent handling
	if handler, ok := ctx.Value(middleware.ErrorHandlerKey{}).(*middleware.ErrorMiddleware); ok {
		return handler.HandleError(ctx, err)
	}

	// If no specific message provided, use a default one
	if defaultMessage == "" {
		defaultMessage = "An error occurred"
	}

	// If it's already an AppError, preserve its information
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		// Use original error's message
		return CustomErrorPresenter(ctx, err)
	}

	// Create a new AppError with the default message
	wrappedErr := appErrors.Wrap(err, appErrors.ErrInternalError, defaultMessage)
	return CustomErrorPresenter(ctx, wrappedErr)
}

package resolvers

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// User Management Queries

// GetUser retrieves a user by ID (Admin only)
func (r *Resolver) GetUser(ctx context.Context, id string) (*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	user, err := r.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ListUsers retrieves a paginated list of users (Admin only)
func (r *Resolver) ListUsers(ctx context.Context, page *int, pageSize *int) ([]*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	// Default pagination
	pageNum := 1
	size := 10

	if page != nil && *page > 0 {
		pageNum = *page
	}
	if pageSize != nil && *pageSize > 0 && *pageSize <= 100 {
		size = *pageSize
	}

	users, err := r.userService.ListUsers(ctx, pageNum, size)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetCurrentUser retrieves the currently authenticated user
func (r *Resolver) GetCurrentUser(ctx context.Context) (*models.User, error) {
	// Get user from context
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, errors.NewUnauthorizedError()
	}

	// Get fresh user data to ensure it's up to date
	freshUser, err := r.userService.GetUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return freshUser, nil
}

// User Management Mutations

// CreateUser creates a new user (Admin only)
func (r *Resolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	// Convert role from GraphQL enum to domain model
	role := convertGraphQLRoleToDomain(input.Role)

	user, err := r.userService.CreateUser(ctx, input.Username, input.Password, role)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user (Admin only)
func (r *Resolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(input.ID)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Password != nil {
		updates["password"] = *input.Password
	}
	if input.Role != nil {
		updates["role"] = convertGraphQLRoleToDomain(*input.Role)
	}
	if input.IsActive != nil {
		updates["isActive"] = *input.IsActive
	}

	user, err := r.userService.UpdateUser(ctx, userID, updates)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user (Admin only)
func (r *Resolver) DeleteUser(ctx context.Context, id string) (*model.MutationResponse, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	// Prevent self-deletion
	currentUser, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if ok && currentUser != nil && currentUser.ID == userID {
		errMsg := "Cannot delete your own user account"
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	err = r.userService.DeleteUser(ctx, userID)
	if err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	successMsg := "User deleted successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// ChangePassword changes the current user's password
func (r *Resolver) ChangePassword(ctx context.Context, input model.ChangePasswordInput) (*model.MutationResponse, error) {
	// Get current user
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, errors.NewUnauthorizedError()
	}

	err := r.userService.ChangePassword(ctx, user.ID, input.CurrentPassword, input.NewPassword)
	if err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	successMsg := "Password changed successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// ResetUserPassword resets a user's password (Admin only)
func (r *Resolver) ResetUserPassword(ctx context.Context, userID string, newPassword string) (*model.MutationResponse, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := parseID(userID)
	if err != nil {
		return nil, err
	}

	err = r.userService.ResetPassword(ctx, id, newPassword)
	if err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	successMsg := "Password reset successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

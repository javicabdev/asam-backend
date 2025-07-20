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
func (r *userResolver) GetUser(ctx context.Context, id string) (*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	return r.userService.GetUser(ctx, userID)
}

// ListUsers retrieves a paginated list of users (Admin only)
func (r *userResolver) ListUsers(ctx context.Context, page *int, pageSize *int) ([]*models.User, error) {
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

	return r.userService.ListUsers(ctx, pageNum, size)
}

// GetCurrentUser retrieves the currently authenticated user
func (r *userResolver) GetCurrentUser(ctx context.Context) (*models.User, error) {
	// Get user from context
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, errors.NewUnauthorizedError()
	}

	// Get fresh user data to ensure it's up to date
	return r.userService.GetUser(ctx, user.ID)
}

// User Management Mutations

// CreateUser creates a new user (Admin only)
func (r *userResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*models.User, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	// Convert role from GraphQL enum to domain model
	role := convertGraphQLRoleToDomain(input.Role)

	// Parse memberID if provided
	var memberID *uint
	if input.MemberID != nil {
		id, err := parseID(*input.MemberID)
		if err != nil {
			return nil, err
		}
		memberID = &id
	}

	return r.userService.CreateUser(ctx, input.Username, input.Email, input.Password, role, memberID)
}

// UpdateUser updates an existing user (Admin only)
func (r *userResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*models.User, error) {
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
	if input.Email != nil {
		updates["email"] = *input.Email
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
	if input.MemberID != nil {
		// Parse memberID
		id, err := parseID(*input.MemberID)
		if err != nil {
			return nil, err
		}
		updates["memberID"] = &id
	}

	return r.userService.UpdateUser(ctx, userID, updates)
}

// DeleteUser deletes a user (Admin only)
func (r *userResolver) DeleteUser(ctx context.Context, id string) (*model.MutationResponse, error) {
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

	if err := r.userService.DeleteUser(ctx, userID); err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	successMsg := "User deleted successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// ChangePassword changes the current user's password
func (r *userResolver) ChangePassword(ctx context.Context, input model.ChangePasswordInput) (*model.MutationResponse, error) {
	// Get current user
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		return nil, errors.NewUnauthorizedError()
	}

	if err := r.userService.ChangePassword(ctx, user.ID, input.CurrentPassword, input.NewPassword); err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	successMsg := "Password changed successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// ResetUserPassword resets a user's password (Admin only)
func (r *userResolver) ResetUserPassword(ctx context.Context, userID string, newPassword string) (*model.MutationResponse, error) {
	// Check admin permission
	if err := middleware.MustBeAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := parseID(userID)
	if err != nil {
		return nil, err
	}

	if err := r.userService.ResetPassword(ctx, id, newPassword); err != nil {
		errMsg := err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	successMsg := "Password reset successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// Helper functions

// convertGraphQLRoleToDomain converts GraphQL UserRole enum to domain Role
func convertGraphQLRoleToDomain(role model.UserRole) models.Role {
	switch role {
	case model.UserRoleAdmin:
		return models.RoleAdmin
	case model.UserRoleUser:
		return models.RoleUser
	default:
		return models.RoleUser
	}
}

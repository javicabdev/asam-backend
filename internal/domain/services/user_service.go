package services

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// userService implements user management operations
type userService struct {
	userRepo output.UserRepository
	logger   logger.Logger
}

// NewUserService creates a new user management service
func NewUserService(userRepo output.UserRepository, logger logger.Logger) input.UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// CreateUser creates a new user with the given details
func (s *userService) CreateUser(ctx context.Context, username, password string, role models.Role) (*models.User, error) {
	// Validate inputs
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}

	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Check if username already exists
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, errors.DB(err, "error checking existing username")
	}
	if existingUser != nil {
		return nil, errors.NewValidationError(
			"Username already exists",
			map[string]string{"username": "This username is already taken"},
		)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Create user
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.DB(err, "error creating user")
	}

	s.logger.Info("User created successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("role", string(user.Role)),
	)

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// UpdateUser updates an existing user's details
func (s *userService) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Apply updates
	if username, ok := updates["username"].(string); ok && username != "" {
		if err := s.validateUsername(username); err != nil {
			return nil, err
		}

		// Check if new username is taken by another user
		if username != user.Username {
			existingUser, err := s.userRepo.FindByUsername(ctx, username)
			if err != nil && !errors.IsNotFoundError(err) {
				return nil, errors.DB(err, "error checking username availability")
			}
			if existingUser != nil && existingUser.ID != user.ID {
				return nil, errors.NewValidationError(
					"Username already exists",
					map[string]string{"username": "This username is already taken"},
				)
			}
		}
		user.Username = username
	}

	if password, ok := updates["password"].(string); ok && password != "" {
		if err := s.validatePassword(password); err != nil {
			return nil, err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrInternalError, "error hashing password")
		}
		user.Password = string(hashedPassword)
	}

	if role, ok := updates["role"].(models.Role); ok {
		user.Role = role
	}

	if isActive, ok := updates["isActive"].(bool); ok {
		user.IsActive = isActive
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.DB(err, "error updating user")
	}

	s.logger.Info("User updated successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Prevent deleting the last admin
	if user.Role == models.RoleAdmin {
		// Count remaining admins (this would need a new repository method)
		// For now, we'll just log a warning
		s.logger.Warn("Deleting an admin user",
			zap.Uint("user_id", user.ID),
			zap.String("username", user.Username),
		)
	}

	// Soft delete by deactivating the user
	user.IsActive = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error deleting user")
	}

	s.logger.Info("User deleted successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	// For now, we'll return all users and handle pagination in memory
	// In a real implementation, this would be done at the repository level

	// This is a simplified implementation - you would need to add a ListUsers method to UserRepository
	// for proper pagination support

	// Placeholder implementation
	users := make([]*models.User, 0)

	// Clear passwords before returning
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return errors.NewBusinessError(errors.ErrUnauthorized, "current password is incorrect")
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error updating password")
	}

	s.logger.Info("Password changed successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// ResetPassword resets a user's password (admin function)
func (s *userService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error resetting password")
	}

	s.logger.Info("Password reset successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// validateUsername validates username requirements
func (s *userService) validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.NewValidationError(
			"Username is required",
			map[string]string{"username": "Username cannot be empty"},
		)
	}

	if len(username) < 3 {
		return errors.NewValidationError(
			"Username too short",
			map[string]string{"username": "Username must be at least 3 characters long"},
		)
	}

	if len(username) > 50 {
		return errors.NewValidationError(
			"Username too long",
			map[string]string{"username": "Username must not exceed 50 characters"},
		)
	}

	// Check for valid characters (alphanumeric, underscore, hyphen, dot)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-' || char == '.') {
			return errors.NewValidationError(
				"Invalid username format",
				map[string]string{"username": "Username can only contain letters, numbers, underscore, hyphen, and dot"},
			)
		}
	}

	return nil
}

// validatePassword validates password requirements
func (s *userService) validatePassword(password string) error {
	if password == "" {
		return errors.NewValidationError(
			"Password is required",
			map[string]string{"password": "Password cannot be empty"},
		)
	}

	if len(password) < 8 {
		return errors.NewValidationError(
			"Password too short",
			map[string]string{"password": "Password must be at least 8 characters long"},
		)
	}

	if len(password) > 100 {
		return errors.NewValidationError(
			"Password too long",
			map[string]string{"password": "Password must not exceed 100 characters"},
		)
	}

	// Check password complexity
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return errors.NewValidationError(
			"Password does not meet complexity requirements",
			map[string]string{
				"password": "Password must contain at least one uppercase letter, one lowercase letter, and one number",
			},
		)
	}

	// hasSpecial is optional but recommended
	if !hasSpecial {
		s.logger.Debug("Password does not contain special characters",
			zap.String("recommendation", "Consider adding special characters for better security"),
		)
	}

	return nil
}

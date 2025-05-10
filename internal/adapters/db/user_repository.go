package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository crea una nueva instancia del repositorio de usuarios
// que implementa la interfaz output.UserRepository.
func NewUserRepository(db *gorm.DB) output.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// Check for specific database errors
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "User with that username already exists")
		}
		return appErrors.DB(result.Error, "Error creating user")
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "Error finding user by ID")
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "Error finding user by username")
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		// Check for specific errors
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("user", nil)
		}
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "Username already taken")
		}
		if IsConstraintViolationError(result.Error) {
			return appErrors.New(appErrors.ErrInvalidOperation, "Cannot update user due to constraint violations")
		}
		return appErrors.DB(result.Error, "Error updating user")
	}

	// Check if any record was actually updated
	if result.RowsAffected == 0 {
		return appErrors.NotFound("user", nil)
	}

	return nil
}

// Additional helper methods to improve consistency

// IsUserActive checks if a user is active
func (r *userRepository) IsUserActive(ctx context.Context, userID uint) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_active = ?", userID, true).
		Count(&count)

	if result.Error != nil {
		return false, appErrors.DB(result.Error, "Error checking user active status")
	}

	return count > 0, nil
}

// DeactivateUser sets a user as inactive
func (r *userRepository) DeactivateUser(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deactivating user")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("user", nil)
	}

	return nil
}

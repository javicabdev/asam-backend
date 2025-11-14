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

// Create implements the output.UserRepository interface
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

// CreateUser is a method that doesn't require a context parameter.
// This method is used by test fixtures.
func (r *userRepository) CreateUser(user *models.User) error {
	return r.Create(context.Background(), user)
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Preload("Member").First(&user, id)
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
	result := r.db.WithContext(ctx).Preload("Member").Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "Error finding user by username")
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Preload("Member").Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "Error finding user by email")
	}
	return &user, nil
}

func (r *userRepository) FindByMemberID(ctx context.Context, memberID uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Preload("Member").Where("member_id = ?", memberID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "Error finding user by member ID")
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

// ListUsers returns a paginated list of users with optional filters
func (r *userRepository) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Calculate offset
	offset := (page - 1) * pageSize

	// Count total records
	countResult := r.db.WithContext(ctx).Model(&models.User{}).Count(&total)
	if countResult.Error != nil {
		return nil, 0, appErrors.DB(countResult.Error, "Error counting users")
	}

	// Fetch paginated results with Member preloaded
	result := r.db.WithContext(ctx).
		Preload("Member").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		return nil, 0, appErrors.DB(result.Error, "Error listing users")
	}

	return users, total, nil
}

// GetUserWithMember retrieves a user with their associated member data
func (r *userRepository) GetUserWithMember(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).
		Preload("Member").
		First(&user, userID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "Error finding user with member")
	}

	return &user, nil
}

// CountUsersByRole counts users by their role
func (r *userRepository) CountUsersByRole(ctx context.Context, role models.Role) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ? AND is_active = ?", role, true).
		Count(&count)

	if result.Error != nil {
		return 0, appErrors.DB(result.Error, "Error counting users by role")
	}

	return count, nil
}

// Delete permanently deletes a user from the database
// This will fail if the user has a member_id set due to OnDelete:RESTRICT constraint
func (r *userRepository) Delete(ctx context.Context, userID uint) error {
	// Start a transaction to ensure atomicity
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, delete all refresh tokens associated with this user
		if err := tx.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
			return appErrors.DB(err, "Error deleting user refresh tokens")
		}

		// Delete all verification tokens associated with this user
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.VerificationToken{}).Error; err != nil {
			return appErrors.DB(err, "Error deleting user verification tokens")
		}

		// Now delete the user
		// This will fail if user has member_id set (OnDelete:RESTRICT constraint)
		result := tx.Unscoped().Delete(&models.User{}, userID)
		if result.Error != nil {
			if IsConstraintViolationError(result.Error) {
				return appErrors.New(appErrors.ErrInvalidOperation, "Cannot delete user with associated member. Please remove member association first")
			}
			return appErrors.DB(result.Error, "Error deleting user")
		}

		if result.RowsAffected == 0 {
			return appErrors.NotFound("user", nil)
		}

		return nil
	})
}

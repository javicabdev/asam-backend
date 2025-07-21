package services

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// validateCreateUserInput validates all inputs for user creation
func (s *userService) validateCreateUserInput(ctx context.Context, username, password string, role models.Role, memberID *uint) error {
	// Validate username
	if err := s.validateUsername(username); err != nil {
		return err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return err
	}

	// Check if username already exists
	if err := s.checkUsernameAvailabilityForNew(ctx, username); err != nil {
		return err
	}

	// Validate member association based on role
	if err := s.validateMemberAssociationForRole(ctx, role, memberID); err != nil {
		return err
	}

	return nil
}

// checkUsernameAvailabilityForNew checks if username is available for a new user
func (s *userService) checkUsernameAvailabilityForNew(ctx context.Context, username string) error {
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.DB(err, "error checking existing username")
	}
	if existingUser != nil {
		return errors.NewValidationError(
			"Username already exists",
			map[string]string{"username": "This username is already taken"},
		)
	}
	return nil
}

// validateMemberAssociationForRole validates member association based on role
func (s *userService) validateMemberAssociationForRole(ctx context.Context, role models.Role, memberID *uint) error {
	if role == models.RoleUser {
		if memberID == nil {
			return errors.NewValidationError(
				"Usuario con rol USER requiere un socio asociado",
				map[string]string{"memberID": "Campo requerido para usuarios no administradores"},
			)
		}

		// Verify member exists
		member, err := s.memberRepo.GetByID(ctx, *memberID)
		if err != nil {
			return errors.DB(err, "error verificando socio")
		}
		if member == nil {
			return errors.NewValidationError(
				"Socio no encontrado",
				map[string]string{"memberID": "El socio especificado no existe"},
			)
		}

		// Verify member doesn't already have a user
		existingUser, err := s.userRepo.FindByMemberID(ctx, *memberID)
		if err != nil && !errors.IsNotFoundError(err) {
			return errors.DB(err, "error verificando usuario existente")
		}
		if existingUser != nil {
			return errors.NewValidationError(
				"El socio ya tiene un usuario asociado",
				map[string]string{"memberID": "Cada socio solo puede tener un usuario"},
			)
		}
	} else if role == models.RoleAdmin && memberID != nil {
		return errors.NewValidationError(
			"Usuario administrador no puede tener socio asociado",
			map[string]string{"memberID": "Los administradores no deben estar asociados a un socio"},
		)
	}

	return nil
}

// createUserWithValidatedData creates a user after all validations have passed
func (s *userService) createUserWithValidatedData(ctx context.Context, username, email, password string, role models.Role, memberID *uint) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Create user
	user := &models.User{
		Username:      username,
		Email:         email,
		Password:      string(hashedPassword),
		Role:          role,
		MemberID:      memberID,
		IsActive:      true,
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.DB(err, "error creating user")
	}

	s.logger.Info("User created successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("role", string(user.Role)),
	)

	// Send verification email if username is an email
	if strings.Contains(username, "@") && !user.EmailVerified {
		if err := s.SendVerificationEmail(ctx, user.ID); err != nil {
			s.logger.Error("Failed to send verification email",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
		}
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// determineFinalRoleAndMember determines the final role and memberID values
func (s *userService) determineFinalRoleAndMember(user *models.User, updates map[string]any) (finalRole models.Role, finalMemberID *uint, hasRole bool, hasMemberID bool) {
	newRole, hasRole := updates["role"].(models.Role)
	newMemberID, hasMemberID := updates["memberID"].(*uint)

	// Determine the final role
	finalRole = user.Role
	if hasRole {
		finalRole = newRole
	}

	// Determine the final memberID
	finalMemberID = user.MemberID
	if hasMemberID {
		finalMemberID = newMemberID
	}

	return finalRole, finalMemberID, hasRole, hasMemberID
}

// validateRoleMemberCombination validates the combination of role and member
func (s *userService) validateRoleMemberCombination(ctx context.Context, user *models.User, finalRole models.Role, finalMemberID *uint) error {
	if finalRole == models.RoleUser {
		if finalMemberID == nil {
			return errors.NewValidationError(
				"Usuario con rol USER requiere un socio asociado",
				map[string]string{"memberID": "Campo requerido para usuarios no administradores"},
			)
		}

		// Verify member exists
		member, err := s.memberRepo.GetByID(ctx, *finalMemberID)
		if err != nil {
			return errors.DB(err, "error verificando socio")
		}
		if member == nil {
			return errors.NewValidationError(
				"Socio no encontrado",
				map[string]string{"memberID": "El socio especificado no existe"},
			)
		}

		// If memberID is changing, verify new member doesn't already have a user
		if user.MemberID == nil || *user.MemberID != *finalMemberID {
			existingUser, err := s.userRepo.FindByMemberID(ctx, *finalMemberID)
			if err != nil && !errors.IsNotFoundError(err) {
				return errors.DB(err, "error verificando usuario existente")
			}
			if existingUser != nil && existingUser.ID != user.ID {
				return errors.NewValidationError(
					"El socio ya tiene un usuario asociado",
					map[string]string{"memberID": "Cada socio solo puede tener un usuario"},
				)
			}
		}
	} else if finalRole == models.RoleAdmin && finalMemberID != nil {
		return errors.NewValidationError(
			"Usuario administrador no puede tener socio asociado",
			map[string]string{"memberID": "Los administradores no deben estar asociados a un socio"},
		)
	}

	return nil
}

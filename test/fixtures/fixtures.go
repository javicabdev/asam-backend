package fixtures

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// CreateTestUser creates a test user with the given username and password
// and returns the created user object.
func CreateTestUser(t *testing.T, userRepo interface{}, username, password string) *models.User {
	// Create a new user with the provided username and password
	user := &models.User{
		Username: username,
		Role:     models.RoleUser,
		IsActive: true,
	}

	// Set the password (this will hash it)
	err := user.SetPassword(password)
	require.NoError(t, err)

	// Use the userRepo to create the user in the database
	// Try to use the context-aware Create method first
	if repo, ok := userRepo.(interface {
		Create(context.Context, *models.User) error
	}); ok {
		err = repo.Create(context.Background(), user)
	} else if repo, ok := userRepo.(interface{ CreateUser(*models.User) error }); ok {
		// Try the CreateUser method
		err = repo.CreateUser(user)
	} else if repo, ok := userRepo.(interface{ Create(*models.User) error }); ok {
		// Fall back to the non-context Create method
		err = repo.Create(user)
	} else {
		// No suitable method found
		t.Fatalf("userRepo does not implement any known Create method")
	}
	require.NoError(t, err)

	return user
}

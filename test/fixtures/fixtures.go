package fixtures

import (
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
	err = userRepo.(interface{ Create(*models.User) error }).Create(user)
	require.NoError(t, err)

	return user
}

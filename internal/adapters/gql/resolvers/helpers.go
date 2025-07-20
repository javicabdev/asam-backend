package resolvers

import (
	"strconv"
	"strings"
)

// parseID convierte un ID de string a uint
func parseID(id string) (uint, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// containsIgnoreCase verifica si s contiene substr sin importar mayúsculas/minúsculas
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

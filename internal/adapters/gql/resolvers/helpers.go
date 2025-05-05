package resolvers

import (
	"strconv"
)

// parseID convierte un ID de string a uint
func parseID(id string) uint {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

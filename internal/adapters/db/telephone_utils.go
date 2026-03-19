package db

import (
	"strings"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// deduplicateTelephones elimina teléfonos duplicados del slice,
// comparando por número normalizado (sin espacios).
func deduplicateTelephones(telephones []models.Telephone) []models.Telephone {
	seen := make(map[string]struct{})
	result := make([]models.Telephone, 0, len(telephones))

	for _, t := range telephones {
		key := strings.ReplaceAll(t.NumeroTelefono, " ", "")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, t)
	}

	return result
}

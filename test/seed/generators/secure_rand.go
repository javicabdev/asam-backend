package generators

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	mathRand "math/rand"
	"time"
)

// GenerateSecureSeed generates a cryptographically secure seed for random number generation
func GenerateSecureSeed() int64 {
	var seed int64
	err := binary.Read(cryptoRand.Reader, binary.BigEndian, &seed)
	if err != nil {
		// Fallback to timestamp if crypto/rand fails
		seed = time.Now().UnixNano()
	}
	return seed
}

// NewSecureRand creates a new math/rand.Rand instance with a cryptographically secure seed
// This is intentionally using math/rand for test data generation, seeded with crypto/rand
//

func NewSecureRand() *mathRand.Rand {
	return mathRand.New(mathRand.NewSource(GenerateSecureSeed())) //nolint:gosec // Intentionally using math/rand for test data
}

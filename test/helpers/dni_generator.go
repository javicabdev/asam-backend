// Package helpers provides test helper functions
package helpers

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// letrasControl contains the control letters for Spanish DNI/NIE validation
const letrasControl = "TRWAGMYFPDXBNJZSQVHLCKE"

// GenerateValidDNI generates a valid Spanish DNI with the given number
// The number should be between 0 and 99999999
func GenerateValidDNI(number int) string {
	// Ensure number is within valid range
	if number < 0 {
		number = 0
	}
	if number > 99999999 {
		number = 99999999
	}

	// Calculate control letter
	resto := number % 23
	letra := string(letrasControl[resto])

	// Format with leading zeros (8 digits) + letter
	return fmt.Sprintf("%08d%s", number, letra)
}

// GenerateValidNIE generates a valid Spanish NIE with the given prefix and number
// Prefix should be 'X', 'Y', or 'Z'
// The number should be between 0 and 9999999 (7 digits)
func GenerateValidNIE(prefix rune, number int) string {
	// Validate prefix
	if prefix != 'X' && prefix != 'Y' && prefix != 'Z' {
		prefix = 'X' // Default to X
	}

	// Ensure number is within valid range
	if number < 0 {
		number = 0
	}
	if number > 9999999 {
		number = 9999999
	}

	// Convert prefix to its numeric value
	var prefixValue int
	switch prefix {
	case 'X':
		prefixValue = 0
	case 'Y':
		prefixValue = 1
	case 'Z':
		prefixValue = 2
	}

	// Construct the full number for calculation
	fullNumberStr := strconv.Itoa(prefixValue) + fmt.Sprintf("%07d", number)
	fullNumber, _ := strconv.Atoi(fullNumberStr)

	// Calculate control letter
	resto := fullNumber % 23
	letra := string(letrasControl[resto])

	// Format: prefix + 7 digits + letter
	return fmt.Sprintf("%c%07d%s", prefix, number, letra)
}

// GenerateRandomValidDNI generates a random valid Spanish DNI
func GenerateRandomValidDNI() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	number := random.Intn(100000000) // 0 to 99999999
	return GenerateValidDNI(number)
}

// GenerateRandomValidNIE generates a random valid Spanish NIE
func GenerateRandomValidNIE() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Random prefix
	prefixes := []rune{'X', 'Y', 'Z'}
	prefix := prefixes[random.Intn(3)]

	// Random number
	number := random.Intn(10000000) // 0 to 9999999

	return GenerateValidNIE(prefix, number)
}

// GenerateSequentialDNIs generates a slice of n valid sequential DNIs starting from the given number
func GenerateSequentialDNIs(startNumber int, count int) []string {
	dnis := make([]string, count)
	for i := 0; i < count; i++ {
		dnis[i] = GenerateValidDNI(startNumber + i)
	}
	return dnis
}

// GenerateSequentialNIEs generates a slice of n valid sequential NIEs with the given prefix
func GenerateSequentialNIEs(prefix rune, startNumber int, count int) []string {
	nies := make([]string, count)
	for i := 0; i < count; i++ {
		nies[i] = GenerateValidNIE(prefix, startNumber+i)
	}
	return nies
}

// GetValidTestDNIs returns a slice of commonly used valid DNIs for testing
func GetValidTestDNIs() []string {
	return []string{
		GenerateValidDNI(12345678), // 12345678Z
		GenerateValidDNI(87654321), // 87654321X
		GenerateValidDNI(11111111), // 11111111H
		GenerateValidDNI(99999999), // 99999999R
		GenerateValidDNI(0),        // 00000000T
		GenerateValidDNI(1),        // 00000001R
	}
}

// GetValidTestNIEs returns a slice of commonly used valid NIEs for testing
func GetValidTestNIEs() []string {
	return []string{
		GenerateValidNIE('X', 1234567), // X1234567L
		GenerateValidNIE('Y', 1234567), // Y1234567X
		GenerateValidNIE('Z', 1234567), // Z1234567M
		GenerateValidNIE('X', 0),       // X0000000T
		GenerateValidNIE('Y', 0),       // Y0000000Z
		GenerateValidNIE('Z', 0),       // Z0000000M
	}
}

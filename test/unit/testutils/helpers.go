package testutils

import (
	"fmt"
	"testing"
	"time"
)

// TimePtr returns a pointer to a time.Time value
// Useful for creating optional time fields in tests
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringPtr returns a pointer to a string value
// Useful for creating optional string fields in tests
func StringPtr(s string) *string {
	return &s
}

// UintPtr returns a pointer to a uint value
// Useful for creating optional uint fields in tests
func UintPtr(u uint) *uint {
	return &u
}

// IntPtr returns a pointer to an int value
// Useful for creating optional int fields in tests
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64 value
// Useful for creating optional float64 fields in tests
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool value
// Useful for creating optional bool fields in tests
func BoolPtr(b bool) *bool {
	return &b
}

// ParseTime parses a date string in the format YYYY-MM-DD and returns a time.Time
// Panics if the date cannot be parsed (only use in tests)
func ParseTime(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic("failed to parse time in test: " + err.Error())
	}
	return t
}

// ParseTimePtr is like ParseTime but returns a pointer
func ParseTimePtr(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	t := ParseTime(dateStr)
	return &t
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("unexpected error: %v - %v", err, msgAndArgs)
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("expected error but got nil - %v", msgAndArgs)
		} else {
			t.Fatal("expected error but got nil")
		}
	}
}

// ValidSpanishDNI returns a valid Spanish DNI for testing
// Uses a known valid DNI number with correct check digit
func ValidSpanishDNI() string {
	return "12345678Z" // This is a valid DNI format with correct check digit
}

// ValidSpanishNIE returns a valid Spanish NIE for testing
// Uses a known valid NIE number with correct check digit
func ValidSpanishNIE() string {
	return "X1234567L" // This is a valid NIE format with correct check digit
}

// InvalidSpanishDNI returns an invalid Spanish DNI for testing
func InvalidSpanishDNI() string {
	return "12345678A" // Invalid check digit
}

// GenerateMemberNumber generates a member number with the specified prefix
func GenerateMemberNumber(prefix string, number int) string {
	return prefix + fmt.Sprintf("%05d", number)
}

// CompareErrors compares two errors and returns true if they are equivalent
// This is useful when comparing domain errors that might have different pointer addresses
func CompareErrors(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

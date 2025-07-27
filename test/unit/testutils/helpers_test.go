package testutils_test

import (
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/test/unit/testutils"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected time.Time
	}{
		{
			name:     "valid date",
			dateStr:  "2024-01-15",
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "leap year date",
			dateStr:  "2020-02-29",
			expected: time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.ParseTime(tt.dateStr)
			if !result.Equal(tt.expected) {
				t.Errorf("ParseTime(%s) = %v, want %v", tt.dateStr, result, tt.expected)
			}
		})
	}
}

func TestParseTimePtr(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantNil bool
	}{
		{
			name:    "valid date returns pointer",
			dateStr: "2024-01-15",
			wantNil: false,
		},
		{
			name:    "empty string returns nil",
			dateStr: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.ParseTimePtr(tt.dateStr)
			if tt.wantNil {
				if result != nil {
					t.Errorf("ParseTimePtr(%s) = %v, want nil", tt.dateStr, result)
				}
			} else {
				if result == nil {
					t.Errorf("ParseTimePtr(%s) = nil, want non-nil", tt.dateStr)
				}
			}
		})
	}
}

func TestGenerateMemberNumber(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		number   int
		expected string
	}{
		{
			name:     "individual member",
			prefix:   "B",
			number:   1,
			expected: "B00001",
		},
		{
			name:     "family member",
			prefix:   "A",
			number:   99999,
			expected: "A99999",
		},
		{
			name:     "with leading zeros",
			prefix:   "B",
			number:   123,
			expected: "B00123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.GenerateMemberNumber(tt.prefix, tt.number)
			if result != tt.expected {
				t.Errorf("GenerateMemberNumber(%s, %d) = %s, want %s",
					tt.prefix, tt.number, result, tt.expected)
			}
		})
	}
}

func TestValidSpanishDocuments(t *testing.T) {
	// Test that the helper functions return valid formats
	dni := testutils.ValidSpanishDNI()
	if len(dni) != 9 || dni[len(dni)-1] < 'A' || dni[len(dni)-1] > 'Z' {
		t.Errorf("ValidSpanishDNI() = %s, expected format NNNNNNNNL", dni)
	}

	nie := testutils.ValidSpanishNIE()
	if len(nie) != 9 || (nie[0] != 'X' && nie[0] != 'Y' && nie[0] != 'Z') {
		t.Errorf("ValidSpanishNIE() = %s, expected format [X|Y|Z]NNNNNNNL", nie)
	}

	invalidDNI := testutils.InvalidSpanishDNI()
	if len(invalidDNI) != 9 {
		t.Errorf("InvalidSpanishDNI() = %s, expected 9 characters", invalidDNI)
	}
}

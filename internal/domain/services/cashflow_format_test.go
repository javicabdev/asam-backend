package services

import "testing"

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{12.5, "12.5"},
		{30, "30"},
		{100, "100"},
		{0.05, "0.05"},
		{1234.567, "1234.57"},
		{0, "0"},
		{-7, "-7"},
		{-0.5, "-0.5"},
	}
	for _, tt := range tests {
		if got := formatAmount(tt.amount); got != tt.want {
			t.Errorf("formatAmount(%v) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

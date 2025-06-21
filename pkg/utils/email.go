package utils

import (
	"regexp"
	"strings"
)

// EmailRegex is the regular expression used to validate email addresses
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsEmail checks if a given string is a valid email address
func IsEmail(str string) bool {
	str = strings.TrimSpace(str)
	return EmailRegex.MatchString(str)
}

// NormalizeEmail normalizes an email address by converting to lowercase
// and trimming whitespace
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ExtractUsernameFromEmail extracts the username part from an email address
// If the input is not a valid email, it returns the original string
func ExtractUsernameFromEmail(email string) string {
	if !IsEmail(email) {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}

	return email
}

// ObfuscateEmail partially hides an email address for privacy
// Example: john.doe@example.com -> j***e@example.com
func ObfuscateEmail(email string) string {
	if !IsEmail(email) {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) <= 2 {
		return localPart + "***@" + domainPart
	}

	return string(localPart[0]) + "***" + string(localPart[len(localPart)-1]) + "@" + domainPart
}

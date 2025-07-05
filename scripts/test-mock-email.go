package main

import (
	"context"
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/email"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

func main() {
	// Initialize logger
	logConfig := logger.DefaultConfig()
	logConfig.Development = true
	log, err := logger.InitLogger(logConfig)
	if err != nil {
		panic(err)
	}

	// Create mock email service
	mockService := email.NewMockNotificationAdapter(log)

	// Create test user
	email := "test@example.com"
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    &email,
	}

	ctx := context.Background()

	fmt.Println("Testing Mock Email Service...")
	fmt.Println("============================")

	// Test verification email
	fmt.Println("\n1. Testing Verification Email:")
	err = mockService.SendVerificationEmail(ctx, user, "https://example.com/verify?token=abc123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Verification email sent successfully")
	}

	// Test password reset email
	fmt.Println("\n2. Testing Password Reset Email:")
	err = mockService.SendPasswordResetEmail(ctx, user, "https://example.com/reset?token=xyz789")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Password reset email sent successfully")
	}

	// Test welcome email
	fmt.Println("\n3. Testing Welcome Email:")
	err = mockService.SendWelcomeEmail(ctx, user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Welcome email sent successfully")
	}

	// Test password changed email
	fmt.Println("\n4. Testing Password Changed Email:")
	err = mockService.SendPasswordChangedEmail(ctx, user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Password changed email sent successfully")
	}

	fmt.Println("\n============================")
	fmt.Println("All tests completed!")
	fmt.Println("\nCheck the logs to see the mock email content.")

	// Give time for logs to flush
	time.Sleep(100 * time.Millisecond)
}

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// MaintenanceResult represents the result of a maintenance task
type MaintenanceResult struct {
	TaskName  string
	StartTime time.Time
	EndTime   time.Time
	Message   string
	Error     error
	DryRun    bool
}

// MaintenanceService handles periodic maintenance tasks
type MaintenanceService struct {
	tokenRepo output.TokenRepository
	logger    logger.Logger
}

// NewMaintenanceService creates a new maintenance service
func NewMaintenanceService(tokenRepo output.TokenRepository, logger logger.Logger) *MaintenanceService {
	return &MaintenanceService{
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

// CleanupExpiredTokens removes all expired refresh tokens from the database
func (s *MaintenanceService) CleanupExpiredTokens(ctx context.Context) error {
	s.logger.Info("Starting expired token cleanup")

	startTime := time.Now()
	err := s.tokenRepo.CleanupExpiredTokens(ctx)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.Error("Failed to cleanup expired tokens",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return err
	}

	s.logger.Info("Expired token cleanup completed",
		zap.Duration("duration", duration),
	)

	return nil
}

// EnforceTokenLimitPerUser ensures no user has more than the specified number of active tokens
func (s *MaintenanceService) EnforceTokenLimitPerUser(ctx context.Context, maxTokens int) error {
	s.logger.Info("Starting token limit enforcement",
		zap.Int("max_tokens_per_user", maxTokens),
	)

	startTime := time.Now()
	err := s.tokenRepo.EnforceTokenLimitPerUser(ctx, maxTokens)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.Error("Failed to enforce token limit",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return err
	}

	s.logger.Info("Token limit enforcement completed",
		zap.Duration("duration", duration),
	)

	return nil
}

// RunScheduledMaintenance runs all scheduled maintenance tasks
func (s *MaintenanceService) RunScheduledMaintenance(ctx context.Context, maxTokensPerUser int) error {
	s.logger.Info("Starting scheduled maintenance")

	// Run cleanup tasks
	if err := s.CleanupExpiredTokens(ctx); err != nil {
		// Log error but continue with other tasks
		s.logger.Error("Token cleanup failed during scheduled maintenance", zap.Error(err))
	}

	// Enforce token limits
	if err := s.EnforceTokenLimitPerUser(ctx, maxTokensPerUser); err != nil {
		s.logger.Error("Token limit enforcement failed during scheduled maintenance", zap.Error(err))
		return err
	}

	s.logger.Info("Scheduled maintenance completed")
	return nil
}

// GenerateReport generates a report of maintenance results
func (s *MaintenanceService) GenerateReport(_ context.Context, results []MaintenanceResult) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("Maintenance Report - %s\n", time.Now().Format(time.RFC3339)))
	report.WriteString(strings.Repeat("-", 50) + "\n\n")

	for _, result := range results {
		report.WriteString(fmt.Sprintf("Task: %s\n", result.TaskName))
		report.WriteString(fmt.Sprintf("Start Time: %s\n", result.StartTime.Format(time.RFC3339)))
		report.WriteString(fmt.Sprintf("End Time: %s\n", result.EndTime.Format(time.RFC3339)))
		report.WriteString(fmt.Sprintf("Duration: %s\n", result.EndTime.Sub(result.StartTime)))

		if result.DryRun {
			report.WriteString("Mode: DRY RUN\n")
		}

		if result.Error != nil {
			report.WriteString("Status: FAILED\n")
			report.WriteString(fmt.Sprintf("Error: %v\n", result.Error))
		} else {
			report.WriteString("Status: SUCCESS\n")
		}

		if result.Message != "" {
			report.WriteString(fmt.Sprintf("Message: %s\n", result.Message))
		}

		report.WriteString("\n")
	}

	return report.String()
}

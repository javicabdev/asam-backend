package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// TokenCleanupService handles automatic cleanup of expired tokens
type TokenCleanupService struct {
	tokenRepo             output.TokenRepository
	verificationTokenRepo output.VerificationTokenRepository
	logger                logger.Logger
	interval              time.Duration
	maxTokensPerUser      int
	ticker                *time.Ticker
	done                  chan bool
}

// NewTokenCleanupService creates a new token cleanup service
func NewTokenCleanupService(
	tokenRepo output.TokenRepository,
	verificationTokenRepo output.VerificationTokenRepository,
	logger logger.Logger,
	interval time.Duration,
	maxTokensPerUser int,
) *TokenCleanupService {
	return &TokenCleanupService{
		tokenRepo:             tokenRepo,
		verificationTokenRepo: verificationTokenRepo,
		logger:                logger,
		interval:              interval,
		maxTokensPerUser:      maxTokensPerUser,
		done:                  make(chan bool),
	}
}

// Start begins the cleanup service
func (s *TokenCleanupService) Start(ctx context.Context) {
	s.logger.Info("Starting token cleanup service",
		zap.Duration("interval", s.interval),
		zap.Int("max_tokens_per_user", s.maxTokensPerUser),
	)

	// Run immediately on start
	s.performCleanup(ctx)

	// Then run periodically
	s.ticker = time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.performCleanup(ctx)
			case <-s.done:
				s.logger.Info("Token cleanup service stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Token cleanup service stopped due to context cancellation")
				return
			}
		}
	}()
}

// Stop stops the cleanup service
func (s *TokenCleanupService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.done)
}

// performCleanup executes the cleanup tasks
func (s *TokenCleanupService) performCleanup(ctx context.Context) {
	s.logger.Info("Performing token cleanup...")

	// 1. Clean expired refresh tokens
	if err := s.tokenRepo.CleanupExpiredTokens(ctx); err != nil {
		s.logger.Error("Error cleaning expired refresh tokens", zap.Error(err))
	} else {
		s.logger.Info("Expired refresh tokens cleaned successfully")
	}

	// 2. Clean expired verification tokens
	if s.verificationTokenRepo != nil {
		if err := s.verificationTokenRepo.DeleteExpired(ctx); err != nil {
			s.logger.Error("Error cleaning expired verification tokens", zap.Error(err))
		} else {
			s.logger.Info("Expired verification tokens cleaned successfully")
		}
	}

	// 3. If max tokens per user is set, enforce the limit for refresh tokens
	if s.maxTokensPerUser > 0 {
		s.enforceTokenLimitPerUser(ctx)
	}
}

// enforceTokenLimitPerUser ensures no user has more than the maximum allowed tokens
func (s *TokenCleanupService) enforceTokenLimitPerUser(_ context.Context) {
	// This would need to be implemented in the repository
	// For now, we'll log that this feature is pending
	s.logger.Debug("Token limit per user enforcement is pending implementation")
}

// CleanupNow performs an immediate cleanup (useful for manual triggers)
func (s *TokenCleanupService) CleanupNow(ctx context.Context) error {
	s.logger.Info("Manual token cleanup triggered")
	s.performCleanup(ctx)
	return nil
}

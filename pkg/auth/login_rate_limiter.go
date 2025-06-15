package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// LoginRateLimiter implements specialized rate limiting for login attempts
type LoginRateLimiter struct {
	attempts        map[string]*loginAttempts
	mtx             sync.RWMutex
	logger          logger.Logger
	maxAttempts     int
	lockoutDuration time.Duration
	windowDuration  time.Duration
	cleanupInterval time.Duration
}

type loginAttempts struct {
	count        int
	firstAttempt time.Time
	lastAttempt  time.Time
	lockedUntil  time.Time
	limiter      *rate.Limiter
}

// NewLoginRateLimiter creates a new login rate limiter
func NewLoginRateLimiter(logger logger.Logger) *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		attempts:        make(map[string]*loginAttempts),
		logger:          logger,
		maxAttempts:     5,                // 5 attempts
		lockoutDuration: 15 * time.Minute, // 15 minutes lockout
		windowDuration:  5 * time.Minute,  // within 5 minutes window
		cleanupInterval: 10 * time.Minute, // cleanup every 10 minutes
	}

	// Start cleanup routine
	go limiter.cleanup()
	return limiter
}

// NewLoginRateLimiterWithConfig creates a rate limiter with custom configuration
func NewLoginRateLimiterWithConfig(
	logger logger.Logger,
	maxAttempts int,
	lockoutDuration time.Duration,
	windowDuration time.Duration,
) *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		attempts:        make(map[string]*loginAttempts),
		logger:          logger,
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
		windowDuration:  windowDuration,
		cleanupInterval: 10 * time.Minute,
	}

	go limiter.cleanup()
	return limiter
}

// getKey generates a unique key for tracking attempts
func (l *LoginRateLimiter) getKey(ctx context.Context, identifier string) string {
	// Combine username with IP for better tracking
	ip := "unknown"
	if ipValue := ctx.Value("ip"); ipValue != nil {
		if ipStr, ok := ipValue.(string); ok {
			ip = ipStr
		}
	}
	return fmt.Sprintf("%s:%s", identifier, ip)
}

// AllowLogin checks if a login attempt is allowed
func (l *LoginRateLimiter) AllowLogin(ctx context.Context, identifier string) (bool, time.Duration) {
	key := l.getKey(ctx, identifier)

	l.mtx.Lock()
	defer l.mtx.Unlock()

	attempt, exists := l.attempts[key]
	now := time.Now()

	// If no previous attempts, create new entry
	if !exists {
		l.attempts[key] = &loginAttempts{
			count:        1,
			firstAttempt: now,
			lastAttempt:  now,
			limiter:      rate.NewLimiter(rate.Every(time.Second), 1), // 1 per second burst protection
		}
		return true, 0
	}

	// Check if account is locked
	if now.Before(attempt.lockedUntil) {
		remainingLockout := attempt.lockedUntil.Sub(now)
		l.logger.Warn("Login attempt denied - account locked",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Duration("remaining_lockout", remainingLockout),
		)
		return false, remainingLockout
	}

	// Check burst protection
	if !attempt.limiter.Allow() {
		l.logger.Warn("Login attempt denied - too fast",
			zap.String("identifier", identifier),
			zap.String("key", key),
		)
		return false, time.Second
	}

	// Reset counter if outside window
	if now.Sub(attempt.firstAttempt) > l.windowDuration {
		attempt.count = 1
		attempt.firstAttempt = now
		attempt.lastAttempt = now
		attempt.lockedUntil = time.Time{}
		return true, 0
	}

	// Increment attempt counter
	attempt.count++
	attempt.lastAttempt = now

	// Check if max attempts exceeded
	if attempt.count >= l.maxAttempts {
		attempt.lockedUntil = now.Add(l.lockoutDuration)
		l.logger.Error("Max login attempts exceeded - account locked",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Int("attempts", attempt.count),
			zap.Duration("lockout_duration", l.lockoutDuration),
		)
		return false, l.lockoutDuration
	}

	// Log warning as attempts increase
	if attempt.count >= 3 {
		l.logger.Warn("Multiple login attempts detected",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Int("attempts", attempt.count),
			zap.Int("remaining", l.maxAttempts-attempt.count),
		)
	}

	return true, 0
}

// RecordSuccess records a successful login and resets the counter
func (l *LoginRateLimiter) RecordSuccess(ctx context.Context, identifier string) {
	key := l.getKey(ctx, identifier)

	l.mtx.Lock()
	defer l.mtx.Unlock()

	delete(l.attempts, key)

	l.logger.Info("Login successful - resetting attempt counter",
		zap.String("identifier", identifier),
		zap.String("key", key),
	)
}

// RecordFailure records a failed login attempt
func (l *LoginRateLimiter) RecordFailure(ctx context.Context, identifier string) {
	// AllowLogin already increments the counter, but we can use this
	// for additional logging or metrics
	key := l.getKey(ctx, identifier)

	l.mtx.RLock()
	attempt, exists := l.attempts[key]
	l.mtx.RUnlock()

	if exists {
		l.logger.Info("Login failure recorded",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Int("total_attempts", attempt.count),
		)
	}
}

// GetAttemptInfo returns information about login attempts for an identifier
func (l *LoginRateLimiter) GetAttemptInfo(ctx context.Context, identifier string) (attempts int, lockedUntil time.Time, exists bool) {
	key := l.getKey(ctx, identifier)

	l.mtx.RLock()
	defer l.mtx.RUnlock()

	attempt, exists := l.attempts[key]
	if !exists {
		return 0, time.Time{}, false
	}

	return attempt.count, attempt.lockedUntil, true
}

// Reset clears all attempts for an identifier (admin function)
func (l *LoginRateLimiter) Reset(ctx context.Context, identifier string) {
	key := l.getKey(ctx, identifier)

	l.mtx.Lock()
	defer l.mtx.Unlock()

	delete(l.attempts, key)

	l.logger.Info("Login attempts reset",
		zap.String("identifier", identifier),
		zap.String("key", key),
	)
}

// cleanup removes old entries periodically
func (l *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mtx.Lock()
		now := time.Now()
		cleaned := 0

		for key, attempt := range l.attempts {
			// Remove entries that are:
			// 1. Not locked and haven't been used for twice the window duration
			// 2. Locked but the lockout has expired for more than the window duration
			if attempt.lockedUntil.IsZero() && now.Sub(attempt.lastAttempt) > 2*l.windowDuration {
				delete(l.attempts, key)
				cleaned++
			} else if !attempt.lockedUntil.IsZero() && now.Sub(attempt.lockedUntil) > l.windowDuration {
				delete(l.attempts, key)
				cleaned++
			}
		}

		if cleaned > 0 {
			l.logger.Debug("Login rate limiter cleanup completed",
				zap.Int("cleaned", cleaned),
				zap.Int("remaining", len(l.attempts)),
			)
		}

		l.mtx.Unlock()
	}
}

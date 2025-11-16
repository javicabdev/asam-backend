package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// BackupService handles automatic database backups
type BackupService struct {
	dbHost       string
	dbPort       string
	dbUser       string
	dbPassword   string
	dbName       string
	storage      BackupStorage
	maxRetention int
	environment  string
	logger       logger.Logger
	interval     time.Duration
	ticker       *time.Ticker
	done         chan bool
	mu           sync.Mutex // Protects ticker
}

// NewBackupService creates a new backup service
func NewBackupService(
	dbHost string,
	dbPort string,
	dbUser string,
	dbPassword string,
	dbName string,
	storage BackupStorage,
	maxRetention int,
	environment string,
	logger logger.Logger,
	interval time.Duration,
) *BackupService {
	return &BackupService{
		dbHost:       dbHost,
		dbPort:       dbPort,
		dbUser:       dbUser,
		dbPassword:   dbPassword,
		dbName:       dbName,
		storage:      storage,
		maxRetention: maxRetention,
		environment:  environment,
		logger:       logger,
		interval:     interval,
		done:         make(chan bool),
	}
}

// Start begins the backup service
func (s *BackupService) Start(ctx context.Context) {
	s.logger.Info("Starting database backup service",
		zap.Duration("interval", s.interval),
		zap.Int("max_retention", s.maxRetention),
		zap.String("environment", s.environment),
	)

	// Run immediately on start
	s.performBackup(ctx)

	// Then run periodically
	s.mu.Lock()
	s.ticker = time.NewTicker(s.interval)
	ticker := s.ticker
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-ticker.C:
				s.performBackup(ctx)
			case <-s.done:
				s.logger.Info("Database backup service stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Database backup service stopped due to context cancellation")
				return
			}
		}
	}()
}

// Stop stops the backup service
func (s *BackupService) Stop() {
	s.mu.Lock()
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.mu.Unlock()

	close(s.done)

	// Close storage connection
	if err := s.storage.Close(); err != nil {
		s.logger.Error("Failed to close storage connection", zap.Error(err))
	}
}

// performBackup executes the backup process
func (s *BackupService) performBackup(ctx context.Context) {
	s.logger.Info("Starting database backup...")
	startTime := time.Now()

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s.dump", s.environment, timestamp)

	// Create backup in temporary directory
	tmpDir, err := os.MkdirTemp("", "asam-backup-*")
	if err != nil {
		s.logger.Error("Failed to create temporary directory", zap.Error(err))
		return
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.logger.Warn("Failed to remove temporary directory", zap.Error(err))
		}
	}()

	backupPath := filepath.Join(tmpDir, filename)

	// Create backup using pg_dump
	if err := s.createBackup(ctx, backupPath); err != nil {
		s.logger.Error("Failed to create backup", zap.Error(err))
		return
	}

	// Upload to storage
	if err := s.storage.Upload(ctx, backupPath, filename); err != nil {
		s.logger.Error("Failed to upload backup to storage", zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	s.logger.Info("Backup created and uploaded successfully",
		zap.String("filename", filename),
		zap.Duration("duration", duration),
	)

	// Cleanup old backups (only if maxRetention > 0)
	// Si maxRetention es 0, usa GCS lifecycle policies en su lugar
	if s.maxRetention > 0 {
		if err := s.cleanupOldBackups(ctx); err != nil {
			s.logger.Error("Failed to cleanup old backups", zap.Error(err))
		}
	} else {
		s.logger.Debug("Backup retention disabled - using storage lifecycle policies")
	}
}

// createBackup creates a database backup using pg_dump
func (s *BackupService) createBackup(ctx context.Context, backupPath string) error {
	// Build the pg_dump command args
	args := []string{
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"--no-owner",
		"--no-acl",
		"-F", "c",
		"-f", backupPath,
	}

	// Execute backup with PGPASSWORD environment variable
	// #nosec G204 -- args are constructed safely from validated config
	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w (output: %s)", err, string(output))
	}

	// Verify backup file was created
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not created: %w", err)
	}

	// Get file size for logging
	fileInfo, err := os.Stat(backupPath)
	if err == nil {
		s.logger.Info("Backup file created locally",
			zap.String("path", backupPath),
			zap.Int64("size_bytes", fileInfo.Size()),
		)
	}

	return nil
}

// cleanupOldBackups removes old backups keeping only the most recent ones
func (s *BackupService) cleanupOldBackups(ctx context.Context) error {
	s.logger.Info("Cleaning up old backups",
		zap.Int("max_retention", s.maxRetention),
	)

	// Get list of backup files from storage
	backups, err := s.storage.List(ctx, s.environment)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// No cleanup needed if we have fewer backups than the max retention
	if len(backups) <= s.maxRetention {
		s.logger.Info("No cleanup needed",
			zap.Int("current_backups", len(backups)),
			zap.Int("max_retention", s.maxRetention),
		)
		return nil
	}

	// Backups are already sorted by modification time (newest first)
	// Remove old backups (keep only maxRetention)
	removed := 0
	for i := s.maxRetention; i < len(backups); i++ {
		if err := s.storage.Delete(ctx, backups[i].Name); err != nil {
			s.logger.Error("Failed to remove old backup",
				zap.String("file", backups[i].Name),
				zap.Error(err),
			)
			continue
		}
		removed++
		s.logger.Info("Removed old backup",
			zap.String("file", backups[i].Name),
			zap.Time("modified", backups[i].ModTime),
		)
	}

	s.logger.Info("Cleanup completed",
		zap.Int("removed", removed),
		zap.Int("remaining", len(backups)-removed),
	)

	return nil
}

// BackupNow performs an immediate backup (useful for manual triggers)
func (s *BackupService) BackupNow(ctx context.Context) error {
	s.logger.Info("Manual backup triggered")
	s.performBackup(ctx)
	return nil
}

// GetBackupInfo returns information about existing backups
func (s *BackupService) GetBackupInfo(ctx context.Context) ([]BackupFileInfo, error) {
	return s.storage.List(ctx, s.environment)
}

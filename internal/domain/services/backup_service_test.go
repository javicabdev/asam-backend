package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

func TestBackupService_CleanupOldBackups(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create test logger
	cfg := logger.DefaultConfig()
	cfg.Development = true
	testLogger, err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create filesystem storage
	storage, err := NewFilesystemStorage(tmpDir, testLogger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	// Create backup service
	service := NewBackupService(
		"localhost",
		"5432",
		"postgres",
		"postgres",
		"test_db",
		storage,
		2, // Keep only 2 backups
		"test",
		testLogger,
		1*time.Hour,
	)

	// Create test backup files
	testFiles := []string{
		"backup_test_20251116_100000.dump",
		"backup_test_20251116_110000.dump",
		"backup_test_20251116_120000.dump",
		"backup_test_20251116_130000.dump",
	}

	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test backup data"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		// Sleep to ensure different modification times
		time.Sleep(10 * time.Millisecond)
	}

	ctx := context.Background()

	// Run cleanup
	if err := service.cleanupOldBackups(ctx); err != nil {
		t.Fatalf("cleanupOldBackups failed: %v", err)
	}

	// Verify only 2 backups remain
	files, err := filepath.Glob(filepath.Join(tmpDir, "backup_test_*.dump"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 backups to remain, got %d", len(files))
	}

	// Verify the newest files were kept
	expectedFiles := map[string]bool{
		"backup_test_20251116_120000.dump": true,
		"backup_test_20251116_130000.dump": true,
	}

	for _, file := range files {
		filename := filepath.Base(file)
		if !expectedFiles[filename] {
			t.Errorf("Unexpected file kept: %s", filename)
		}
	}
}

func TestBackupService_StartStop(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create test logger
	cfg := logger.DefaultConfig()
	cfg.Development = true
	testLogger, err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create filesystem storage
	storage, err := NewFilesystemStorage(tmpDir, testLogger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	// Create backup service with short interval
	service := NewBackupService(
		"localhost",
		"5432",
		"postgres",
		"postgres",
		"test_db",
		storage,
		2,
		"test",
		testLogger,
		100*time.Millisecond, // Very short interval for testing
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service (but it will fail to create backups since DB doesn't exist)
	// This is OK - we're just testing the start/stop mechanism
	go service.Start(ctx)

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Stop service
	service.Stop()

	// If we get here without panic, the start/stop mechanism works
	t.Log("Service started and stopped successfully")
}

func TestBackupService_GetBackupInfo(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create test logger
	cfg := logger.DefaultConfig()
	cfg.Development = true
	testLogger, err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create filesystem storage
	storage, err := NewFilesystemStorage(tmpDir, testLogger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	// Create backup service
	service := NewBackupService(
		"localhost",
		"5432",
		"postgres",
		"postgres",
		"test_db",
		storage,
		2,
		"test",
		testLogger,
		1*time.Hour,
	)

	// Create test backup files
	testFiles := []string{
		"backup_test_20251116_100000.dump",
		"backup_test_20251116_110000.dump",
	}

	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test backup data"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		// Sleep to ensure different modification times
		time.Sleep(10 * time.Millisecond)
	}

	ctx := context.Background()

	// Get backup info
	backups, err := service.GetBackupInfo(ctx)
	if err != nil {
		t.Fatalf("GetBackupInfo failed: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify files are sorted (newest first)
	if len(backups) == 2 {
		if backups[0].Name != "backup_test_20251116_110000.dump" {
			t.Errorf("Expected newest backup first, got %s", backups[0].Name)
		}
	}
}

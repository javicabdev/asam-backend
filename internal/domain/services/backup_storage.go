package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// BackupStorage defines the interface for backup storage backends
type BackupStorage interface {
	// Upload uploads a backup file to the storage
	Upload(ctx context.Context, localPath, filename string) error
	// List returns a list of backup files sorted by modification time (newest first)
	List(ctx context.Context, environment string) ([]BackupFileInfo, error)
	// Delete removes a backup file from storage
	Delete(ctx context.Context, filename string) error
	// Close closes the storage connection
	Close() error
}

// BackupFileInfo contains information about a backup file
type BackupFileInfo struct {
	Name    string
	ModTime time.Time
	Size    int64
}

// FilesystemStorage implements BackupStorage for local filesystem
type FilesystemStorage struct {
	backupDir string
	logger    logger.Logger
}

// NewFilesystemStorage creates a new filesystem storage backend
func NewFilesystemStorage(backupDir string, logger logger.Logger) (*FilesystemStorage, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &FilesystemStorage{
		backupDir: backupDir,
		logger:    logger,
	}, nil
}

// Upload copies a file to the backup directory
func (fs *FilesystemStorage) Upload(ctx context.Context, localPath, filename string) error {
	destPath := filepath.Join(fs.backupDir, filename)

	// Since the file is already in the backup directory (created by pg_dump),
	// we just need to verify it exists
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	fs.logger.Info("Backup file saved to filesystem",
		zap.String("path", destPath),
	)

	return nil
}

// List returns all backup files for the given environment
func (fs *FilesystemStorage) List(ctx context.Context, environment string) ([]BackupFileInfo, error) {
	pattern := filepath.Join(fs.backupDir, fmt.Sprintf("backup_%s_*.dump", environment))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	backups := make([]BackupFileInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fs.logger.Warn("Failed to stat backup file", zap.String("file", file), zap.Error(err))
			continue
		}
		backups = append(backups, BackupFileInfo{
			Name:    filepath.Base(file),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// Delete removes a backup file
func (fs *FilesystemStorage) Delete(ctx context.Context, filename string) error {
	path := filepath.Join(fs.backupDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}
	return nil
}

// Close is a no-op for filesystem storage
func (fs *FilesystemStorage) Close() error {
	return nil
}

// GCSStorage implements BackupStorage for Google Cloud Storage
type GCSStorage struct {
	client     *storage.Client
	bucketName string
	prefix     string // Optional prefix for organizing backups
	logger     logger.Logger
}

// NewGCSStorage creates a new Google Cloud Storage backend
func NewGCSStorage(ctx context.Context, bucketName, prefix string, logger logger.Logger) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Verify bucket exists and is accessible
	bucket := client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			logger.Error("Failed to close GCS client after bucket access error", zap.Error(closeErr))
		}
		return nil, fmt.Errorf("failed to access bucket %s: %w", bucketName, err)
	}

	logger.Info("Connected to Google Cloud Storage",
		zap.String("bucket", bucketName),
		zap.String("prefix", prefix),
	)

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
		prefix:     prefix,
		logger:     logger,
	}, nil
}

// Upload uploads a file to Google Cloud Storage
func (gcs *GCSStorage) Upload(ctx context.Context, localPath, filename string) error {
	// Open local file
	// #nosec G304 -- localPath is from trusted temporary directory created by the app
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			gcs.logger.Warn("Failed to close local file", zap.Error(err))
		}
	}()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Construct object name with prefix
	objectName := filename
	if gcs.prefix != "" {
		objectName = gcs.prefix + "/" + filename
	}

	// Create GCS object
	bucket := gcs.client.Bucket(gcs.bucketName)
	object := bucket.Object(objectName)
	writer := object.NewWriter(ctx)

	// Set metadata
	writer.ContentType = "application/octet-stream"
	writer.Metadata = map[string]string{
		"created-by": "asam-backend",
		"backup-env": strings.Split(filename, "_")[1], // Extract environment from filename
	}

	// Copy file to GCS
	gcs.logger.Info("Uploading backup to Google Cloud Storage...",
		zap.String("bucket", gcs.bucketName),
		zap.String("object", objectName),
		zap.Int64("size_bytes", fileInfo.Size()),
	)

	startTime := time.Now()
	written, err := io.Copy(writer, file)
	if err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			gcs.logger.Warn("Failed to close GCS writer after upload error", zap.Error(closeErr))
		}
		return fmt.Errorf("failed to upload to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	duration := time.Since(startTime)
	gcs.logger.Info("Backup uploaded successfully to GCS",
		zap.String("object", objectName),
		zap.Int64("bytes_written", written),
		zap.Duration("duration", duration),
	)

	return nil
}

// List returns all backup files for the given environment
func (gcs *GCSStorage) List(ctx context.Context, environment string) ([]BackupFileInfo, error) {
	bucket := gcs.client.Bucket(gcs.bucketName)

	// Build query prefix
	queryPrefix := gcs.prefix
	if queryPrefix != "" && !strings.HasSuffix(queryPrefix, "/") {
		queryPrefix += "/"
	}
	queryPrefix += fmt.Sprintf("backup_%s_", environment)

	// List objects
	query := &storage.Query{Prefix: queryPrefix}
	it := bucket.Objects(ctx, query)

	var backups []BackupFileInfo
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}

		// Extract filename from full path
		filename := attrs.Name
		if gcs.prefix != "" {
			filename = strings.TrimPrefix(filename, gcs.prefix+"/")
		}

		backups = append(backups, BackupFileInfo{
			Name:    filename,
			ModTime: attrs.Updated,
			Size:    attrs.Size,
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// Delete removes a backup file from GCS
func (gcs *GCSStorage) Delete(ctx context.Context, filename string) error {
	objectName := filename
	if gcs.prefix != "" {
		objectName = gcs.prefix + "/" + filename
	}

	bucket := gcs.client.Bucket(gcs.bucketName)
	object := bucket.Object(objectName)

	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object from GCS: %w", err)
	}

	gcs.logger.Info("Deleted backup from GCS",
		zap.String("object", objectName),
	)

	return nil
}

// Close closes the GCS client
func (gcs *GCSStorage) Close() error {
	return gcs.client.Close()
}

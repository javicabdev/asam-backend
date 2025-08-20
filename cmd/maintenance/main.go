// Package main provides a maintenance command for executing periodic cleanup tasks.
// This command can be run as a standalone process or scheduled via cron.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// maintenanceFlags holds command-line flags
type maintenanceFlags struct {
	cleanupTokens  bool
	enforceLimit   bool
	runAll         bool
	dryRun         bool
	customLimit    int
	generateReport bool
}

func main() {
	// Parse command-line flags
	flags := parseFlags()

	// Initialize application components
	cfg, log, maintenanceService, err := initializeApplication()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	startTime := time.Now()

	// Determine token limit
	tokenLimit := determineTokenLimit(cfg, flags)

	log.Info("Starting maintenance tasks",
		zap.Bool("dry_run", flags.dryRun),
		zap.Bool("cleanup_tokens", flags.cleanupTokens || flags.runAll),
		zap.Bool("enforce_limit", flags.enforceLimit || flags.runAll),
		zap.Int("token_limit", tokenLimit),
	)

	// Execute maintenance tasks
	results := executeTasks(ctx, log, maintenanceService, flags, tokenLimit)

	// Generate report and exit
	handleCompletion(log, maintenanceService, results, flags, startTime)
}

// executeTask executes a maintenance task and returns the result
func executeTask(_ context.Context, taskName string, dryRun bool, task func() error) (services.MaintenanceResult, error) {
	result := services.MaintenanceResult{
		TaskName:  taskName,
		StartTime: time.Now(),
		DryRun:    dryRun,
	}

	if dryRun {
		result.Message = "Dry run - no changes made"
		result.EndTime = time.Now()
		return result, nil
	}

	err := task()
	result.EndTime = time.Now()
	result.Error = err

	if err != nil {
		result.Message = fmt.Sprintf("Task failed: %v", err)
	} else {
		result.Message = "Task completed successfully"
	}

	return result, err
}

// parseFlags parses command-line flags
func parseFlags() maintenanceFlags {
	// Define command-line flags
	cleanupTokens := flag.Bool("cleanup-tokens", false, "Clean up expired tokens")
	enforceLimit := flag.Bool("enforce-token-limit", false, "Enforce token limit per user")
	runAll := flag.Bool("all", false, "Run all maintenance tasks")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without executing")
	customLimit := flag.Int("token-limit", 0, "Custom token limit (overrides config)")
	generateReport := flag.Bool("report", false, "Generate maintenance report")

	flag.Parse()

	return maintenanceFlags{
		cleanupTokens:  *cleanupTokens,
		enforceLimit:   *enforceLimit,
		runAll:         *runAll,
		dryRun:         *dryRun,
		customLimit:    *customLimit,
		generateReport: *generateReport,
	}
}

// initializeApplication initializes all application components
func initializeApplication() (*config.Config, logger.Logger, *services.MaintenanceService, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logCfg := logger.DefaultConfig()
	log, err := logger.InitLogger(logCfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database
	database, err := db.InitDB(cfg, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize repositories
	tokenRepo := db.NewTokenRepository(database)

	// Initialize maintenance service
	maintenanceService := services.NewMaintenanceService(tokenRepo, log)

	return cfg, log, maintenanceService, nil
}

// determineTokenLimit determines the token limit based on config and flags
func determineTokenLimit(cfg *config.Config, flags maintenanceFlags) int {
	if flags.customLimit > 0 {
		return flags.customLimit
	}
	return cfg.MaxTokensPerUser
}

// executeTasks executes all requested maintenance tasks
func executeTasks(ctx context.Context, log logger.Logger, maintenanceService *services.MaintenanceService, flags maintenanceFlags, tokenLimit int) []services.MaintenanceResult {
	var results []services.MaintenanceResult

	// Execute token cleanup if requested
	if flags.cleanupTokens || flags.runAll {
		result, err := executeTask(ctx, "Token Cleanup", flags.dryRun, func() error {
			return maintenanceService.CleanupExpiredTokens(ctx)
		})
		if err != nil {
			log.Error("Token cleanup failed", zap.Error(err))
		}
		results = append(results, result)
	}

	// Enforce token limit if requested
	if flags.enforceLimit || flags.runAll {
		result, err := executeTask(ctx, "Enforce Token Limit", flags.dryRun, func() error {
			return maintenanceService.EnforceTokenLimitPerUser(ctx, tokenLimit)
		})
		if err != nil {
			log.Error("Token limit enforcement failed", zap.Error(err))
		}
		results = append(results, result)
	}

	return results
}

// handleCompletion handles report generation and exit code
func handleCompletion(log logger.Logger, maintenanceService *services.MaintenanceService, results []services.MaintenanceResult, flags maintenanceFlags, startTime time.Time) {
	// Generate report if requested
	if flags.generateReport {
		report := maintenanceService.GenerateReport(context.Background(), results)
		fmt.Println("\n=== Maintenance Report ===")
		fmt.Println(report)
	}

	// Log completion
	duration := time.Since(startTime)
	log.Info("Maintenance tasks completed",
		zap.Duration("duration", duration),
		zap.Int("tasks_executed", len(results)),
	)

	// Exit with appropriate code
	for _, result := range results {
		if result.Error != nil {
			os.Exit(1)
		}
	}
}

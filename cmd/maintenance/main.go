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

func main() {
	// Define command-line flags
	var (
		cleanupTokens  = flag.Bool("cleanup-tokens", false, "Clean up expired tokens")
		enforceLimit   = flag.Bool("enforce-token-limit", false, "Enforce token limit per user")
		runAll         = flag.Bool("all", false, "Run all maintenance tasks")
		dryRun         = flag.Bool("dry-run", false, "Show what would be done without executing")
		customLimit    = flag.Int("token-limit", 0, "Custom token limit (overrides config)")
		generateReport = flag.Bool("report", false, "Generate maintenance report")
	)

	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logCfg := logger.DefaultConfig()
	log, err := logger.InitLogger(logCfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.InitDB(cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize repositories
	tokenRepo := db.NewTokenRepository(database)

	// Initialize maintenance service
	maintenanceService := services.NewMaintenanceService(tokenRepo, log)

	ctx := context.Background()
	startTime := time.Now()

	// Determine token limit
	tokenLimit := cfg.MaxTokensPerUser
	if *customLimit > 0 {
		tokenLimit = *customLimit
	}

	log.Info("Starting maintenance tasks",
		zap.Bool("dry_run", *dryRun),
		zap.Bool("cleanup_tokens", *cleanupTokens || *runAll),
		zap.Bool("enforce_limit", *enforceLimit || *runAll),
		zap.Int("token_limit", tokenLimit),
	)

	// Track results
	var results []services.MaintenanceResult

	// Execute requested maintenance tasks
	if *cleanupTokens || *runAll {
		result, err := executeTask(ctx, "Token Cleanup", *dryRun, func() error {
			return maintenanceService.CleanupExpiredTokens(ctx)
		})
		if err != nil {
			log.Error("Token cleanup failed", zap.Error(err))
		}
		results = append(results, result)
	}

	if *enforceLimit || *runAll {
		result, err := executeTask(ctx, "Enforce Token Limit", *dryRun, func() error {
			return maintenanceService.EnforceTokenLimitPerUser(ctx, tokenLimit)
		})
		if err != nil {
			log.Error("Token limit enforcement failed", zap.Error(err))
		}
		results = append(results, result)
	}

	// Generate report if requested
	if *generateReport {
		report := maintenanceService.GenerateReport(ctx, results)
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

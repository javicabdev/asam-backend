package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	appLogger "github.com/javicabdev/asam-backend/pkg/logger"
)

// QueryMonitor implements a GORM logger interface to monitor slow queries
type QueryMonitor struct {
	logger         appLogger.Logger
	SlowThreshold  time.Duration
	LogLevel       logger.LogLevel
	IgnoreNotFound bool
	slowQueries    map[string]int // Map to track frequency of slow queries
}

// NewQueryMonitor creates a new QueryMonitor
func NewQueryMonitor(log appLogger.Logger, slowThreshold time.Duration) *QueryMonitor {
	return &QueryMonitor{
		logger:         log,
		SlowThreshold:  slowThreshold,
		LogLevel:       logger.Info,
		IgnoreNotFound: true,
		slowQueries:    make(map[string]int),
	}
}

// LogMode sets the log level
func (m *QueryMonitor) LogMode(level logger.LogLevel) logger.Interface {
	m.LogLevel = level
	return m
}

// Info logs info messages
func (m *QueryMonitor) Info(_ context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= logger.Info {
		m.logger.Info(fmt.Sprintf(msg, args...))
	}
}

// Warn logs warn messages
func (m *QueryMonitor) Warn(_ context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= logger.Warn {
		m.logger.Warn(fmt.Sprintf(msg, args...))
	}
}

// Error logs error messages
func (m *QueryMonitor) Error(_ context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= logger.Error {
		m.logger.Error(fmt.Sprintf(msg, args...))
	}
}

// Trace logs SQL statements and execution time
func (m *QueryMonitor) Trace(_ context.Context, begin time.Time, fc func() (string, int64), _ error) {
	if m.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	
	// Normalize the query for better grouping (remove specific values)
	normalizedSQL := normalizeSQL(sql)

	if elapsed > m.SlowThreshold {
		// Increment counter for this slow query
		m.slowQueries[normalizedSQL]++
		
		// Log slow query with additional context
		m.logger.Warn("SLOW SQL QUERY",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
			zap.Int("occurrence", m.slowQueries[normalizedSQL]),
		)
	} else if m.LogLevel >= logger.Info {
		m.logger.Debug("SQL QUERY",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}

// GetSlowQueriesReport returns a report of all slow queries and their frequency
func (m *QueryMonitor) GetSlowQueriesReport() map[string]int {
	return m.slowQueries
}

// normalizeSQL simplifies SQL queries for better grouping by replacing literal values
func normalizeSQL(sql string) string {
	// Replace numeric literals
	sql = strings.ReplaceAll(sql, "?", "?")
	
	// You can add more normalization rules as needed
	return sql
}

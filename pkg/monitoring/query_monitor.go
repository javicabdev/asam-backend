package monitoring

import (
	"context"
	"sync"
	"time"
	
	"go.uber.org/zap"
	"gorm.io/gorm"
	
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// QueryStat tracks query performance statistics
type QueryStat struct {
	Query       string
	Count       int
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	AvgTime     time.Duration
	LastSeen    time.Time
}

// QueryMonitor monitors database query performance
type QueryMonitor struct {
	db           *gorm.DB
	logger       logger.Logger
	mu           sync.RWMutex
	slowQueries  map[string]int
	queryStats   map[string]*QueryStat
	slowThreshold time.Duration
}

// NewQueryMonitor creates a new query monitor
func NewQueryMonitor(db *gorm.DB, log logger.Logger, slowThreshold time.Duration) *QueryMonitor {
	return &QueryMonitor{
		db:            db,
		logger:        log,
		slowQueries:   make(map[string]int),
		queryStats:    make(map[string]*QueryStat),
		slowThreshold: slowThreshold,
	}
}

// TrackQuery records performance metrics for a query
func (m *QueryMonitor) TrackQuery(ctx context.Context, query string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record slow query if it exceeds the threshold
	if duration >= m.slowThreshold {
		m.slowQueries[query]++
		m.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Int("occurrences", m.slowQueries[query]),
		)
	}
	
	// Update query statistics
	stat, exists := m.queryStats[query]
	if !exists {
		stat = &QueryStat{
			Query:   query,
			MinTime: duration,
			MaxTime: duration,
		}
		m.queryStats[query] = stat
	}
	
	stat.Count++
	stat.TotalTime += duration
	stat.AvgTime = stat.TotalTime / time.Duration(stat.Count)
	stat.LastSeen = time.Now()
	
	if duration < stat.MinTime {
		stat.MinTime = duration
	}
	if duration > stat.MaxTime {
		stat.MaxTime = duration
	}
}

// GetSlowQueriesReport returns a report of all slow queries and their frequency
func (m *QueryMonitor) GetSlowQueriesReport() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Make a copy to avoid concurrent map access
	report := make(map[string]int, len(m.slowQueries))
	for query, count := range m.slowQueries {
		report[query] = count
	}
	
	return report
}

// GetQueryStats returns statistics for all tracked queries
func (m *QueryMonitor) GetQueryStats() map[string]QueryStat {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Make a copy to avoid concurrent map access
	stats := make(map[string]QueryStat, len(m.queryStats))
	for query, stat := range m.queryStats {
		stats[query] = *stat
	}
	
	return stats
}

// ResetStats resets all query statistics
func (m *QueryMonitor) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.slowQueries = make(map[string]int)
	m.queryStats = make(map[string]*QueryStat)
}

// SetupQueryMonitoring configures GORM hooks to track query performance
func SetupQueryMonitoring(db *gorm.DB, logger logger.Logger, slowThreshold time.Duration) *QueryMonitor {
	monitor := NewQueryMonitor(db, logger, slowThreshold)
	
	// Register callbacks to track query performance
	_ = db.Callback().Query().Before("gorm:query").Register("monitor:before_query", func(db *gorm.DB) {
		db.Set("query_start_time", time.Now())
	})
	
	_ = db.Callback().Query().After("gorm:query").Register("monitor:after_query", func(db *gorm.DB) {
		// Get the start time from the DB context
		startTimeValue, exists := db.Get("query_start_time")
		if !exists {
			return
		}
		
		startTime, ok := startTimeValue.(time.Time)
		if !ok {
			return
		}
		
		// Calculate query duration
		duration := time.Since(startTime)
		
		// Track the query performance
		if db.Statement != nil && db.Statement.SQL.String() != "" {
			monitor.TrackQuery(db.Statement.Context, db.Statement.SQL.String(), duration)
		}
	})
	
	return monitor
}

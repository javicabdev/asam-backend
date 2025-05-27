package monitoring

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	
	"go.uber.org/zap"
	
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// MemoryStats contains information about memory usage
type MemoryStats struct {
	Alloc        uint64  `json:"alloc"`
	TotalAlloc   uint64  `json:"totalAlloc"`
	Sys          uint64  `json:"sys"`
	NumGC        uint32  `json:"numGC"`
	PauseTotalNs uint64  `json:"pauseTotalNs"`
	HeapObjects  uint64  `json:"heapObjects"`
	GCCPUFraction float64 `json:"gcCPUFraction"`
}

// MemoryMonitor monitors memory usage and triggers alerts when thresholds are exceeded
type MemoryMonitor struct {
	logger           logger.Logger
	alertThreshold   uint64 // MB
	criticalThreshold uint64 // MB
	monitorInterval   time.Duration
	outputDir         string
	ctx               context.Context
	cancel            context.CancelFunc
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(
	log logger.Logger,
	alertThreshold uint64,
	criticalThreshold uint64,
	monitorInterval time.Duration,
	outputDir string,
) *MemoryMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Error("Failed to create memory profile output directory", zap.Error(err))
		}
	}
	
	return &MemoryMonitor{
		logger:           log,
		alertThreshold:   alertThreshold,
		criticalThreshold: criticalThreshold,
		monitorInterval:   monitorInterval,
		outputDir:         outputDir,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start begins monitoring memory usage
func (m *MemoryMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.monitorInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				stats := m.collectStats()
				
				// Log periodic memory stats
				m.logger.Info("Memory usage stats",
					zap.Uint64("allocMB", stats.Alloc/1024/1024),
					zap.Uint64("sysMB", stats.Sys/1024/1024),
					zap.Uint32("numGC", stats.NumGC),
					zap.Float64("gcCPUFraction", stats.GCCPUFraction),
				)
				
				// Check thresholds
				allocMB := stats.Alloc / 1024 / 1024
				
				if allocMB > m.criticalThreshold {
					m.logger.Error("CRITICAL: Memory usage exceeds critical threshold",
						zap.Uint64("allocMB", allocMB),
						zap.Uint64("threshold", m.criticalThreshold),
					)
					m.captureHeapProfile("critical")
					m.triggerGarbageCollection()
				} else if allocMB > m.alertThreshold {
					m.logger.Warn("ALERT: Memory usage exceeds alert threshold",
						zap.Uint64("allocMB", allocMB),
						zap.Uint64("threshold", m.alertThreshold),
					)
					m.captureHeapProfile("alert")
				}
				
			case <-m.ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the memory monitor
func (m *MemoryMonitor) Stop() {
	m.cancel()
}

// collectStats collects memory statistics
func (m *MemoryMonitor) collectStats() MemoryStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	
	return MemoryStats{
		Alloc:        stats.Alloc,
		TotalAlloc:   stats.TotalAlloc,
		Sys:          stats.Sys,
		NumGC:        stats.NumGC,
		PauseTotalNs: stats.PauseTotalNs,
		HeapObjects:  stats.HeapObjects,
		GCCPUFraction: stats.GCCPUFraction,
	}
}

// captureHeapProfile captures a heap profile when memory usage exceeds thresholds
func (m *MemoryMonitor) captureHeapProfile(level string) {
	if m.outputDir == "" {
		return
	}
	
	timestamp := time.Now().Format("20060102-150405")
	filename := m.outputDir + "/heap-" + level + "-" + timestamp + ".pprof"
	
	f, err := os.Create(filename)
	if err != nil {
		m.logger.Error("Failed to create memory profile", zap.Error(err))
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			m.logger.Error("Failed to close heap profile file", zap.Error(err))
		}
	}()
	
	if err := pprof.WriteHeapProfile(f); err != nil {
		m.logger.Error("Failed to write memory profile", zap.Error(err))
	} else {
		m.logger.Info("Memory profile captured", zap.String("filename", filename))
	}
	
	// Also save memory stats as JSON
	statsFilename := m.outputDir + "/memstats-" + level + "-" + timestamp + ".json"
	statsFile, err := os.Create(statsFilename)
	if err != nil {
		m.logger.Error("Failed to create memory stats file", zap.Error(err))
		return
	}
	defer func() {
		if err := statsFile.Close(); err != nil {
			m.logger.Error("Failed to close stats file", zap.Error(err))
		}
	}()
	
	statsData, err := json.MarshalIndent(m.collectStats(), "", "  ")
	if err != nil {
		m.logger.Error("Failed to marshal memory stats", zap.Error(err))
		return
	}
	
	if _, err := statsFile.Write(statsData); err != nil {
		m.logger.Error("Failed to write memory stats", zap.Error(err))
	}
}

// triggerGarbageCollection manually triggers garbage collection
func (m *MemoryMonitor) triggerGarbageCollection() {
	m.logger.Info("Manually triggering garbage collection")
	runtime.GC()
}

// GetMemoryUsage returns the current memory usage in MB
func (m *MemoryMonitor) GetMemoryUsage() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc / 1024 / 1024
}

// DumpFullMemoryStats dumps detailed memory statistics
func (m *MemoryMonitor) DumpFullMemoryStats() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	
	m.logger.Info("Detailed memory statistics",
		zap.Uint64("alloc", stats.Alloc),
		zap.Uint64("totalAlloc", stats.TotalAlloc),
		zap.Uint64("sys", stats.Sys),
		zap.Uint64("lookups", stats.Lookups),
		zap.Uint64("mallocs", stats.Mallocs),
		zap.Uint64("frees", stats.Frees),
		zap.Uint64("heapAlloc", stats.HeapAlloc),
		zap.Uint64("heapSys", stats.HeapSys),
		zap.Uint64("heapIdle", stats.HeapIdle),
		zap.Uint64("heapInuse", stats.HeapInuse),
		zap.Uint64("heapReleased", stats.HeapReleased),
		zap.Uint64("heapObjects", stats.HeapObjects),
		zap.Uint64("stackInuse", stats.StackInuse),
		zap.Uint64("stackSys", stats.StackSys),
		zap.Uint64("mSpanInuse", stats.MSpanInuse),
		zap.Uint64("mSpanSys", stats.MSpanSys),
		zap.Uint64("mCacheInuse", stats.MCacheInuse),
		zap.Uint64("mCacheSys", stats.MCacheSys),
		zap.Uint64("buckHashSys", stats.BuckHashSys),
		zap.Uint64("gCSys", stats.GCSys),
		zap.Uint64("otherSys", stats.OtherSys),
		zap.Uint32("numGC", stats.NumGC),
		zap.Uint64("nextGC", stats.NextGC),
		zap.Uint64("lastGC", stats.LastGC),
		zap.Uint64("pauseTotalNs", stats.PauseTotalNs),
		zap.Float64("gcCPUFraction", stats.GCCPUFraction),
	)
}

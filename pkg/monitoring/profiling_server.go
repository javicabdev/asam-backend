package monitoring

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sort"
	"time"
	
	"go.uber.org/zap"
	"gorm.io/gorm"
	
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// ProfilingServer provides HTTP endpoints for profiling and metrics
type ProfilingServer struct {
	addr             string
	logger           logger.Logger
	gqlTracer        *middleware.GraphQLTracer
	queryMonitor     *QueryMonitor
	memoryMonitor    *MemoryMonitor
	server           *http.Server
	db               *gorm.DB
}

// NewProfilingServer creates a new profiling server
func NewProfilingServer(
	addr string,
	log logger.Logger,
	gqlTracer *middleware.GraphQLTracer,
	queryMonitor *QueryMonitor,
	memoryMonitor *MemoryMonitor,
	db *gorm.DB,
) *ProfilingServer {
	return &ProfilingServer{
		addr:          addr,
		logger:        log,
		gqlTracer:     gqlTracer,
		queryMonitor:  queryMonitor,
		memoryMonitor: memoryMonitor,
		db:            db,
	}
}

// Start starts the profiling server
func (p *ProfilingServer) Start() {
	mux := http.NewServeMux()
	
	// Standard pprof endpoints
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	
	// Custom monitoring endpoints
	mux.HandleFunc("/debug/memory", p.handleMemoryStats)
	mux.HandleFunc("/debug/gql/metrics", p.handleGraphQLMetrics)
	mux.HandleFunc("/debug/db/metrics", p.handleDBMetrics)
	mux.HandleFunc("/debug/db/slow-queries", p.handleSlowQueries)
	
	// System info
	mux.HandleFunc("/debug/sys/info", p.handleSystemInfo)
	
	// Start server
	p.server = &http.Server{
		Addr:    p.addr,
		Handler: mux,
	}
	
	go func() {
		p.logger.Info("Starting profiling server", zap.String("addr", p.addr))
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Error("Profiling server failed", zap.Error(err))
		}
	}()
}

// Stop stops the profiling server
func (p *ProfilingServer) Stop() error {
	if p.server != nil {
		p.logger.Info("Stopping profiling server")
		return p.server.Close()
	}
	return nil
}

// handleMemoryStats handles the memory stats endpoint
func (p *ProfilingServer) handleMemoryStats(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	stats := map[string]interface{}{
		"alloc":         memStats.Alloc,
		"totalAlloc":    memStats.TotalAlloc,
		"sys":           memStats.Sys,
		"lookups":       memStats.Lookups,
		"mallocs":       memStats.Mallocs,
		"frees":         memStats.Frees,
		"heapAlloc":     memStats.HeapAlloc,
		"heapSys":       memStats.HeapSys,
		"heapIdle":      memStats.HeapIdle,
		"heapInuse":     memStats.HeapInuse,
		"heapReleased":  memStats.HeapReleased,
		"heapObjects":   memStats.HeapObjects,
		"numGC":         memStats.NumGC,
		"gcPauseNs":     memStats.PauseNs,
		"gcCPUFraction": memStats.GCCPUFraction,
		"numGoroutines": runtime.NumGoroutine(),
		"numCPU":        runtime.NumCPU(),
		"goVersion":     runtime.Version(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleGraphQLMetrics handles the GraphQL metrics endpoint
func (p *ProfilingServer) handleGraphQLMetrics(w http.ResponseWriter, r *http.Request) {
	if p.gqlTracer == nil {
		http.Error(w, "GraphQL tracer not configured", http.StatusNotFound)
		return
	}
	
	metrics := p.gqlTracer.GetResolverMetrics()
	
	// Return top slowest resolvers
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleDBMetrics handles the database metrics endpoint
func (p *ProfilingServer) handleDBMetrics(w http.ResponseWriter, r *http.Request) {
	if p.db == nil {
		http.Error(w, "Database not configured", http.StatusNotFound)
		return
	}
	
	// Get DB statistics
	sqlDB, err := p.db.DB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	stats := sqlDB.Stats()
	
	dbStats := map[string]interface{}{
		"maxOpenConnections": stats.MaxOpenConnections,
		"openConnections":    stats.OpenConnections,
		"inUse":              stats.InUse,
		"idle":               stats.Idle,
		"waitCount":          stats.WaitCount,
		"waitDuration":       stats.WaitDuration.String(),
		"maxIdleClosed":      stats.MaxIdleClosed,
		"maxLifetimeClosed":  stats.MaxLifetimeClosed,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbStats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleSlowQueries handles the slow queries endpoint
func (p *ProfilingServer) handleSlowQueries(w http.ResponseWriter, r *http.Request) {
	if p.queryMonitor == nil {
		http.Error(w, "Query monitor not configured", http.StatusNotFound)
		return
	}
	
	slowQueries := p.queryMonitor.GetSlowQueriesReport()
	
	// Convert map to slice for sorting
	type slowQuery struct {
		Query     string `json:"query"`
		Frequency int    `json:"frequency"`
	}
	
	var queries []slowQuery
	for query, frequency := range slowQueries {
		queries = append(queries, slowQuery{
			Query:     query,
			Frequency: frequency,
		})
	}
	
	// Sort by frequency (descending)
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Frequency > queries[j].Frequency
	})
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(queries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleSystemInfo handles the system info endpoint
func (p *ProfilingServer) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"goVersion":      runtime.Version(),
		"goOS":           runtime.GOOS,
		"goArch":         runtime.GOARCH,
		"numCPU":         runtime.NumCPU(),
		"numGoroutines":  runtime.NumGoroutine(),
		"currentTime":    time.Now().Format(time.RFC3339),
		"uptime":         time.Since(startTime).String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// startTime tracks application start time
var startTime = time.Now()

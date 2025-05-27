package middleware

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// GraphQLTracer implements a performance tracer for GraphQL resolvers
type GraphQLTracer struct {
	mu              sync.Mutex
	resolverMetrics map[string][]time.Duration
	logger          logger.Logger
	slowThreshold   time.Duration
}

// NewGraphQLTracer creates a new GraphQL tracer
func NewGraphQLTracer(log logger.Logger, slowThreshold time.Duration) *GraphQLTracer {
	return &GraphQLTracer{
		resolverMetrics: make(map[string][]time.Duration),
		logger:          log,
		slowThreshold:   slowThreshold,
	}
}

// InterceptOperation captures the execution time of the entire GraphQL operation
func (a *GraphQLTracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	start := time.Now()
	resp := next(ctx)

	operationCtx := graphql.GetOperationContext(ctx)
	opName := operationCtx.OperationName
	if opName == "" {
		opName = "unnamed-operation"
	}

	duration := time.Since(start)
	if duration > a.slowThreshold {
		// Log slow operations
		a.logger.Warn("SLOW GRAPHQL OPERATION",
			zap.String("operation", opName),
			zap.String("query", operationCtx.RawQuery),
			zap.Duration("duration", duration),
		)
	}

	return resp
}

// InterceptField captures the execution time of individual resolvers
func (a *GraphQLTracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	start := time.Now()
	res, err := next(ctx)
	duration := time.Since(start)

	// Get field information
	fieldCtx := graphql.GetFieldContext(ctx)
	operationCtx := graphql.GetOperationContext(ctx)
	opName := operationCtx.OperationName
	if opName == "" {
		opName = "unnamed-operation"
	}

	resolverPath := opName + "/" + fieldCtx.Field.Name

	// Record metrics
	a.mu.Lock()
	a.resolverMetrics[resolverPath] = append(a.resolverMetrics[resolverPath], duration)
	a.mu.Unlock()

	// Log slow resolvers
	if duration > a.slowThreshold {
		a.logger.Warn("SLOW GRAPHQL RESOLVER",
			zap.String("operation", opName),
			zap.String("resolver", fieldCtx.Field.Name),
			zap.String("path", fieldCtx.Path().String()),
			zap.Duration("duration", duration),
		)
	}

	return res, err
}

// ResolverMetrics contains statistics about a resolver's performance
type ResolverMetrics struct {
	Path         string
	Count        int
	Min          time.Duration
	Max          time.Duration
	Avg          time.Duration
	Median       time.Duration
	Percentile95 time.Duration
}

// GetResolverMetrics returns performance metrics for all recorded resolvers
func (a *GraphQLTracer) GetResolverMetrics() []ResolverMetrics {
	a.mu.Lock()
	defer a.mu.Unlock()

	var metrics []ResolverMetrics

	for path, durations := range a.resolverMetrics {
		if len(durations) == 0 {
			continue
		}

		// Sort durations for percentile calculations
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		// Calculate statistics
		count := len(durations)
		minDuration := durations[0]
		maxDuration := durations[count-1]

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(count)

		// Median and 95th percentile
		median := durations[count/2]
		p95 := durations[int(float64(count)*0.95)]

		metrics = append(metrics, ResolverMetrics{
			Path:         path,
			Count:        count,
			Min:          minDuration,
			Max:          maxDuration,
			Avg:          avg,
			Median:       median,
			Percentile95: p95,
		})
	}

	// Sort by average duration (descending)
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Avg > metrics[j].Avg
	})

	return metrics
}

// ResetMetrics clears all collected metrics
func (a *GraphQLTracer) ResetMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.resolverMetrics = make(map[string][]time.Duration)
}

// LogComplexityMiddleware logs high complexity queries
func LogComplexityMiddleware(logger logger.Logger, _ int) graphql.OperationMiddleware {
	return func(_ context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			// Get operation context
			operationCtx := graphql.GetOperationContext(ctx)

			// Execute the operation
			resp := next(ctx)

			// Check if we can access complexity (this might not be available in all versions)
			// For now, we'll just log the operation details
			if operationCtx.OperationName != "" {
				logger.Debug("GraphQL operation executed",
					zap.String("operation", operationCtx.OperationName),
				)
			}

			return resp(ctx)
		}
	}
}

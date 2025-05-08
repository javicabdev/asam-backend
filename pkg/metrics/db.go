package metrics

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Callback para métricas de GORM
func RegisterMetrics(db *gorm.DB) {
	err := db.Callback().Create().Before("gorm:create").Register("metrics:create", func(db *gorm.DB) {
		startOperation(db)
	})
	if err != nil {
		panic(err)
	}

	err = db.Callback().Create().After("gorm:create").Register("metrics:create:after", func(db *gorm.DB) {
		endOperation(db, "create")
	})
	if err != nil {
		panic(err)
	}

	// Registrar callbacks similares para Update, Delete, Query...
}

func startOperation(db *gorm.DB) {
	ctx := db.Statement.Context
	ctx = context.WithValue(ctx, "metrics:start_time", time.Now())
	db.Statement.Context = ctx
}

func endOperation(db *gorm.DB, operation string) {
	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	if startTime, ok := ctx.Value("metrics:start_time").(time.Time); ok {
		duration := time.Since(startTime).Seconds()
		table := db.Statement.Table

		status := "success"
		if db.Error != nil {
			status = "error"
		}

		DBOperationsTotal.WithLabelValues(operation, table, status).Inc()
		DBOperationDuration.WithLabelValues(operation, table).Observe(duration)
	}
}

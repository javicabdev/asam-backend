package health

import (
	"context"

	"gorm.io/gorm"
)

type DatabaseChecker struct {
	db *gorm.DB
}

func (c *DatabaseChecker) Check(ctx context.Context) ComponentHealth {
	sqlDB, err := c.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "Could not get database instance",
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "Database ping failed",
		}
	}

	return ComponentHealth{Status: StatusUp}
}

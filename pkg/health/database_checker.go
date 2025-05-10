// Package health proporciona funcionalidades para verificar el estado de salud del sistema
package health

import (
	"context"

	"gorm.io/gorm"
)

// DatabaseChecker verifica la conexión a la base de datos y proporciona un mecanismo
// para realizar health checks sobre el sistema de base de datos.
type DatabaseChecker struct {
	db *gorm.DB
}

// Check verifica el estado de la base de datos ejecutando un ping para comprobar
// la conectividad. Devuelve un ComponentHealth con el estado y mensaje correspondiente.
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

// Package health proporciona funcionalidades para verificar el estado de salud del sistema
package health

import (
	"context"
	"runtime"
)

// SystemChecker verifica el estado general del sistema operativo y recursos disponibles
// para la aplicación, como la utilización de memoria.
type SystemChecker struct{}

// Check verifica el estado del sistema monitorizando el uso de memoria y otros recursos.
// Retorna un ComponentHealth con estado DOWN si el uso de memoria supera el 90%.
func (c *SystemChecker) Check(ctx context.Context) ComponentHealth {
	// Verificar si el contexto está cancelado antes de ejecutar la lógica
	select {
	case <-ctx.Done():
		return ComponentHealth{
			Status:  StatusDown,
			Message: "Context canceled or deadline exceeded",
		}
	default:
		// Continuar si el contexto está activo
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Ejemplo de umbral: 90% de uso de memoria
	if float64(m.Alloc)/float64(m.Sys) > 0.9 {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "High memory usage",
		}
	}

	return ComponentHealth{Status: StatusUp}
}

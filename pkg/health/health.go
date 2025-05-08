package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Status representa el estado de un componente
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// ComponentHealth representa el estado de salud de un componente
type ComponentHealth struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Check representa el estado general del sistema
type Check struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// Checker define la interfaz para verificar la salud de un componente
type Checker interface {
	Check(ctx context.Context) ComponentHealth
}

// Handler maneja los health checks del sistema
type Handler struct {
	db       *gorm.DB
	checkers map[string]Checker
	mu       sync.RWMutex
}

func NewHandler(db *gorm.DB) *Handler {
	h := &Handler{
		db:       db,
		checkers: make(map[string]Checker),
	}

	// Registrar checkers por defecto
	h.RegisterChecker("database", &DatabaseChecker{db: db})
	h.RegisterChecker("system", &SystemChecker{})

	return h
}

func (h *Handler) RegisterChecker(name string, checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	health := h.CheckHealth(r.Context())

	w.Header().Set("Content-Type", "application/json")

	if health.Status == StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		// Log error pero no cambia la respuesta HTTP
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":"DOWN","components":{},"error":"Failed to encode health check"}`))
	}
}

func (h *Handler) CheckHealth(ctx context.Context) Check {
	h.mu.RLock()
	defer h.mu.RUnlock()

	health := Check{
		Status:     StatusUp,
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	for name, checker := range h.checkers {
		componentHealth := checker.Check(ctx)
		health.Components[name] = componentHealth

		if componentHealth.Status == StatusDown {
			health.Status = StatusDown
		}
	}

	return health
}

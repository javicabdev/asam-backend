// Package audit proporciona funcionalidad para el registro de eventos de auditoría
package audit

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

// Logger defines the interface for audit logging operations
type Logger interface {
	LogAction(ctx context.Context, action ActionType, entity EntityType, entityID string, description string)
	LogChange(ctx context.Context, action ActionType, entity EntityType, entityID string,
		previous, newData any, description string)
	LogError(ctx context.Context, action ActionType, entity EntityType, entityID string, description string, err error)
}

// ActionType define los tipos de acciones que se pueden auditar
type ActionType string

// EntityType define los tipos de entidades que se pueden auditar
type EntityType string

const (
	// EntityMember representa una entidad de tipo miembro
	EntityMember EntityType = "member"
)

// Metadata representa información adicional para el log de auditoría
type Metadata map[string]any

// Entry representa una entrada en el log de auditoría
type Entry struct {
	Timestamp    time.Time  `json:"timestamp"`
	Action       ActionType `json:"action"`
	Entity       EntityType `json:"entity"`
	EntityID     string     `json:"entity_id"`
	UserID       string     `json:"user_id"`
	Description  string     `json:"description"`
	PreviousData any        `json:"previous_data,omitempty"`
	NewData      any        `json:"new_data,omitempty"`
	Metadata     Metadata   `json:"metadata,omitempty"`
	Status       string     `json:"status"`
}

// loggerImpl is the concrete implementation of the Logger interface
type loggerImpl struct {
	logger logger.Logger
}

// NewLogger creates a new instance of the audit logger
func NewLogger(logger logger.Logger) Logger {
	return &loggerImpl{
		logger: logger,
	}
}

// logAuditEntry registers the audit entry using the main logger
func (a *loggerImpl) logAuditEntry(entry Entry) {
	// Convertir la entrada a JSON para un logging estructurado
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Check if the logger supports zap fields
		if zapLogger, ok := a.logger.(interface {
			Error(msg string, fields ...zap.Field)
		}); ok {
			zapLogger.Error("Failed to marshal audit entry",
				zap.Error(err),
				zap.String("action", string(entry.Action)),
				zap.String("entity", string(entry.Entity)),
			)
		} else {
			// Fallback for non-zap loggers
			a.logger.Error("Failed to marshal audit entry: " + err.Error())
		}
		return
	}

	// If logger supports zap fields, use them
	if zapLogger, ok := a.logger.(interface {
		Info(msg string, fields ...zap.Field)
	}); ok {
		// Crear campos para el log
		fields := []zap.Field{
			zap.String("audit_type", "audit_log"),
			zap.String("action", string(entry.Action)),
			zap.String("entity", string(entry.Entity)),
			zap.String("entity_id", entry.EntityID),
			zap.String("user_id", entry.UserID),
			zap.String("status", entry.Status),
			zap.ByteString("audit_data", jsonData),
		}

		zapLogger.Info(entry.Description, fields...)
	} else {
		// Fallback for non-zap loggers
		a.logger.Info(entry.Description + " - Audit data: " + string(jsonData))
	}
}

// LogAction registra una acción simple en el log de auditoría
func (a *loggerImpl) LogAction(ctx context.Context, action ActionType, entity EntityType,
	entityID string, description string) {
	entry := Entry{
		Timestamp:   time.Now().UTC(),
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		UserID:      getUserFromContext(ctx),
		Description: description,
		Status:      "success",
	}

	a.logAuditEntry(entry)
}

// LogChange registra un cambio en una entidad, incluyendo los datos anteriores y nuevos
func (a *loggerImpl) LogChange(ctx context.Context, action ActionType, entity EntityType,
	entityID string, previous, newData any, description string) {
	entry := Entry{
		Timestamp:    time.Now().UTC(),
		Action:       action,
		Entity:       entity,
		EntityID:     entityID,
		UserID:       getUserFromContext(ctx),
		Description:  description,
		PreviousData: previous,
		NewData:      newData,
		Status:       "success",
	}

	a.logAuditEntry(entry)
}

// LogError registra una acción fallida en el log de auditoría
func (a *loggerImpl) LogError(ctx context.Context, action ActionType, entity EntityType,
	entityID string, description string, err error) {
	entry := Entry{
		Timestamp:   time.Now().UTC(),
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		UserID:      getUserFromContext(ctx),
		Description: description,
		Metadata: Metadata{
			"error": err.Error(),
		},
		Status: "error",
	}

	a.logAuditEntry(entry)
}

// getUserFromContext obtiene el ID del usuario del contexto
func getUserFromContext(_ context.Context) string {
	// Por ahora retornamos un valor por defecto
	// TODO: Implementar cuando tengamos la autenticación
	return "system"
}

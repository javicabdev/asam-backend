package audit

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/pkg/logger"
)

type Logger interface {
	LogAction(ctx context.Context, action ActionType, entity EntityType, entityID string, description string)
	LogChange(ctx context.Context, action ActionType, entity EntityType, entityID string,
		previous, new any, description string)
	LogError(ctx context.Context, action ActionType, entity EntityType, entityID string, description string, err error)
}

// ActionType define los tipos de acciones que se pueden auditar
type ActionType string

const (
	// Acciones de autenticación
	ActionLogin  ActionType = "login"
	ActionLogout ActionType = "logout"

	// Acciones CRUD
	ActionCreate ActionType = "create"
	ActionRead   ActionType = "read"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"

	// Acciones financieras
	ActionPayment       ActionType = "payment"
	ActionRefund        ActionType = "refund"
	ActionBalanceAdjust ActionType = "balance_adjust"
)

// EntityType define los tipos de entidades que se pueden auditar
type EntityType string

const (
	EntityMember   EntityType = "member"
	EntityFamily   EntityType = "family"
	EntityFamiliar EntityType = "familiar"
	EntityPayment  EntityType = "payment"
	EntityCashFlow EntityType = "cash_flow"
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

type AuditLogger struct {
	Logger logger.Logger
}

func NewAudit(logger logger.Logger) Logger {
	return &AuditLogger{
		Logger: logger,
	}
}

// logAuditEntry registra la entrada de auditoría usando el logger principal
func (a *AuditLogger) logAuditEntry(entry Entry) {
	// Convertir la entrada a JSON para un logging estructurado
	jsonData, err := json.Marshal(entry)
	if err != nil {
		a.Logger.Error("Failed to marshal audit entry",
			zap.Error(err),
			zap.String("action", string(entry.Action)),
			zap.String("entity", string(entry.Entity)),
		)
		return
	}

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

	a.Logger.Info(entry.Description, fields...)
}

// LogAction registra una acción simple en el log de auditoría
func (a *AuditLogger) LogAction(ctx context.Context, action ActionType, entity EntityType,
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
func (a *AuditLogger) LogChange(ctx context.Context, action ActionType, entity EntityType,
	entityID string, previous, new any, description string) {
	entry := Entry{
		Timestamp:    time.Now().UTC(),
		Action:       action,
		Entity:       entity,
		EntityID:     entityID,
		UserID:       getUserFromContext(ctx),
		Description:  description,
		PreviousData: previous,
		NewData:      new,
		Status:       "success",
	}

	a.logAuditEntry(entry)
}

// LogError registra una acción fallida en el log de auditoría
func (a *AuditLogger) LogError(ctx context.Context, action ActionType, entity EntityType,
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
// TODO: Implementar cuando tengamos la autenticación
func getUserFromContext(ctx context.Context) string {
	// Por ahora retornamos un valor por defecto
	return "system"
}

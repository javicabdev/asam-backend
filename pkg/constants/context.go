// Package constants define constantes utilizadas en toda la aplicación
package constants

// ContextKey tipo utilizado para las claves en el contexto
type ContextKey string

const (
	// UserContextKey clave para el usuario en el contexto
	UserContextKey ContextKey = "user"
	// UserIDContextKey clave para el ID del usuario en el contexto
	UserIDContextKey ContextKey = "user_id"
	// UserRoleContextKey clave para el rol del usuario en el contexto
	UserRoleContextKey ContextKey = "user_role"
	// AuthorizedContextKey clave para el estado de autorización en el contexto
	AuthorizedContextKey ContextKey = "authorized"
	// IPContextKey clave para la dirección IP en el contexto
	IPContextKey ContextKey = "ip"
	// UserAgentContextKey clave para el agente de usuario en el contexto
	UserAgentContextKey ContextKey = "user_agent"
)

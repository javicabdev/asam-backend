// Package constants define constantes utilizadas en toda la aplicación
package constants

// ContextKey tipo utilizado para las claves en el contexto
type contextKey string

const (
	// UserContextKey clave para el usuario en el contexto
	UserContextKey contextKey = "user"
	// UserIDContextKey clave para el ID del usuario en el contexto
	UserIDContextKey contextKey = "user_id"
	// UserRoleContextKey clave para el rol del usuario en el contexto
	UserRoleContextKey contextKey = "user_role"
	// AuthorizedContextKey clave para el estado de autorización en el contexto
	AuthorizedContextKey contextKey = "authorized"
	// IPContextKey clave para la dirección IP en el contexto
	IPContextKey contextKey = "ip"
	// UserAgentContextKey clave para el agente de usuario en el contexto
	UserAgentContextKey contextKey = "user_agent"
	// DeviceNameContextKey clave para el nombre del dispositivo en el contexto
	DeviceNameContextKey contextKey = "device_name"
	// IPAddressContextKey clave para la dirección IP en el contexto
	IPAddressContextKey contextKey = "ip_address"
	// RequestIDContextKey clave para el ID de la solicitud en el contexto
	RequestIDContextKey contextKey = "request_id"
	// AuthTokenContextKey clave para el token de autorización en el contexto
	AuthTokenContextKey contextKey = "auth_token"
)

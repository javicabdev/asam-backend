package constants

type ContextKey string

const (
	UserContextKey       ContextKey = "user"
	UserIDContextKey     ContextKey = "user_id"
	UserRoleContextKey   ContextKey = "user_role"
	AuthorizedContextKey ContextKey = "authorized"
	IPContextKey         ContextKey = "ip"
	UserAgentContextKey  ContextKey = "user_agent"
)

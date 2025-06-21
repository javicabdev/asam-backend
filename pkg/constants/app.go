package constants

// Environment constants
const (
	// EnvDevelopment represents the development environment
	EnvDevelopment = "development"
	// EnvLocal represents the local environment
	EnvLocal = "local"
	// EnvAiven represents the Aiven environment
	EnvAiven = "aiven"
	// EnvAll represents all environments
	EnvAll = "all"
)

// Health status constants
const (
	// HealthStatusHealthy indicates a healthy service state
	HealthStatusHealthy = "healthy"
	// HealthStatusUnhealthy indicates an unhealthy service state
	HealthStatusUnhealthy = "unhealthy"
	// HealthStatusDegraded indicates a degraded service state
	HealthStatusDegraded = "DEGRADED"
	// HealthStatusUp indicates the service is up
	HealthStatusUp = "UP"
)

// Database SSL mode constants
const (
	// SSLModeDisable disables SSL for database connections
	SSLModeDisable = "disable"
)

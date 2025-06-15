package resolvers

import (
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// convertGraphQLRoleToDomain converts GraphQL UserRole to domain Role
func convertGraphQLRoleToDomain(role model.UserRole) models.Role {
	switch role {
	case model.UserRoleAdmin:
		return models.RoleAdmin
	case model.UserRoleUser:
		return models.RoleUser
	default:
		return models.RoleUser
	}
}

// convertDomainRoleToGraphQL converts domain Role to GraphQL UserRole
func convertDomainRoleToGraphQL(role models.Role) model.UserRole {
	switch role {
	case models.RoleAdmin:
		return model.UserRoleAdmin
	case models.RoleUser:
		return model.UserRoleUser
	default:
		return model.UserRoleUser
	}
}

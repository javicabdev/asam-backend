package resolvers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
)

// parseID convierte un ID de string a uint
func parseID(id string) (uint, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// containsIgnoreCase verifica si s contiene substr sin importar mayúsculas/minúsculas
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// mapMemberFilterToDomain convierte un filtro GraphQL a un filtro del dominio
func (r *queryResolver) mapMemberFilterToDomain(filter *model.MemberFilter) input.MemberFilters {
	// Define default values
	page := 1
	pageSize := 10
	var estado *string
	var tipoMembresia *string
	var searchTerm *string
	var orderBy string

	if filter != nil {
		// Extract pagination
		if filter.Pagination != nil {
			page = filter.Pagination.Page
			pageSize = filter.Pagination.PageSize
		}
		// Map State (ACTIVE / INACTIVE) → (activo / inactivo)
		if filter.Estado != nil {
			tmp := ""
			switch *filter.Estado {
			case model.MemberStatusActive:
				tmp = models.EstadoActivo
			case model.MemberStatusInactive:
				tmp = models.EstadoInactivo
			}
			estado = &tmp
		}
		// Map membership type (INDIVIDUAL / FAMILY) → (individual / familiar)
		if filter.TipoMembresia != nil {
			tmp := ""
			switch *filter.TipoMembresia {
			case model.MembershipTypeIndividual:
				tmp = models.TipoMembresiaPIndividual
			case model.MembershipTypeFamily:
				tmp = models.TipoMembresiaPFamiliar
			}
			tipoMembresia = &tmp
		}
		// Extract search term
		if filter.SearchTerm != nil {
			searchTerm = filter.SearchTerm
		}
		// Extract sort
		if filter.Sort != nil {
			orderBy = fmt.Sprintf("%s %s", filter.Sort.Field, filter.Sort.Direction)
		}
	}

	return input.MemberFilters{
		State:          estado,
		MembershipType: tipoMembresia,
		SearchTerm:     searchTerm,
		Page:           page,
		PageSize:       pageSize,
		OrderBy:        orderBy,
	}
}

// buildMemberConnection construye una respuesta de conexión para miembros
func (r *queryResolver) buildMemberConnection(members []*models.Member, page int) *model.MemberConnection {
	// Convert slice to pointer slice if needed
	memberPtrs := make([]*models.Member, len(members))
	for i, m := range members {
		// Create a copy to avoid pointer issues
		copy := *m
		memberPtrs[i] = &copy
	}

	// Build PageInfo
	pageInfo := &model.PageInfo{
		HasNextPage:     false, // without totalCount, we can't know
		HasPreviousPage: page > 1,
		TotalCount:      len(members), // placeholder
	}

	return &model.MemberConnection{
		Nodes:    memberPtrs,
		PageInfo: pageInfo,
	}
}

// mapPaymentFilterToDomain converts a GraphQL PaymentFilter to domain PaymentFilters
func (r *queryResolver) mapPaymentFilterToDomain(filter *model.PaymentFilter) (input.PaymentFilters, error) {
	// Set default pagination
	page := 1
	pageSize := 10
	var orderBy string

	var filters input.PaymentFilters

	if filter != nil {
		// Pagination
		if filter.Pagination != nil {
			page = filter.Pagination.Page
			pageSize = filter.Pagination.PageSize
		}

		// Sorting
		if filter.Sort != nil {
			orderBy = fmt.Sprintf("%s %s", filter.Sort.Field, filter.Sort.Direction)
		}

		// Status filter
		if filter.Status != nil {
			status := *filter.Status
			filters.Status = &status
		}

		// Payment method filter
		if filter.PaymentMethod != nil {
			filters.PaymentMethod = filter.PaymentMethod
		}

		// Date range filters
		if filter.StartDate != nil {
			filters.StartDate = filter.StartDate
		}
		if filter.EndDate != nil {
			filters.EndDate = filter.EndDate
		}

		// Amount range filters
		if filter.MinAmount != nil {
			filters.MinAmount = filter.MinAmount
		}
		if filter.MaxAmount != nil {
			filters.MaxAmount = filter.MaxAmount
		}

		// Member filter
		if filter.MemberID != nil {
			memberID, err := parseID(*filter.MemberID)
			if err != nil {
				return input.PaymentFilters{}, err
			}
			filters.MemberID = &memberID
		}

		// Family filter
		if filter.FamilyID != nil {
			familyID, err := parseID(*filter.FamilyID)
			if err != nil {
				return input.PaymentFilters{}, err
			}
			filters.FamilyID = &familyID
		}
	}

	filters.Page = page
	filters.PageSize = pageSize
	filters.OrderBy = orderBy

	return filters, nil
}

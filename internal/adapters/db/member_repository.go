package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type memberRepository struct {
	db *gorm.DB
}

// NewMemberRepository crea una nueva instancia del repositorio
func NewMemberRepository(db *gorm.DB) output.MemberRepository {
	return &memberRepository{db: db}
}

// Create crea un nuevo miembro en la base de datos
func (r *memberRepository) Create(ctx context.Context, member *models.Member) error {
	result := r.db.WithContext(ctx).Create(member)
	if result.Error != nil {
		// Check for specific database errors
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "member with the same key already exists")
		}
		return appErrors.DB(result.Error, "error creating member")
	}
	return nil
}

// GetByID busca un miembro por su ID
func (r *memberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).First(&member, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Consistently return nil, nil for not found
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by ID")
	}

	return &member, nil
}

// GetByNumeroSocio busca un miembro por su número de socio
func (r *memberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).Where("membership_number = ?", numeroSocio).First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Consistently return nil, nil for not found
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by numero socio")
	}

	return &member, nil
}

// GetByIdentityCard busca un miembro por su documento de identidad
func (r *memberRepository) GetByIdentityCard(ctx context.Context, identityCard string) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).Where("identity_card = ?", identityCard).First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Consistently return nil, nil for not found
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by identity card")
	}

	return &member, nil
}

// Update actualiza un miembro existente
func (r *memberRepository) Update(ctx context.Context, member *models.Member) error {
	result := r.db.WithContext(ctx).Save(member)
	if result.Error != nil {
		// Check for specific database errors
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("member", result.Error)
		}
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "member with the same key already exists")
		}
		return appErrors.DB(result.Error, "error updating member")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("member", nil)
	}

	return nil
}

// Delete elimina un miembro por su ID
func (r *memberRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Member{}, id)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error deleting member")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("member", nil)
	}

	return nil
}

// List obtiene una lista de miembros según los filtros proporcionados
func (r *memberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	var members []models.Member
	query := r.db.WithContext(ctx)

	// Aplicar filtros
	if filters.Estado != nil {
		query = query.Where("state = ?", *filters.Estado)
	}

	if filters.TipoMembresia != nil {
		query = query.Where("membership_type = ?", *filters.TipoMembresia)
	}

	if filters.SearchTerm != nil {
		searchTerm := "%" + *filters.SearchTerm + "%"
		query = query.Where(
			"membership_number ILIKE ? OR name ILIKE ? OR surnames ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Aplicar ordenamiento
	if filters.OrderBy != "" {
		query = query.Order(filters.OrderBy)
	}

	// Aplicar paginación
	if filters.Page > 0 && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	result := query.Find(&members)
	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error listing members")
	}

	return members, nil
}

// Transaction support methods

// BeginTransaction starts a new database transaction
func (r *memberRepository) BeginTransaction(ctx context.Context) (output.Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, appErrors.DB(tx.Error, "error beginning transaction")
	}
	return &gormTransaction{tx: tx}, nil
}

// CreateWithTx creates a member within a transaction
func (r *memberRepository) CreateWithTx(ctx context.Context, tx output.Transaction, member *models.Member) error {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	result := gormTx.tx.WithContext(ctx).Create(member)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "member with the same key already exists")
		}
		return appErrors.DB(result.Error, "error creating member")
	}
	return nil
}

// GetByIDWithTx gets a member by ID within a transaction
func (r *memberRepository) GetByIDWithTx(ctx context.Context, tx output.Transaction, id uint) (*models.Member, error) {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return nil, appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	var member models.Member
	result := gormTx.tx.WithContext(ctx).First(&member, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by ID")
	}

	return &member, nil
}

// GetByNumeroSocioWithTx gets a member by numero socio within a transaction
func (r *memberRepository) GetByNumeroSocioWithTx(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error) {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return nil, appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	var member models.Member
	result := gormTx.tx.WithContext(ctx).Where("membership_number = ?", numeroSocio).First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by numero socio")
	}

	return &member, nil
}

// GetByIdentityCardWithTx gets a member by identity card within a transaction
func (r *memberRepository) GetByIdentityCardWithTx(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error) {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return nil, appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	var member models.Member
	result := gormTx.tx.WithContext(ctx).Where("identity_card = ?", identityCard).First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by identity card")
	}

	return &member, nil
}

// GetLastMemberNumberByPrefix obtiene el último número de socio con un prefijo específico
func (r *memberRepository) GetLastMemberNumberByPrefix(ctx context.Context, prefix string) (string, error) {
	var lastNumber string

	// Query que busca el último número con el prefijo dado y ordena por la parte numérica
	// PostgreSQL query: SELECT membership_number FROM members WHERE membership_number LIKE 'A%'
	// ORDER BY CAST(SUBSTRING(membership_number FROM 2) AS INTEGER) DESC LIMIT 1
	result := r.db.WithContext(ctx).
		Model(&models.Member{}).
		Where("membership_number LIKE ?", prefix+"%").
		Select("membership_number").
		Order("CAST(SUBSTRING(membership_number FROM 2) AS INTEGER) DESC").
		Limit(1).
		Scan(&lastNumber)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Si no hay registros, devolver string vacío
			return "", nil
		}
		return "", appErrors.DB(result.Error, "error getting last member number")
	}

	// Si no se encontró ningún registro
	if lastNumber == "" {
		return "", nil
	}

	return lastNumber, nil
}

// SearchWithoutUser busca miembros que coincidan con el criterio y no tengan usuario asociado
func (r *memberRepository) SearchWithoutUser(ctx context.Context, criteria string) ([]models.Member, error) {
	var members []models.Member

	// Búsqueda por nombre, apellidos o número de socio
	// Excluir miembros que ya tienen usuario asociado
	searchPattern := "%" + criteria + "%"

	query := r.db.WithContext(ctx).
		Table("members").
		Select("members.*").
		Joins("LEFT JOIN users ON users.member_id = members.id").
		Where("users.id IS NULL").                       // Excluir miembros con usuario
		Where("members.state = ?", models.EstadoActivo). // Solo miembros activos
		Where(
			"members.name ILIKE ? OR members.surnames ILIKE ? OR members.membership_number ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)

	result := query.Find(&members)
	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error searching members without user")
	}

	return members, nil
}

package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// gormTransaction wraps *gorm.DB to implement output.Transaction interface
type gormTransaction struct {
	tx *gorm.DB
}

func (t *gormTransaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *gormTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

type familyRepository struct {
	db *gorm.DB
}

// NewFamilyRepository creates a new instance of the repository
func NewFamilyRepository(db *gorm.DB) output.FamilyRepository {
	return &familyRepository{
		db: db,
	}
}

// Create inserts a new family into the database
func (r *familyRepository) Create(ctx context.Context, family *models.Family) error {
	result := r.db.WithContext(ctx).Create(family)
	if result.Error != nil {
		// Check for specific database errors
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "family with the same number already exists")
		}
		return appErrors.DB(result.Error, "error creating family")
	}
	return nil
}

// GetByID gets a family by its ID
func (r *familyRepository) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	var family models.Family
	result := r.db.WithContext(ctx).
		Preload("MiembroOrigen").
		Preload("Familiares").
		Preload("Telefonos").
		First(&family, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "error getting family by ID")
	}
	return &family, nil
}

// GetByNumeroSocio gets a family by its member number
func (r *familyRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	var family models.Family
	result := r.db.WithContext(ctx).
		Preload("MiembroOrigen").
		Preload("Familiares").
		Preload("Telefonos").
		Where("numero_socio = ?", numeroSocio).
		First(&family)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "error getting family by numero socio")
	}
	return &family, nil
}

// GetByOriginMemberID gets a family by the origin member ID
func (r *familyRepository) GetByOriginMemberID(ctx context.Context, memberID uint) (*models.Family, error) {
	var family models.Family
	result := r.db.WithContext(ctx).
		Preload("MiembroOrigen").
		Preload("Familiares").
		Preload("Telefonos").
		Where("miembro_origen_id = ?", memberID).
		First(&family)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Consistent pattern: nil, nil for "not found"
		}
		return nil, appErrors.DB(result.Error, "error getting family by origin member ID")
	}
	return &family, nil
}

// Update updates an existing family's data
func (r *familyRepository) Update(ctx context.Context, family *models.Family) error {
	result := r.db.WithContext(ctx).Save(family)
	if result.Error != nil {
		// Check for specific database errors
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("family", result.Error)
		}
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "family with the same number already exists")
		}
		if IsConstraintViolationError(result.Error) {
			return appErrors.New(appErrors.ErrInvalidOperation, "cannot update family due to constraint violations")
		}
		return appErrors.DB(result.Error, "error updating family")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("family", nil)
	}

	return nil
}

// Delete removes a family (soft delete)
func (r *familyRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Family{}, id)
	if result.Error != nil {
		if IsConstraintViolationError(result.Error) {
			return appErrors.New(appErrors.ErrInvalidOperation, "cannot delete family due to dependent records")
		}
		return appErrors.DB(result.Error, "error deleting family")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("family", nil)
	}

	return nil
}

// List gets a paginated list of families
func (r *familyRepository) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	var families []*models.Family
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Family{})

	// Apply search if provided
	if searchTerm != nil && *searchTerm != "" {
		searchQuery := "%" + *searchTerm + "%"
		query = query.Where(
			"numero_socio ILIKE ? OR esposo_nombre ILIKE ? OR esposo_apellidos ILIKE ? OR esposa_nombre ILIKE ? OR esposa_apellidos ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
		)
	}

	// Get total record count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, appErrors.DB(err, "error counting families")
	}

	// Apply sorting if provided
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// Apply pagination
	query = query.Offset((page - 1) * pageSize).Limit(pageSize)

	// Load relationships
	query = query.Preload("MiembroOrigen").
		Preload("Familiares").
		Preload("Telefonos")

	// Execute the query
	result := query.Find(&families)
	if result.Error != nil {
		return nil, 0, appErrors.DB(result.Error, "error listing families")
	}

	return families, int(total), nil
}

// Operations for family members

// AddFamiliar adds a new family member to a family
func (r *familyRepository) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	familiar.FamiliaID = familyID
	result := r.db.WithContext(ctx).Create(familiar)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "duplicate family member")
		}
		return appErrors.DB(result.Error, "error adding familiar")
	}
	return nil
}

// UpdateFamiliar updates a family member's data
func (r *familyRepository) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	result := r.db.WithContext(ctx).Save(familiar)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("familiar", result.Error)
		}
		return appErrors.DB(result.Error, "error updating familiar")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("familiar", nil)
	}

	return nil
}

// RemoveFamiliar removes a family member
func (r *familyRepository) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Familiar{}, familiarID)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error removing familiar")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("familiar", nil)
	}

	return nil
}

// GetFamiliares gets all family members of a family
func (r *familyRepository) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	var familiares []*models.Familiar
	result := r.db.WithContext(ctx).
		Where("familia_id = ?", familyID).
		Find(&familiares)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error getting familiares")
	}
	return familiares, nil
}

// Transaction support methods

// BeginTransaction starts a new database transaction
func (r *familyRepository) BeginTransaction(ctx context.Context) (output.Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, appErrors.DB(tx.Error, "error beginning transaction")
	}
	return &gormTransaction{tx: tx}, nil
}

// CreateWithTx creates a family within a transaction
func (r *familyRepository) CreateWithTx(ctx context.Context, tx output.Transaction, family *models.Family) error {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	result := gormTx.tx.WithContext(ctx).Create(family)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "family with the same number already exists")
		}
		return appErrors.DB(result.Error, "error creating family")
	}
	return nil
}

// GetByIDWithTx gets a family by ID within a transaction
func (r *familyRepository) GetByIDWithTx(ctx context.Context, tx output.Transaction, id uint) (*models.Family, error) {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return nil, appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	var family models.Family
	result := gormTx.tx.WithContext(ctx).
		Preload("MiembroOrigen").
		Preload("Familiares").
		Preload("Telefonos").
		First(&family, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting family by ID")
	}
	return &family, nil
}

// AddFamiliarWithTx adds a familiar within a transaction
func (r *familyRepository) AddFamiliarWithTx(ctx context.Context, tx output.Transaction, familyID uint, familiar *models.Familiar) error {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	familiar.FamiliaID = familyID
	result := gormTx.tx.WithContext(ctx).Create(familiar)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "duplicate family member")
		}
		return appErrors.DB(result.Error, "error adding familiar")
	}
	return nil
}

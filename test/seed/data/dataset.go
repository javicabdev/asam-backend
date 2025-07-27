package data

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Seedable is an interface for objects that can seed the database
type Seedable interface {
	Seed(ctx context.Context) error
}

// DatasetType represents the type of dataset to seed
type DatasetType string

const (
	// MinimalType represents a minimal dataset for quick testing
	MinimalType DatasetType = "minimal"

	// FullType represents a full dataset for comprehensive testing
	FullType DatasetType = "full"

	// ScenarioType represents a specific scenario dataset
	ScenarioType DatasetType = "scenario"

	// CustomType represents a custom dataset with specified counts
	CustomType DatasetType = "custom"
)

// Seeder interface for objects that can seed individual entity types
type Seeder interface {
	Clean(ctx context.Context) error
	SeedMiembros(ctx context.Context) error
	SeedFamilias(ctx context.Context) error
	SeedFamiliares(ctx context.Context) error
	SeedCuotasMembresia(ctx context.Context) error
	SeedCaja(ctx context.Context) error
	Logf(format string, args ...any)
}

// dataset represents a collection of seed data
type dataset struct {
	db          *sqlx.DB
	seeder      Seeder
	datasetType DatasetType
	args        []any
}

// Dataset returns a new dataset with the specified type and optional arguments
func Dataset(db *sqlx.DB, seeder Seeder, datasetType DatasetType, args ...any) Seedable {
	return &dataset{
		db:          db,
		seeder:      seeder,
		datasetType: datasetType,
		args:        args,
	}
}

// Seed seeds the database with the dataset
func (d *dataset) Seed(ctx context.Context) error {
	switch d.datasetType {
	case MinimalType:
		return NewMinimalDataset(d.db, d.seeder).Seed(ctx)
	case FullType:
		return NewFullDataset(d.db, d.seeder).Seed(ctx)
	case ScenarioType:
		if len(d.args) > 0 {
			if scenario, ok := d.args[0].(string); ok {
				return NewScenarioDataset(d.db, d.seeder, scenario).Seed(ctx)
			}
		}
		return NewScenarioDataset(d.db, d.seeder, "default").Seed(ctx)
	case CustomType:
		// The custom dataset is handled directly in the main.go seedCustom function
		// This is just a placeholder to match the interface
		return nil
	default:
		// Default to minimal dataset
		return NewMinimalDataset(d.db, d.seeder).Seed(ctx)
	}
}

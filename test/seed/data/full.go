package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// FullDataset represents a full dataset for comprehensive testing
type FullDataset struct {
	db     *sqlx.DB
	seeder Seeder
}

// NewFullDataset creates a new full dataset
func NewFullDataset(db *sqlx.DB, seeder Seeder) *FullDataset {
	return &FullDataset{
		db:     db,
		seeder: seeder,
	}
}

// Seed seeds the database with the full dataset
func (f *FullDataset) Seed(ctx context.Context) error {
	log.Println("Seeding full dataset")

	// Clean the database first
	if err := f.seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Seed a larger dataset for comprehensive testing

	// Seed 50 members
	if err := f.seeder.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("failed to seed members: %w", err)
	}

	// Seed 20 families
	if err := f.seeder.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("failed to seed families: %w", err)
	}

	// Seed 40 family members
	if err := f.seeder.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("failed to seed family members: %w", err)
	}

	// Seed 100 payments
	if err := f.seeder.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("failed to seed payments: %w", err)
	}

	// Seed 200 cashflows
	if err := f.seeder.SeedCaja(ctx); err != nil {
		return fmt.Errorf("failed to seed cashflows: %w", err)
	}

	log.Println("Full dataset seeding completed")
	return nil
}

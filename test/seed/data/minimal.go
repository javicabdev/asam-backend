package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// MinimalDataset represents a minimal dataset for quick testing
type MinimalDataset struct {
	db     *sqlx.DB
	seeder Seeder
}

// NewMinimalDataset creates a new minimal dataset
func NewMinimalDataset(db *sqlx.DB, seeder Seeder) *MinimalDataset {
	return &MinimalDataset{
		db:     db,
		seeder: seeder,
	}
}

// Seed seeds the database with the minimal dataset
func (m *MinimalDataset) Seed(ctx context.Context) error {
	log.Println("Seeding minimal dataset")

	// Clean the database first
	if err := m.seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Seed minimal number of records
	// Just enough to test basic functionality

	// Seed 10 members
	if err := m.seeder.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("failed to seed members: %w", err)
	}

	// Seed 5 families
	if err := m.seeder.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("failed to seed families: %w", err)
	}

	// Seed 10 family members
	if err := m.seeder.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("failed to seed family members: %w", err)
	}

	// Seed 20 payments
	if err := m.seeder.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("failed to seed payments: %w", err)
	}

	// Seed 30 cashflows
	if err := m.seeder.SeedCaja(ctx); err != nil {
		return fmt.Errorf("failed to seed cashflows: %w", err)
	}

	log.Println("Minimal dataset seeding completed")
	return nil
}

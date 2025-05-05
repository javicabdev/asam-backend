package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// MinimalDataset represents a minimal dataset for quick testing
type MinimalDataset struct {
	BaseDataset
}

// NewMinimalDataset creates a new minimal dataset
func NewMinimalDataset(db *sqlx.DB, seeder Seeder) *MinimalDataset {
	return &MinimalDataset{
		BaseDataset: BaseDataset{
			DB:     db,
			Seeder: seeder,
		},
	}
}

// Seed seeds the database with the minimal dataset
func (m *MinimalDataset) Seed(ctx context.Context) error {
	if err := m.SeedDatabase(ctx, "minimal"); err != nil {
		return err
	}

	// Seed minimal number of records
	// Just enough to test basic functionality
	if err := m.SeedAll(ctx, "failed to seed"); err != nil {
		return err
	}

	log.Println("Minimal dataset seeding completed")
	return nil
}

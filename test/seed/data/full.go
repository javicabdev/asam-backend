package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// FullDataset represents a full dataset for comprehensive testing
type FullDataset struct {
	BaseDataset
}

// NewFullDataset creates a new full dataset
func NewFullDataset(db *sqlx.DB, seeder Seeder) *FullDataset {
	return &FullDataset{
		BaseDataset: BaseDataset{
			DB:     db,
			Seeder: seeder,
		},
	}
}

// Seed seeds the database with the full dataset
func (f *FullDataset) Seed(ctx context.Context) error {
	if err := f.SeedDatabase(ctx, "full"); err != nil {
		return err
	}

	// Seed a larger dataset for comprehensive testing
	if err := f.SeedAll(ctx, "failed to seed"); err != nil {
		return err
	}

	log.Println("Full dataset seeding completed")
	return nil
}

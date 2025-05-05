package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// BaseDataset proporciona funcionalidad común para los diferentes conjuntos de datos
type BaseDataset struct {
	DB     *sqlx.DB
	Seeder Seeder
}

// CleanDatabase limpia la base de datos antes de sembrar nuevos datos
func (b *BaseDataset) CleanDatabase(ctx context.Context) error {
	if err := b.Seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}
	return nil
}

// SeedDatabase ejecuta las operaciones comunes de siembra
func (b *BaseDataset) SeedDatabase(ctx context.Context, datasetName string) error {
	log.Printf("Seeding %s dataset", datasetName)

	// Clean the database first
	if err := b.CleanDatabase(ctx); err != nil {
		return err
	}

	return nil
}

// SeedAll siembra todos los tipos de datos en la base de datos
func (b *BaseDataset) SeedAll(ctx context.Context, errorPrefix string) error {
	// Seed members
	if err := b.Seeder.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("%s members: %w", errorPrefix, err)
	}

	// Seed families
	if err := b.Seeder.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("%s families: %w", errorPrefix, err)
	}

	// Seed family members
	if err := b.Seeder.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("%s family members: %w", errorPrefix, err)
	}

	// Seed membership payments
	if err := b.Seeder.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("%s payments: %w", errorPrefix, err)
	}

	// Seed cashflows
	if err := b.Seeder.SeedCaja(ctx); err != nil {
		return fmt.Errorf("%s cashflows: %w", errorPrefix, err)
	}

	return nil
}

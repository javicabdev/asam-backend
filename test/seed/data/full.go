package data

import (
	"context"
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/test/seed/generators"
	"github.com/jmoiron/sqlx"
)

// FullDataset represents a complete dataset for comprehensive testing
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

// Seed populates the database with a complete dataset
func (d *FullDataset) Seed(ctx context.Context) error {
	d.seeder.Logf("Seeding full dataset")

	// Clean the database first
	if err := d.seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Seed members (50 individual, 20 family)
	if err := d.seedMembers(ctx); err != nil {
		return fmt.Errorf("failed to seed members: %w", err)
	}

	// Seed families (20)
	if err := d.seedFamilies(ctx); err != nil {
		return fmt.Errorf("failed to seed families: %w", err)
	}

	// Seed family members (60)
	if err := d.seedFamilyMembers(ctx); err != nil {
		return fmt.Errorf("failed to seed family members: %w", err)
	}

	// Seed payments (150)
	if err := d.seedPayments(ctx); err != nil {
		return fmt.Errorf("failed to seed payments: %w", err)
	}

	// Seed cash movements (200)
	if err := d.seedCashflows(ctx); err != nil {
		return fmt.Errorf("failed to seed cash movements: %w", err)
	}

	d.seeder.Logf("Full dataset seeding completed successfully")
	return nil
}

// seedMembers seeds a comprehensive set of members
func (d *FullDataset) seedMembers(ctx context.Context) error {
	d.seeder.Logf("Seeding members for full dataset")

	// Generate 70 members (50 individual, 20 family)
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Setup seed for reproducibility
	seed := time.Now().UnixNano()

	// Individual members (50)
	individualMemberGen := generators.NewMemberGenerator(d.db, seed)
	if err := individualMemberGen.Generate(ctx, 50); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate individual members: %w", err)
	}

	// Family members (20)
	familyMemberGen := generators.NewMemberGenerator(d.db, seed+1)
	if err := familyMemberGen.Generate(ctx, 20); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate family members: %w", err)
	}

	return tx.Commit()
}

// seedFamilies seeds a comprehensive set of families
func (d *FullDataset) seedFamilies(ctx context.Context) error {
	d.seeder.Logf("Seeding families for full dataset")

	// Generate 20 families
	familyGen := generators.NewFamilyGenerator(d.db, time.Now().UnixNano())
	if err := familyGen.Generate(ctx, 20); err != nil {
		return fmt.Errorf("failed to generate families: %w", err)
	}

	return nil
}

// seedFamilyMembers seeds a comprehensive set of family members
func (d *FullDataset) seedFamilyMembers(ctx context.Context) error {
	d.seeder.Logf("Seeding family members for full dataset")

	// Generate 60 family members
	familyMemberGen := generators.NewFamilyMemberGenerator(d.db, time.Now().UnixNano())
	if err := familyMemberGen.Generate(ctx, 60); err != nil {
		return fmt.Errorf("failed to generate family members: %w", err)
	}

	return nil
}

// seedPayments seeds a comprehensive set of membership payments
func (d *FullDataset) seedPayments(ctx context.Context) error {
	d.seeder.Logf("Seeding payments for full dataset")

	// Generate 150 payments
	paymentGen := generators.NewPaymentGenerator(d.db, time.Now().UnixNano())
	if err := paymentGen.Generate(ctx, 150); err != nil {
		return fmt.Errorf("failed to generate payments: %w", err)
	}

	return nil
}

// seedCashflows seeds a comprehensive set of cash movements
func (d *FullDataset) seedCashflows(ctx context.Context) error {
	d.seeder.Logf("Seeding cash movements for full dataset")

	// Generate 200 cash movements
	cashflowGen := generators.NewCashflowGenerator(d.db, time.Now().UnixNano())
	if err := cashflowGen.Generate(ctx, 200); err != nil {
		return fmt.Errorf("failed to generate cash movements: %w", err)
	}

	return nil
}

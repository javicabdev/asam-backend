package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ScenarioDataset represents a dataset for specific testing scenarios
type ScenarioDataset struct {
	db     *sqlx.DB
	seeder Seeder
	name   string
}

// NewScenarioDataset creates a new scenario dataset
func NewScenarioDataset(db *sqlx.DB, seeder Seeder, name string) Seedable {
	return &ScenarioDataset{
		db:     db,
		seeder: seeder,
		name:   name,
	}
}

// Seed populates the database with a scenario-specific dataset
func (d *ScenarioDataset) Seed(ctx context.Context) error {
	d.seeder.Logf("Seeding scenario dataset: %s", d.name)

	// Clean the database first
	if err := d.seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Select and run the specific scenario
	switch d.name {
	case "payment_overdue":
		return d.seedPaymentOverdueScenario(ctx)
	case "membership_expired":
		return d.seedMembershipExpiredScenario(ctx)
	case "large_family":
		return d.seedLargeFamilyScenario(ctx)
	case "financial_emergency":
		return d.seedFinancialEmergencyScenario(ctx)
	default:
		return fmt.Errorf("unknown scenario: %s", d.name)
	}
}

// seedPaymentOverdueScenario creates a dataset with overdue payments
func (d *ScenarioDataset) seedPaymentOverdueScenario(ctx context.Context) error {
	d.seeder.Logf("Seeding payment overdue scenario")

	// Generate a base dataset
	minimal := NewMinimalDataset(d.db, d.seeder)
	if err := minimal.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed base dataset: %w", err)
	}

	// Add members with missing payments
	currentYear := time.Now().Year()
	lastYear := currentYear - 1

	// Get 5 active members
	type Member struct {
		ID            int    `db:"miembro_id"`
		NumeroSocio   string `db:"numero_socio"`
		TipoMembresia string `db:"tipo_membresia"`
	}

	var members []Member
	err := d.db.SelectContext(ctx, &members,
		"SELECT miembro_id, numero_socio, tipo_membresia FROM miembros WHERE estado = 'activo' LIMIT 5")
	if err != nil {
		return fmt.Errorf("failed to get members: %w", err)
	}

	// Add missing payments for half the members
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Delete current year payments for these members
	for i, member := range members {
		if i < 3 { // Only for first 3 members
			_, err := tx.ExecContext(ctx,
				"DELETE FROM cuotas_membresia WHERE miembro_id = $1 AND ano = $2",
				member.ID, currentYear)
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					d.seeder.Logf("Error during rollback: %v", rbErr)
				}
				return fmt.Errorf("failed to delete payments: %w", err)
			}

			// For one member, also delete last year's payment
			if i == 0 {
				_, err := tx.ExecContext(ctx,
					"DELETE FROM cuotas_membresia WHERE miembro_id = $1 AND ano = $2",
					member.ID, lastYear)
				if err != nil {
					if rbErr := tx.Rollback(); rbErr != nil {
						d.seeder.Logf("Error during rollback: %v", rbErr)
					}
					return fmt.Errorf("failed to delete payments: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// seedMembershipExpiredScenario creates a dataset with expired memberships
func (d *ScenarioDataset) seedMembershipExpiredScenario(ctx context.Context) error {
	d.seeder.Logf("Seeding membership expired scenario")

	// Generate a base dataset
	minimal := NewMinimalDataset(d.db, d.seeder)
	if err := minimal.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed base dataset: %w", err)
	}

	// Set some members as inactive with expiration dates
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Get 5 active members
	type Member struct {
		ID          int    `db:"miembro_id"`
		NumeroSocio string `db:"numero_socio"`
	}

	var members []Member
	err = tx.SelectContext(ctx, &members,
		"SELECT miembro_id, numero_socio FROM miembros WHERE estado = 'activo' LIMIT 5")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to get members: %w", err)
	}

	// Set 3 members as inactive with different expiration dates
	for i, member := range members {
		if i < 3 { // Only for first 3 members
			fechaBaja := time.Now().AddDate(0, -(i + 1), 0) // 1, 2, 3 months ago

			_, err := tx.ExecContext(ctx,
				"UPDATE miembros SET estado = 'inactivo', fecha_baja = $1 WHERE miembro_id = $2",
				fechaBaja, member.ID)
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					d.seeder.Logf("Error during rollback: %v", rbErr)
				}
				return fmt.Errorf("failed to update member status: %w", err)
			}
		}
	}

	return tx.Commit()
}

// seedLargeFamilyScenario creates a dataset with large families
func (d *ScenarioDataset) seedLargeFamilyScenario(ctx context.Context) error {
	d.seeder.Logf("Seeding large family scenario")

	// Generate a base dataset
	minimal := NewMinimalDataset(d.db, d.seeder)
	if err := minimal.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed base dataset: %w", err)
	}

	// Add more children to existing families
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Get families
	type Family struct {
		ID          int    `db:"familia_id"`
		NumeroSocio string `db:"numero_socio"`
	}

	var families []Family
	err = tx.SelectContext(ctx, &families, "SELECT familia_id, numero_socio FROM familias LIMIT 3")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to get families: %w", err)
	}

	// Current date and ages for children
	now := time.Now()

	// Add 5 children to the first family
	for i := 0; i < 5; i++ {
		childAge := 5 + i // 5, 6, 7, 8, 9 years old
		birthYear := now.Year() - childAge
		birthMonth := time.Month(1 + i%12)
		birthDay := 1 + i%28
		birthDate := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)

		childName := fmt.Sprintf("Child%d", i+1)

		_, err := tx.ExecContext(ctx,
			`INSERT INTO familiares (familia_id, nombre, fecha_nacimiento)
			VALUES ($1, $2, $3)`,
			families[0].ID, childName, birthDate)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				d.seeder.Logf("Error during rollback: %v", rbErr)
			}
			return fmt.Errorf("failed to insert family member: %w", err)
		}
	}

	// Add 4 children to the second family
	for i := 0; i < 4; i++ {
		childAge := 7 + i*3 // 7, 10, 13, 16 years old
		birthYear := now.Year() - childAge
		birthMonth := time.Month(1 + i%12)
		birthDay := 1 + i%28
		birthDate := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)

		childName := fmt.Sprintf("Child%d", i+1)

		_, err := tx.ExecContext(ctx,
			`INSERT INTO familiares (familia_id, nombre, fecha_nacimiento)
			VALUES ($1, $2, $3)`,
			families[1].ID, childName, birthDate)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				d.seeder.Logf("Error during rollback: %v", rbErr)
			}
			return fmt.Errorf("failed to insert family member: %w", err)
		}
	}

	return tx.Commit()
}

// seedFinancialEmergencyScenario creates a dataset with financial emergencies
func (d *ScenarioDataset) seedFinancialEmergencyScenario(ctx context.Context) error {
	d.seeder.Logf("Seeding financial emergency scenario")

	// Generate a base dataset
	minimal := NewMinimalDataset(d.db, d.seeder)
	if err := minimal.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed base dataset: %w", err)
	}

	// Add emergency cash movements
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Get some member IDs
	type Member struct {
		ID int `db:"miembro_id"`
	}

	var members []Member
	err = tx.SelectContext(ctx, &members,
		"SELECT miembro_id FROM miembros WHERE estado = 'activo' LIMIT 3")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to get members: %w", err)
	}

	// Get some family IDs
	type Family struct {
		ID int `db:"familia_id"`
	}

	var families []Family
	err = tx.SelectContext(ctx, &families, "SELECT familia_id FROM familias LIMIT 2")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to get families: %w", err)
	}

	// Add emergency funds for a member
	_, err = tx.ExecContext(ctx,
		`INSERT INTO caja (miembro_id, tipo_operacion, monto, fecha, detalle)
		VALUES ($1, 'entrega_fondo', -1000, $2, 'Emergencia médica')`,
		members[0].ID, time.Now().AddDate(0, -1, -5))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to insert cash movement: %w", err)
	}

	// Add emergency funds for a family
	_, err = tx.ExecContext(ctx,
		`INSERT INTO caja (familia_id, tipo_operacion, monto, fecha, detalle)
		VALUES ($1, 'entrega_fondo', -1500, $2, 'Repatriación urgente')`,
		families[0].ID, time.Now().AddDate(0, -2, -10))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to insert cash movement: %w", err)
	}

	// Add a large donation
	_, err = tx.ExecContext(ctx,
		`INSERT INTO caja (tipo_operacion, monto, fecha, detalle)
		VALUES ('otros_ingresos', 3000, $1, 'Donación extraordinaria para emergencias')`,
		time.Now().AddDate(0, -3, 0))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.seeder.Logf("Error during rollback: %v", rbErr)
		}
		return fmt.Errorf("failed to insert cash movement: %w", err)
	}

	return tx.Commit()
}

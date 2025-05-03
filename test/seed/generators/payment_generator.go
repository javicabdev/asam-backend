package generators

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

// PaymentGenerator generates test payments (cuotas_membresia)
type PaymentGenerator struct {
	db   *sqlx.DB
	rand *rand.Rand
}

// Payment represents a cuota_membresia record for generation
type Payment struct {
	CuotaID        int       `db:"cuota_id"`
	MiembroID      int       `db:"miembro_id"`
	Ano            int       `db:"ano"`
	CantidadPagada float64   `db:"cantidad_pagada"`
	FechaPago      time.Time `db:"fecha_pago"`
}

// NewPaymentGenerator creates a new payment generator
func NewPaymentGenerator(db *sqlx.DB, seed int64) *PaymentGenerator {
	return &PaymentGenerator{
		db:   db,
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Generate creates n random payments
func (g *PaymentGenerator) Generate(ctx context.Context, n int) error {
	// Get members to associate with
	var members []struct {
		MiembroID     int       `db:"miembro_id"`
		FechaAlta     time.Time `db:"fecha_alta"`
		TipoMembresia string    `db:"tipo_membresia"`
	}

	err := g.db.SelectContext(ctx, &members,
		"SELECT miembro_id, fecha_alta, tipo_membresia FROM miembros WHERE estado = 'activo'")
	if err != nil {
		return fmt.Errorf("failed to get members: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("no active members found to associate with payments")
	}

	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert payments in batches
	batchSize := 20
	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}

		if err := g.generateBatch(ctx, tx, members, i, end); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// generateBatch generates a batch of payments
func (g *PaymentGenerator) generateBatch(
	ctx context.Context,
	tx *sqlx.Tx,
	members []struct {
		MiembroID     int       `db:"miembro_id"`
		FechaAlta     time.Time `db:"fecha_alta"`
		TipoMembresia string    `db:"tipo_membresia"`
	},
	start,
	end int,
) error {
	// Prepare query
	query := `
		INSERT INTO cuotas_membresia (
			miembro_id, ano, cantidad_pagada, fecha_pago
		) VALUES (
			:miembro_id, :ano, :cantidad_pagada, :fecha_pago
		)
		ON CONFLICT (miembro_id, ano) DO NOTHING
	`

	// Generate payments
	payments := make([]Payment, 0, end-start)

	// Current year and past years to generate payments for
	currentYear := time.Now().Year()

	// For each member, generate historical payments
	for _, member := range members {
		// Get the year the member joined
		joinYear := member.FechaAlta.Year()

		// Generate payments for years between join year and current year
		for year := joinYear; year <= currentYear; year++ {
			// 85% chance of payment for each year
			if g.rand.Float64() < 0.85 {
				// Payment date - sometime during the year
				paymentMonth := time.Month(g.rand.Intn(12) + 1)
				paymentDay := g.rand.Intn(28) + 1
				paymentDate := time.Date(year, paymentMonth, paymentDay, 0, 0, 0, 0, time.UTC)

				// Skip if payment date is in the future
				if paymentDate.After(time.Now()) {
					continue
				}

				// Skip if payment date is before join date
				if paymentDate.Before(member.FechaAlta) {
					continue
				}

				// Determine payment amount based on membership type
				var amount float64
				if member.TipoMembresia == "individual" {
					// Individual membership: 20-40€
					amount = GenerateRandomAmount(g.rand, 20, 40)
				} else {
					// Family membership: 30-60€
					amount = GenerateRandomAmount(g.rand, 30, 60)
				}

				payment := Payment{
					MiembroID:      member.MiembroID,
					Ano:            year,
					CantidadPagada: amount,
					FechaPago:      paymentDate,
				}

				payments = append(payments, payment)

				if len(payments) >= end-start {
					break
				}
			}
		}

		if len(payments) >= end-start {
			break
		}
	}

	// If we didn't generate enough payments, repeat until we reach the target
	if len(payments) < end-start {
		for len(payments) < end-start {
			member := members[g.rand.Intn(len(members))]
			joinYear := member.FechaAlta.Year()

			// Generate a payment for a random year
			year := joinYear + g.rand.Intn(currentYear-joinYear+1)

			// Payment date
			paymentMonth := time.Month(g.rand.Intn(12) + 1)
			paymentDay := g.rand.Intn(28) + 1
			paymentDate := time.Date(year, paymentMonth, paymentDay, 0, 0, 0, 0, time.UTC)

			// Skip if payment date is in the future or before join date
			if paymentDate.After(time.Now()) || paymentDate.Before(member.FechaAlta) {
				continue
			}

			// Determine payment amount
			var amount float64
			if member.TipoMembresia == "individual" {
				amount = GenerateRandomAmount(g.rand, 20, 40)
			} else {
				amount = GenerateRandomAmount(g.rand, 30, 60)
			}

			payment := Payment{
				MiembroID:      member.MiembroID,
				Ano:            year,
				CantidadPagada: amount,
				FechaPago:      paymentDate,
			}

			// Check for duplicates (same member and year)
			isDuplicate := false
			for _, p := range payments {
				if p.MiembroID == payment.MiembroID && p.Ano == payment.Ano {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				payments = append(payments, payment)
			}
		}
	}

	// Limit to requested count
	if len(payments) > end-start {
		payments = payments[:end-start]
	}

	// Insert batch
	_, err := tx.NamedExecContext(ctx, query, payments)
	if err != nil {
		return fmt.Errorf("failed to insert payments: %w", err)
	}

	return nil
}

// GetPaymentsByMember retrieves all payments for a given member
func (g *PaymentGenerator) GetPaymentsByMember(ctx context.Context, miembroID int) ([]Payment, error) {
	var payments []Payment

	query := `
		SELECT cuota_id, miembro_id, ano, cantidad_pagada, fecha_pago
		FROM cuotas_membresia
		WHERE miembro_id = $1
		ORDER BY ano
	`

	err := g.db.SelectContext(ctx, &payments, query, miembroID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, nil
}

// GetPaymentsByYear retrieves all payments for a given year
func (g *PaymentGenerator) GetPaymentsByYear(ctx context.Context, year int) ([]Payment, error) {
	var payments []Payment

	query := `
		SELECT cuota_id, miembro_id, ano, cantidad_pagada, fecha_pago
		FROM cuotas_membresia
		WHERE ano = $1
		ORDER BY miembro_id
	`

	err := g.db.SelectContext(ctx, &payments, query, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, nil
}

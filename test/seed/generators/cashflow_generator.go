package generators

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

// CashflowGenerator generates test cash movements (caja)
type CashflowGenerator struct {
	db   *sqlx.DB
	rand *rand.Rand
}

// Cashflow represents a caja record for generation
type Cashflow struct {
	CajaID        int       `db:"caja_id"`
	MiembroID     *int      `db:"miembro_id"`
	FamiliaID     *int      `db:"familia_id"`
	TipoOperacion string    `db:"tipo_operacion"`
	Monto         float64   `db:"monto"`
	Fecha         time.Time `db:"fecha"`
	Detalle       string    `db:"detalle"`
}

// NewCashflowGenerator creates a new cashflow generator
func NewCashflowGenerator(db *sqlx.DB, seed int64) *CashflowGenerator {
	return &CashflowGenerator{
		db:   db,
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Generate creates n random cash movements
func (g *CashflowGenerator) Generate(ctx context.Context, n int) error {
	// Get members and families to associate with
	var members []struct {
		MiembroID int `db:"miembro_id"`
	}

	err := g.db.SelectContext(ctx, &members, "SELECT miembro_id FROM miembros WHERE estado = 'activo'")
	if err != nil {
		return fmt.Errorf("failed to get members: %w", err)
	}

	var families []struct {
		FamiliaID int `db:"familia_id"`
	}

	err = g.db.SelectContext(ctx, &families, "SELECT familia_id FROM familias")
	if err != nil {
		return fmt.Errorf("failed to get families: %w", err)
	}

	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert cash movements in batches
	batchSize := 20
	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}

		if err := g.generateBatch(ctx, tx, members, families, i, end); err != nil {
			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return err
		}
	}

	return tx.Commit()
}

// generateBatch generates a batch of cash movements
func (g *CashflowGenerator) generateBatch(
	ctx context.Context,
	tx *sqlx.Tx,
	members []struct {
		MiembroID int `db:"miembro_id"`
	},
	families []struct {
		FamiliaID int `db:"familia_id"`
	},
	start,
	end int,
) error {
	// Prepare query
	query := `
		INSERT INTO caja (
			miembro_id, familia_id, tipo_operacion, monto, fecha, detalle
		) VALUES (
			:miembro_id, :familia_id, :tipo_operacion, :monto, :fecha, :detalle
		)
	`

	// Operation types
	operationTypes := []string{
		"ingreso_cuota",
		"gasto_corriente",
		"entrega_fondo",
		"otros_ingresos",
	}

	// Details for each operation type
	operationDetails := map[string][]string{
		"ingreso_cuota": {
			"Cuota de membresía anual",
			"Cuota extraordinaria para evento",
			"Pago atrasado de cuota",
			"Adelanto de cuota",
		},
		"gasto_corriente": {
			"Material de oficina",
			"Alquiler del local",
			"Suministros",
			"Transporte",
			"Gastos evento comunitario",
			"Mantenimiento equipos",
			"Pago servicios internet",
			"Imprenta",
		},
		"entrega_fondo": {
			"Ayuda emergencia familiar",
			"Fondo para repatriación",
			"Ayuda a familia afectada",
			"Préstamo personal",
			"Asistencia médica",
		},
		"otros_ingresos": {
			"Donación particular",
			"Subvención ayuntamiento",
			"Ingreso por evento",
			"Venta de comida comunitaria",
			"Aportación extraordinaria",
			"Devolución préstamo",
		},
	}

	// Generate cash movements
	cashflows := make([]Cashflow, 0, end-start)
	for i := start; i < end; i++ {
		// Select operation type
		operationType := operationTypes[g.rand.Intn(len(operationTypes))]

		// Generate entity association based on operation type
		var miembroID *int
		var familiaID *int

		switch operationType {
		case "ingreso_cuota":
			// Associate with a member 80% of the time, otherwise with a family
			if g.rand.Float64() < 0.8 && len(members) > 0 {
				id := members[g.rand.Intn(len(members))].MiembroID
				miembroID = &id
			} else if len(families) > 0 {
				id := families[g.rand.Intn(len(families))].FamiliaID
				familiaID = &id
			}
		case "entrega_fondo":
			// Always associate with a member or family
			if g.rand.Float64() < 0.5 && len(members) > 0 {
				id := members[g.rand.Intn(len(members))].MiembroID
				miembroID = &id
			} else if len(families) > 0 {
				id := families[g.rand.Intn(len(families))].FamiliaID
				familiaID = &id
			}
		case "gasto_corriente", "otros_ingresos":
			// Usually not associated with a specific member/family
			if g.rand.Float64() < 0.2 { // 20% chance to associate
				if g.rand.Float64() < 0.5 && len(members) > 0 {
					id := members[g.rand.Intn(len(members))].MiembroID
					miembroID = &id
				} else if len(families) > 0 {
					id := families[g.rand.Intn(len(families))].FamiliaID
					familiaID = &id
				}
			}
		}

		// Generate amount based on operation type
		var amount float64
		switch operationType {
		case "ingreso_cuota":
			amount = GenerateRandomAmount(g.rand, 20, 60)
		case "gasto_corriente":
			amount = -GenerateRandomAmount(g.rand, 30, 500) // Negative for expenses
		case "entrega_fondo":
			amount = -GenerateRandomAmount(g.rand, 100, 1000) // Negative for funds given
		case "otros_ingresos":
			amount = GenerateRandomAmount(g.rand, 50, 2000)
		}

		// Generate date - last 3 years
		now := time.Now()
		startDate := now.AddDate(-3, 0, 0)
		date := GenerateRandomDate(g.rand, startDate, now)

		// Select a random detail for this operation type
		detail := operationDetails[operationType][g.rand.Intn(len(operationDetails[operationType]))]

		// Create cashflow record
		cashflow := Cashflow{
			MiembroID:     miembroID,
			FamiliaID:     familiaID,
			TipoOperacion: operationType,
			Monto:         amount,
			Fecha:         date,
			Detalle:       detail,
		}

		cashflows = append(cashflows, cashflow)
	}

	// Insert batch
	_, err := tx.NamedExecContext(ctx, query, cashflows)
	if err != nil {
		return fmt.Errorf("failed to insert cash movements: %w", err)
	}

	return nil
}

// GetCashflowsByPeriod retrieves cash movements for a given period
func (g *CashflowGenerator) GetCashflowsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]Cashflow, error) {
	var cashflows []Cashflow

	query := `
		SELECT caja_id, miembro_id, familia_id, tipo_operacion, monto, fecha, detalle
		FROM caja
		WHERE fecha BETWEEN $1 AND $2
		ORDER BY fecha
	`

	err := g.db.SelectContext(ctx, &cashflows, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash movements: %w", err)
	}

	return cashflows, nil
}

// GetCashflowsByType retrieves cash movements for a given operation type
func (g *CashflowGenerator) GetCashflowsByType(ctx context.Context, operationType string) ([]Cashflow, error) {
	var cashflows []Cashflow

	query := `
		SELECT caja_id, miembro_id, familia_id, tipo_operacion, monto, fecha, detalle
		FROM caja
		WHERE tipo_operacion = $1
		ORDER BY fecha
	`

	err := g.db.SelectContext(ctx, &cashflows, query, operationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash movements: %w", err)
	}

	return cashflows, nil
}

// GetCashflowsByMember retrieves cash movements for a given member
func (g *CashflowGenerator) GetCashflowsByMember(ctx context.Context, miembroID int) ([]Cashflow, error) {
	var cashflows []Cashflow

	query := `
		SELECT caja_id, miembro_id, familia_id, tipo_operacion, monto, fecha, detalle
		FROM caja
		WHERE miembro_id = $1
		ORDER BY fecha
	`

	err := g.db.SelectContext(ctx, &cashflows, query, miembroID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash movements: %w", err)
	}

	return cashflows, nil
}

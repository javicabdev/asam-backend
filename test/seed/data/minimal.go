package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// MinimalDataset represents a minimal dataset for testing
type MinimalDataset struct {
	db     *sqlx.DB
	seeder Seeder
}

// Seeder interface defines the required methods for a seeder
type Seeder interface {
	Clean(ctx context.Context) error
	ExecuteTx(ctx context.Context, fn func(*sqlx.Tx) error) error
	Logf(format string, args ...interface{})
}

// NewMinimalDataset creates a new minimal dataset
func NewMinimalDataset(db *sqlx.DB, seeder Seeder) *MinimalDataset {
	return &MinimalDataset{
		db:     db,
		seeder: seeder,
	}
}

// Seed populates the database with a minimal dataset
func (d *MinimalDataset) Seed(ctx context.Context) error {
	d.seeder.Logf("Seeding minimal dataset")

	// Clean the database first
	if err := d.seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Seed minimal dataset within a transaction
	return d.seeder.ExecuteTx(ctx, func(tx *sqlx.Tx) error {
		// Seed 10 members (5 individual, 5 family)
		if err := d.seedMembers(ctx, tx); err != nil {
			return err
		}

		// Seed 3 families
		if err := d.seedFamilies(ctx, tx); err != nil {
			return err
		}

		// Seed 6 family members
		if err := d.seedFamilyMembers(ctx, tx); err != nil {
			return err
		}

		// Seed 15 payments
		if err := d.seedPayments(ctx, tx); err != nil {
			return err
		}

		// Seed 20 cash movements
		if err := d.seedCashflows(ctx, tx); err != nil {
			return err
		}

		return nil
	})
}

// seedMembers seeds a minimal set of members
func (d *MinimalDataset) seedMembers(ctx context.Context, tx *sqlx.Tx) error {
	d.seeder.Logf("Seeding minimal members dataset")

	// Constants for predictable test data
	now := time.Now()

	// 5 individual members and 5 family members
	members := []map[string]interface{}{
		// Individual members
		{
			"numero_socio":        "SOC-001",
			"tipo_membresia":      "individual",
			"nombre":              "Antonio",
			"apellidos":           "García López",
			"calle_numero_piso":   "Calle Mayor, 15, 3ºA",
			"codigo_postal":       "08001",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-2, -3, 0),
			"fecha_nacimiento":    now.AddDate(-45, 0, 0),
			"documento_identidad": "12345678Z",
			"correo_electronico":  "antonio.garcia@example.com",
			"profesion":           "Profesor",
			"nacionalidad":        "España",
		},
		{
			"numero_socio":        "SOC-002",
			"tipo_membresia":      "individual",
			"nombre":              "María",
			"apellidos":           "Rodríguez Martínez",
			"calle_numero_piso":   "Avenida Diagonal, 230, 5ºB",
			"codigo_postal":       "08013",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-1, -6, 0),
			"fecha_nacimiento":    now.AddDate(-38, 0, 0),
			"documento_identidad": "87654321X",
			"correo_electronico":  "maria.rodriguez@example.com",
			"profesion":           "Abogada",
			"nacionalidad":        "España",
		},
		{
			"numero_socio":        "SOC-003",
			"tipo_membresia":      "individual",
			"nombre":              "Mamadou",
			"apellidos":           "Diop",
			"calle_numero_piso":   "Calle del Carme, 45",
			"codigo_postal":       "08001",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-3, -2, 0),
			"fecha_nacimiento":    now.AddDate(-29, 0, 0),
			"documento_identidad": "X1234567L",
			"correo_electronico":  "mamadou.diop@example.com",
			"profesion":           "Comerciante",
			"nacionalidad":        "Senegal",
		},
		{
			"numero_socio":        "SOC-004",
			"tipo_membresia":      "individual",
			"nombre":              "Fatou",
			"apellidos":           "Ndiaye",
			"calle_numero_piso":   "Rambla Catalunya, 78, 4º3ª",
			"codigo_postal":       "08008",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-2, -8, 0),
			"fecha_nacimiento":    now.AddDate(-32, 0, 0),
			"documento_identidad": "Y7654321P",
			"correo_electronico":  "fatou.ndiaye@example.com",
			"profesion":           "Enfermera",
			"nacionalidad":        "Senegal",
		},
		{
			"numero_socio":        "SOC-005",
			"tipo_membresia":      "individual",
			"nombre":              "Carlos",
			"apellidos":           "Fernández Sánchez",
			"calle_numero_piso":   "Passeig de Gràcia, 55, 2ºC",
			"codigo_postal":       "08007",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "inactivo",
			"fecha_alta":          now.AddDate(-3, -5, 0),
			"fecha_baja":          now.AddDate(-1, -2, 0),
			"fecha_nacimiento":    now.AddDate(-50, 0, 0),
			"documento_identidad": "98765432A",
			"correo_electronico":  "carlos.fernandez@example.com",
			"profesion":           "Arquitecto",
			"nacionalidad":        "España",
		},

		// Family members
		{
			"numero_socio":        "FAM-006",
			"tipo_membresia":      "familiar",
			"nombre":              "Ibrahima",
			"apellidos":           "Fall",
			"calle_numero_piso":   "Carrer de Sants, 120, 1º2ª",
			"codigo_postal":       "08028",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-2, -1, 0),
			"fecha_nacimiento":    now.AddDate(-42, 0, 0),
			"documento_identidad": "Z9876543R",
			"correo_electronico":  "ibrahima.fall@example.com",
			"profesion":           "Mecánico",
			"nacionalidad":        "Senegal",
		},
		{
			"numero_socio":        "FAM-007",
			"tipo_membresia":      "familiar",
			"nombre":              "José",
			"apellidos":           "Gómez Ruiz",
			"calle_numero_piso":   "Calle del Sol, 34, 3ºD",
			"codigo_postal":       "08003",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-1, -9, 0),
			"fecha_nacimiento":    now.AddDate(-48, 0, 0),
			"documento_identidad": "76543210B",
			"correo_electronico":  "jose.gomez@example.com",
			"profesion":           "Funcionario",
			"nacionalidad":        "España",
		},
		{
			"numero_socio":        "FAM-008",
			"tipo_membresia":      "familiar",
			"nombre":              "Modou",
			"apellidos":           "Mbaye",
			"calle_numero_piso":   "Avenida de la Libertad, 88",
			"codigo_postal":       "08018",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-3, -7, 0),
			"fecha_nacimiento":    now.AddDate(-39, 0, 0),
			"documento_identidad": "X8765432T",
			"correo_electronico":  "modou.mbaye@example.com",
			"profesion":           "Cocinero",
			"nacionalidad":        "Senegal",
		},
		{
			"numero_socio":        "FAM-009",
			"tipo_membresia":      "familiar",
			"nombre":              "Ana",
			"apellidos":           "Martínez López",
			"calle_numero_piso":   "Calle Nueva, 12, 5ºA",
			"codigo_postal":       "08005",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "activo",
			"fecha_alta":          now.AddDate(-2, -4, 0),
			"fecha_nacimiento":    now.AddDate(-36, 0, 0),
			"documento_identidad": "65432109C",
			"correo_electronico":  "ana.martinez@example.com",
			"profesion":           "Profesora",
			"nacionalidad":        "España",
		},
		{
			"numero_socio":        "FAM-010",
			"tipo_membresia":      "familiar",
			"nombre":              "Aissatou",
			"apellidos":           "Gueye",
			"calle_numero_piso":   "Calle Alameda, 67, 4ºB",
			"codigo_postal":       "08024",
			"poblacion":           "Barcelona",
			"provincia":           "Barcelona",
			"pais":                "España",
			"estado":              "inactivo",
			"fecha_alta":          now.AddDate(-4, -3, 0),
			"fecha_baja":          now.AddDate(0, -3, 0),
			"fecha_nacimiento":    now.AddDate(-45, 0, 0),
			"documento_identidad": "Y5432109F",
			"correo_electronico":  "aissatou.gueye@example.com",
			"profesion":           "Empresaria",
			"nacionalidad":        "Senegal",
		},
	}

	// Insert members
	for _, member := range members {
		var fechaBaja interface{} = nil
		if member["fecha_baja"] != nil {
			fechaBaja = member["fecha_baja"]
		}

		query := `
			INSERT INTO miembros (
				numero_socio, tipo_membresia, nombre, apellidos,
				calle_numero_piso, codigo_postal, poblacion, provincia, pais,
				estado, fecha_alta, fecha_baja, fecha_nacimiento,
				documento_identidad, correo_electronico, profesion, nacionalidad
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7, $8, $9,
				$10, $11, $12, $13,
				$14, $15, $16, $17
			)
		`

		_, err := tx.ExecContext(ctx, query,
			member["numero_socio"],
			member["tipo_membresia"],
			member["nombre"],
			member["apellidos"],
			member["calle_numero_piso"],
			member["codigo_postal"],
			member["poblacion"],
			member["provincia"],
			member["pais"],
			member["estado"],
			member["fecha_alta"],
			fechaBaja,
			member["fecha_nacimiento"],
			member["documento_identidad"],
			member["correo_electronico"],
			member["profesion"],
			member["nacionalidad"],
		)

		if err != nil {
			return fmt.Errorf("failed to insert member %s: %w", member["numero_socio"], err)
		}
	}

	d.seeder.Logf("Successfully seeded %d members", len(members))
	return nil
}

// seedFamilies seeds a minimal set of families
func (d *MinimalDataset) seedFamilies(ctx context.Context, tx *sqlx.Tx) error {
	d.seeder.Logf("Seeding minimal families dataset")

	// Constants for predictable test data
	now := time.Now()

	// Get family type members IDs (members 6, 7, 8)
	type MemberID struct {
		ID int `db:"miembro_id"`
	}

	var memberIDs []MemberID
	err := tx.SelectContext(ctx, &memberIDs,
		"SELECT miembro_id FROM miembros WHERE tipo_membresia = 'familiar' AND estado = 'activo' LIMIT 3")
	if err != nil {
		return fmt.Errorf("failed to get family member IDs: %w", err)
	}

	if len(memberIDs) < 3 {
		return fmt.Errorf("not enough family members found, needed 3, got %d", len(memberIDs))
	}

	// 3 families
	families := []map[string]interface{}{
		{
			"numero_socio":               "FAM-001",
			"miembro_origen_id":          memberIDs[0].ID, // First family member
			"esposo_nombre":              "Ibrahima",
			"esposo_apellidos":           "Fall",
			"esposa_nombre":              "Mariama",
			"esposa_apellidos":           "Diallo",
			"esposo_fecha_nacimiento":    now.AddDate(-42, 0, 0),
			"esposo_documento_identidad": "Z9876543R",
			"esposo_correo_electronico":  "ibrahima.fall@example.com",
			"esposa_fecha_nacimiento":    now.AddDate(-38, 0, 0),
			"esposa_documento_identidad": "Y2345678M",
			"esposa_correo_electronico":  "mariama.diallo@example.com",
		},
		{
			"numero_socio":               "FAM-002",
			"miembro_origen_id":          memberIDs[1].ID, // Second family member
			"esposo_nombre":              "José",
			"esposo_apellidos":           "Gómez Ruiz",
			"esposa_nombre":              "Carmen",
			"esposa_apellidos":           "Sánchez Martín",
			"esposo_fecha_nacimiento":    now.AddDate(-48, 0, 0),
			"esposo_documento_identidad": "76543210B",
			"esposo_correo_electronico":  "jose.gomez@example.com",
			"esposa_fecha_nacimiento":    now.AddDate(-46, 0, 0),
			"esposa_documento_identidad": "87654321D",
			"esposa_correo_electronico":  "carmen.sanchez@example.com",
		},
		{
			"numero_socio":               "FAM-003",
			"miembro_origen_id":          memberIDs[2].ID, // Third family member
			"esposo_nombre":              "Modou",
			"esposo_apellidos":           "Mbaye",
			"esposa_nombre":              "Khady",
			"esposa_apellidos":           "Sow",
			"esposo_fecha_nacimiento":    now.AddDate(-39, 0, 0),
			"esposo_documento_identidad": "X8765432T",
			"esposo_correo_electronico":  "modou.mbaye@example.com",
			"esposa_fecha_nacimiento":    now.AddDate(-35, 0, 0),
			"esposa_documento_identidad": "Y3456789N",
			"esposa_correo_electronico":  "khady.sow@example.com",
		},
	}

	// Insert families
	for _, family := range families {
		query := `
			INSERT INTO familias (
				numero_socio, miembro_origen_id,
				esposo_nombre, esposo_apellidos, esposa_nombre, esposa_apellidos,
				esposo_fecha_nacimiento, esposo_documento_identidad, esposo_correo_electronico,
				esposa_fecha_nacimiento, esposa_documento_identidad, esposa_correo_electronico
			) VALUES (
				$1, $2,
				$3, $4, $5, $6,
				$7, $8, $9,
				$10, $11, $12
			)
		`

		_, err := tx.ExecContext(ctx, query,
			family["numero_socio"],
			family["miembro_origen_id"],
			family["esposo_nombre"],
			family["esposo_apellidos"],
			family["esposa_nombre"],
			family["esposa_apellidos"],
			family["esposo_fecha_nacimiento"],
			family["esposo_documento_identidad"],
			family["esposo_correo_electronico"],
			family["esposa_fecha_nacimiento"],
			family["esposa_documento_identidad"],
			family["esposa_correo_electronico"],
		)

		if err != nil {
			return fmt.Errorf("failed to insert family %s: %w", family["numero_socio"], err)
		}
	}

	d.seeder.Logf("Successfully seeded %d families", len(families))
	return nil
}

// seedFamilyMembers seeds a minimal set of family members
func (d *MinimalDataset) seedFamilyMembers(ctx context.Context, tx *sqlx.Tx) error {
	d.seeder.Logf("Seeding minimal family members dataset")

	// Get family IDs
	type FamilyID struct {
		ID int `db:"familia_id"`
	}

	var familyIDs []FamilyID
	err := tx.SelectContext(ctx, &familyIDs,
		"SELECT familia_id FROM familias LIMIT 3")
	if err != nil {
		return fmt.Errorf("failed to get family IDs: %w", err)
	}

	if len(familyIDs) < 3 {
		return fmt.Errorf("not enough families found, needed 3, got %d", len(familyIDs))
	}

	// Constants for predictable test data
	now := time.Now()

	// 6 family members (2 per family)
	familyMembers := []map[string]interface{}{
		// First family
		{
			"familia_id":         familyIDs[0].ID,
			"nombre":             "Ousmane",
			"dni_nie":            "",
			"fecha_nacimiento":   now.AddDate(-15, 0, 0),
			"correo_electronico": "ousmane.fall@example.com",
		},
		{
			"familia_id":         familyIDs[0].ID,
			"nombre":             "Rama",
			"dni_nie":            "",
			"fecha_nacimiento":   now.AddDate(-12, 0, 0),
			"correo_electronico": "",
		},

		// Second family
		{
			"familia_id":         familyIDs[1].ID,
			"nombre":             "Pablo",
			"dni_nie":            "12345678X",
			"fecha_nacimiento":   now.AddDate(-18, 0, 0),
			"correo_electronico": "pablo.gomez@example.com",
		},
		{
			"familia_id":         familyIDs[1].ID,
			"nombre":             "Laura",
			"dni_nie":            "12345678Y",
			"fecha_nacimiento":   now.AddDate(-16, 0, 0),
			"correo_electronico": "laura.gomez@example.com",
		},

		// Third family
		{
			"familia_id":         familyIDs[2].ID,
			"nombre":             "Abdoulaye",
			"dni_nie":            "",
			"fecha_nacimiento":   now.AddDate(-10, 0, 0),
			"correo_electronico": "",
		},
		{
			"familia_id":         familyIDs[2].ID,
			"nombre":             "Fatou",
			"dni_nie":            "",
			"fecha_nacimiento":   now.AddDate(-8, 0, 0),
			"correo_electronico": "",
		},
	}

	// Insert family members
	for _, member := range familyMembers {
		query := `
			INSERT INTO familiares (
				familia_id, nombre, dni_nie, fecha_nacimiento, correo_electronico
			) VALUES (
				$1, $2, $3, $4, $5
			)
		`

		_, err := tx.ExecContext(ctx, query,
			member["familia_id"],
			member["nombre"],
			member["dni_nie"],
			member["fecha_nacimiento"],
			member["correo_electronico"],
		)

		if err != nil {
			return fmt.Errorf("failed to insert family member %s: %w", member["nombre"], err)
		}
	}

	d.seeder.Logf("Successfully seeded %d family members", len(familyMembers))
	return nil
}

// seedPayments seeds a minimal set of membership payments
func (d *MinimalDataset) seedPayments(ctx context.Context, tx *sqlx.Tx) error {
	d.seeder.Logf("Seeding minimal payments dataset")

	// Get active member IDs
	type MemberID struct {
		ID            int       `db:"miembro_id"`
		TipoMembresia string    `db:"tipo_membresia"`
		FechaAlta     time.Time `db:"fecha_alta"`
	}

	var memberIDs []MemberID
	err := tx.SelectContext(ctx, &memberIDs,
		"SELECT miembro_id, tipo_membresia, fecha_alta FROM miembros WHERE estado = 'activo'")
	if err != nil {
		return fmt.Errorf("failed to get member IDs: %w", err)
	}

	// Current year and last year
	currentYear := time.Now().Year()
	lastYear := currentYear - 1

	// Generate payments for each member
	var payments []map[string]interface{}

	for _, member := range memberIDs {
		// Skip if member joined after last year
		if member.FechaAlta.Year() > lastYear {
			continue
		}

		// Determine payment amount based on membership type
		amount := 30.0
		if member.TipoMembresia == "familiar" {
			amount = 50.0
		}

		// Create payment for last year
		lastYearPayment := map[string]interface{}{
			"miembro_id":      member.ID,
			"ano":             lastYear,
			"cantidad_pagada": amount,
			"fecha_pago":      time.Date(lastYear, 3, 15, 0, 0, 0, 0, time.UTC),
		}
		payments = append(payments, lastYearPayment)

		// 80% chance to have payment for current year
		if member.FechaAlta.Year() <= currentYear && time.Now().Month() > 2 {
			currentYearPayment := map[string]interface{}{
				"miembro_id":      member.ID,
				"ano":             currentYear,
				"cantidad_pagada": amount,
				"fecha_pago":      time.Date(currentYear, 2, 20, 0, 0, 0, 0, time.UTC),
			}
			payments = append(payments, currentYearPayment)
		}
	}

	// Insert payments
	for _, payment := range payments {
		query := `
			INSERT INTO cuotas_membresia (
				miembro_id, ano, cantidad_pagada, fecha_pago
			) VALUES (
				$1, $2, $3, $4
			)
			ON CONFLICT (miembro_id, ano) DO NOTHING
		`

		_, err := tx.ExecContext(ctx, query,
			payment["miembro_id"],
			payment["ano"],
			payment["cantidad_pagada"],
			payment["fecha_pago"],
		)

		if err != nil {
			return fmt.Errorf("failed to insert payment for member %d, year %d: %w",
				payment["miembro_id"], payment["ano"], err)
		}
	}

	d.seeder.Logf("Successfully seeded %d payments", len(payments))
	return nil
}

// seedCashflows seeds a minimal set of cash movements
func (d *MinimalDataset) seedCashflows(ctx context.Context, tx *sqlx.Tx) error {
	d.seeder.Logf("Seeding minimal cash movements dataset")

	// Get some member IDs
	type MemberID struct {
		ID int `db:"miembro_id"`
	}

	var memberIDs []MemberID
	err := tx.SelectContext(ctx, &memberIDs,
		"SELECT miembro_id FROM miembros WHERE estado = 'activo' LIMIT 5")
	if err != nil {
		return fmt.Errorf("failed to get member IDs: %w", err)
	}

	// Get some family IDs
	type FamilyID struct {
		ID int `db:"familia_id"`
	}

	var familyIDs []FamilyID
	err = tx.SelectContext(ctx, &familyIDs,
		"SELECT familia_id FROM familias LIMIT 3")
	if err != nil {
		return fmt.Errorf("failed to get family IDs: %w", err)
	}

	// Current date and previous dates
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)
	lastYear := now.AddDate(-1, 0, 0)

	// Generate cash movements
	var cashflows []map[string]interface{}

	// Membership fee income
	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     memberIDs[0].ID,
		"familia_id":     nil,
		"tipo_operacion": "ingreso_cuota",
		"monto":          30.0,
		"fecha":          lastMonth,
		"detalle":        "Cuota de membresía anual",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     memberIDs[1].ID,
		"familia_id":     nil,
		"tipo_operacion": "ingreso_cuota",
		"monto":          30.0,
		"fecha":          lastMonth.AddDate(0, 0, -2),
		"detalle":        "Cuota de membresía anual",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     familyIDs[0].ID,
		"tipo_operacion": "ingreso_cuota",
		"monto":          50.0,
		"fecha":          lastMonth.AddDate(0, 0, -5),
		"detalle":        "Cuota de membresía anual",
	})

	// Current expenses
	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     nil,
		"tipo_operacion": "gasto_corriente",
		"monto":          -120.0,
		"fecha":          twoMonthsAgo,
		"detalle":        "Alquiler del local",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     nil,
		"tipo_operacion": "gasto_corriente",
		"monto":          -45.0,
		"fecha":          lastMonth.AddDate(0, 0, -10),
		"detalle":        "Material de oficina",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     nil,
		"tipo_operacion": "gasto_corriente",
		"monto":          -80.0,
		"fecha":          lastMonth,
		"detalle":        "Gastos evento comunitario",
	})

	// Fund deliveries
	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     memberIDs[2].ID,
		"familia_id":     nil,
		"tipo_operacion": "entrega_fondo",
		"monto":          -300.0,
		"fecha":          twoMonthsAgo.AddDate(0, 0, 5),
		"detalle":        "Ayuda emergencia familiar",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     familyIDs[1].ID,
		"tipo_operacion": "entrega_fondo",
		"monto":          -500.0,
		"fecha":          lastYear.AddDate(0, 2, 0),
		"detalle":        "Fondo para repatriación",
	})

	// Other income
	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     nil,
		"tipo_operacion": "otros_ingresos",
		"monto":          750.0,
		"fecha":          lastYear.AddDate(0, 6, 0),
		"detalle":        "Subvención ayuntamiento",
	})

	cashflows = append(cashflows, map[string]interface{}{
		"miembro_id":     nil,
		"familia_id":     nil,
		"tipo_operacion": "otros_ingresos",
		"monto":          320.0,
		"fecha":          twoMonthsAgo.AddDate(0, 0, -5),
		"detalle":        "Ingreso por evento",
	})

	// Insert cash movements
	for _, cashflow := range cashflows {
		query := `
			INSERT INTO caja (
				miembro_id, familia_id, tipo_operacion, monto, fecha, detalle
			) VALUES (
				$1, $2, $3, $4, $5, $6
			)
		`

		var miembroID interface{} = nil
		if cashflow["miembro_id"] != nil {
			miembroID = cashflow["miembro_id"]
		}

		var familiaID interface{} = nil
		if cashflow["familia_id"] != nil {
			familiaID = cashflow["familia_id"]
		}

		_, err := tx.ExecContext(ctx, query,
			miembroID,
			familiaID,
			cashflow["tipo_operacion"],
			cashflow["monto"],
			cashflow["fecha"],
			cashflow["detalle"],
		)

		if err != nil {
			return fmt.Errorf("failed to insert cash movement: %w", err)
		}
	}

	d.seeder.Logf("Successfully seeded %d cash movements", len(cashflows))
	return nil
}

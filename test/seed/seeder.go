package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	// Assuming data package is correctly located relative to this seed package
	"github.com/javicabdev/asam-backend/test/seed/data"
	"github.com/javicabdev/asam-backend/test/seed/generators"
)

// Seeder represents the main seeding coordinator
type Seeder struct {
	DB *sqlx.DB
	// Configuration options
	RandomSeed  int64 // Used for deterministic random data generation
	EnableLog   bool
	Concurrency int // Number of concurrent operations for batch seeding
}

// NewSeeder creates a new seeder instance
func NewSeeder(db *sqlx.DB) *Seeder {
	return &Seeder{
		DB:          db,
		RandomSeed:  time.Now().UnixNano(),
		EnableLog:   true,
		Concurrency: 5, // Default concurrency
	}
}

// GetDB returns the database connection
func (s *Seeder) GetDB() *sqlx.DB {
	return s.DB
}

// WithRandomSeed sets a specific random seed for deterministic generation
func (s *Seeder) WithRandomSeed(seed int64) *Seeder {
	s.RandomSeed = seed
	return s
}

// WithLogging enables or disables logging
func (s *Seeder) WithLogging(enabled bool) *Seeder {
	s.EnableLog = enabled
	return s
}

// WithConcurrency sets the concurrency level for batch operations
func (s *Seeder) WithConcurrency(c int) *Seeder {
	s.Concurrency = c
	return s
}

// Logf logs a message if logging is enabled
func (s *Seeder) Logf(format string, args ...any) {
	if s.EnableLog {
		log.Printf(format, args...)
	}
}

// ExecuteTx executes a callback within a transaction
func (s *Seeder) ExecuteTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.DB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err) // Added context to error
	}

	// Use a deferred function for robust rollback/commit logic
	defer func() {
		if p := recover(); p != nil {
			// A panic occurred during fn(tx)
			s.Logf("Panic recovered during transaction: %v. Rolling back.", p)
			_ = tx.Rollback() // Ignore rollback error after panic
			panic(p)          // Re-panic after rollback
		} else if err != nil {
			// fn(tx) returned an error
			s.Logf("Transaction function failed: %v. Rolling back.", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				s.Logf("Rollback failed after transaction error: %v", rbErr)
				// Combine errors to provide more context
				err = fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
			}
		} else {
			// fn(tx) succeeded, commit
			s.Logf("Transaction function succeeded. Committing.")
			err = tx.Commit()
			if err != nil {
				s.Logf("Commit failed: %v", err)
				err = fmt.Errorf("failed to commit transaction: %w", err) // Add context
			}
		}
	}()

	err = fn(tx) // Execute the provided function
	return err   // Return the error from fn or commit
}

// Clean removes all data from tables in the correct order to handle foreign key constraints
func (s *Seeder) Clean(ctx context.Context) error {
	// IMPORTANT: Double-check these names match your DB schema exactly (case-sensitive)
	// Order matters for deletion if constraints are not disabled or deferred properly.
	// Listing tables from "most dependent" to "least dependent" is critical for DELETE.
	tables := []string{
		"caja",                // Depends on miembros, familias
		"cuotas_membresia",    // Depends on miembros
		"historial_membresia", // Depends on miembros (assuming)
		"telefonos",           // Depends on miembros (assuming)
		"familiares",          // Depends on familias
		"familias",            // Depends on miembros (miembro_origen_id)
		"refresh_tokens",      // Depends on users (assuming)
		"miembros",            // Depends on users (if linked)
		"users",               // Least dependent (often)
	}

	s.Logf("Starting database cleaning process...")

	return s.ExecuteTx(ctx, func(tx *sqlx.Tx) error {
		// Ya no intentamos deshabilitar las restricciones de clave foránea
		// Simplemente eliminamos las tablas en el orden correcto

		// Iterate through tables to delete data
		for _, table := range tables {
			s.Logf("Cleaning table: %s", table)

			// Use DELETE for all tables
			sqlQuery := fmt.Sprintf(`DELETE FROM "%s";`, table)
			s.Logf("Executing SQL: %s", sqlQuery)

			// Execute the query
			if _, err := tx.ExecContext(ctx, sqlQuery); err != nil {
				// Si hay un error, podría ser porque la tabla no existe
				if strings.Contains(err.Error(), "does not exist") {
					s.Logf("Table '%s' does not exist, skipping", table)
					continue
				}
				// Cualquier otro error es un problema real
				return fmt.Errorf("failed to execute query [%s] for table %s: %w", sqlQuery, table, err)
			}

			// Reset sequence after DELETE
			// Determine the sequence name based on the table name
			var seqName string
			parts := splitTableName(table)
			if len(parts) == 2 {
				seqName = fmt.Sprintf("%s_%s_id_seq", parts[0], parts[1]) // e.g. public_users_id_seq
			} else {
				seqName = fmt.Sprintf("%s_id_seq", table) // Fallback
			}

			resetSeqQuery := fmt.Sprintf("ALTER SEQUENCE IF EXISTS \"%s\" RESTART WITH 1;", seqName)
			s.Logf("Executing SQL: %s", resetSeqQuery)
			if _, err := tx.ExecContext(ctx, resetSeqQuery); err != nil {
				// Log warning instead of failing, as sequence might not exist or name differs.
				s.Logf("Warning: Failed to reset sequence %s (it might not exist or name is different): %v", seqName, err)
			}
		}

		s.Logf("Finished cleaning tables within transaction.")
		return nil // Signal success for this part of the transaction
	})
}

// Helper function to split schema.table names (basic version)
func splitTableName(fqtn string) []string {
	// This is a simplified version, might need refinement for complex names
	parts := make([]string, 0, 2)
	if dotIndex := strings.Index(fqtn, "."); dotIndex != -1 {
		parts = append(parts, fqtn[:dotIndex])   // schema
		parts = append(parts, fqtn[dotIndex+1:]) // table
	} else {
		parts = append(parts, fqtn) // table only
	}
	return parts
}

// SeedAll runs all seeders to populate the database
func (s *Seeder) SeedAll(ctx context.Context) error {
	s.Logf("Starting full database seeding process...")

	// Clean the database first
	s.Logf("Step 1: Cleaning the database...")
	if err := s.Clean(ctx); err != nil {
		// Make sure the error from Clean provides enough detail
		return fmt.Errorf("failed during database cleaning phase: %w", err)
	}
	s.Logf("Database cleaning successful.")

	// Seed in order of dependencies - IMPORTANT
	// Ensure the order respects foreign key constraints if not handled by deferring/disabling them perfectly.
	// Example: Users -> Miembros -> Familias -> Familiares -> Cuotas -> Caja

	// s.Logf("Step 2: Seeding 'users'...")
	// if err := s.SeedUsers(ctx); err != nil { // Assuming you have a SeedUsers method
	//    return fmt.Errorf("failed to seed users: %w", err)
	// }

	s.Logf("Step 2: Seeding 'miembros'...")
	if err := s.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("failed to seed miembros: %w", err)
	}

	s.Logf("Step 3: Seeding 'familias'...")
	if err := s.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("failed to seed familias: %w", err)
	}

	s.Logf("Step 4: Seeding 'familiares'...")
	if err := s.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("failed to seed familiares: %w", err)
	}

	// s.Logf("Step 5: Seeding 'telefonos'...")
	// if err := s.SeedTelefonos(ctx); err != nil {
	//    return fmt.Errorf("failed to seed telefonos: %w", err)
	// }

	s.Logf("Step 5: Seeding 'cuotas_membresia'...")
	if err := s.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("failed to seed cuotas_membresia: %w", err)
	}

	// s.Logf("Step 7: Seeding 'historial_membresia'...")
	// if err := s.SeedHistorialMembresia(ctx); err != nil {
	//    return fmt.Errorf("failed to seed historial_membresia: %w", err)
	// }

	s.Logf("Step 6: Seeding 'caja'...")
	if err := s.SeedCaja(ctx); err != nil {
		return fmt.Errorf("failed to seed caja: %w", err)
	}

	// s.Logf("Step 9: Seeding 'refresh_tokens'...")
	// if err := s.SeedRefreshTokens(ctx); err != nil {
	//    return fmt.Errorf("failed to seed refresh_tokens: %w", err)
	// }

	s.Logf("Database seeding completed successfully.")
	return nil
}

// SeedMiembros populates the miembros table using its generator
func (s *Seeder) SeedMiembros(ctx context.Context) error {
	s.Logf("--> Seeding Miembros table using generator")
	// Ensure generator exists and handles potential errors
	gen := generators.NewMemberGenerator(s.DB, s.RandomSeed)
	if gen == nil {
		return fmt.Errorf("failed to create member generator")
	}
	// Define how many members to generate for the standard SeedAll call
	numMembers := 50 // Example number
	s.Logf("----> Generating %d members", numMembers)
	return gen.Generate(ctx, numMembers)
}

// SeedFamilias populates the familias table using its generator
func (s *Seeder) SeedFamilias(ctx context.Context) error {
	s.Logf("--> Seeding Familias table using generator")
	gen := generators.NewFamilyGenerator(s.DB, s.RandomSeed)
	if gen == nil {
		return fmt.Errorf("failed to create family generator")
	}
	numFamilies := 20 // Example number
	s.Logf("----> Generating %d families", numFamilies)
	// Ensure the generator handles dependencies (e.g., requires existing 'miembros' of type 'familiar')
	return gen.Generate(ctx, numFamilies)
}

// SeedFamiliares populates the familiares table using its generator
func (s *Seeder) SeedFamiliares(ctx context.Context) error {
	s.Logf("--> Seeding Familiares table using generator")
	gen := generators.NewFamilyMemberGenerator(s.DB, s.RandomSeed)
	if gen == nil {
		return fmt.Errorf("failed to create family member generator")
	}
	numFamilyMembers := 40 // Example number
	s.Logf("----> Generating %d family members", numFamilyMembers)
	// Ensure this generator correctly handles dependencies (requires existing 'familias')
	return gen.Generate(ctx, numFamilyMembers)
}

// SeedCuotasMembresia populates the cuotas_membresia table using its generator
func (s *Seeder) SeedCuotasMembresia(ctx context.Context) error {
	s.Logf("--> Seeding CuotasMembresia table using generator")
	gen := generators.NewPaymentGenerator(s.DB, s.RandomSeed)
	if gen == nil {
		return fmt.Errorf("failed to create payment generator")
	}
	numPayments := 100 // Example number
	s.Logf("----> Generating %d payments", numPayments)
	// Ensure this generator correctly handles dependencies (requires existing 'miembros')
	return gen.Generate(ctx, numPayments)
}

// SeedCaja populates the caja table using its generator
func (s *Seeder) SeedCaja(ctx context.Context) error {
	s.Logf("--> Seeding Caja table using generator")
	gen := generators.NewCashflowGenerator(s.DB, s.RandomSeed)
	if gen == nil {
		return fmt.Errorf("failed to create cashflow generator")
	}
	numCashflows := 200 // Example number
	s.Logf("----> Generating %d cashflow entries", numCashflows)
	// Ensure this generator correctly handles dependencies (e.g., related 'miembros' or 'familias')
	return gen.Generate(ctx, numCashflows)
}

// SeedMinimalDataset seeds a minimal dataset for testing using the data package
func (s *Seeder) SeedMinimalDataset(ctx context.Context) error {
	s.Logf("Seeding minimal dataset using data package")

	// *** CORRECTED FUNCTION NAME ***
	// Pass the current Seeder (s) which should implement the data.Seeder interface
	// NOTE: This assumes *seed.Seeder satisfies the data.Seeder interface defined in minimal.go
	// If not, you'll get a type error here. See explanation above.
	minimalSeederRunner := data.NewMinimalDataset(s.DB, s) // Pass DB and the seeder instance

	if minimalSeederRunner == nil {
		// This check might not be necessary if NewMinimalDataset never returns nil,
		// but good practice if it could.
		return fmt.Errorf("failed to create minimal dataset runner (NewMinimalDataset returned nil)")
	}

	// The Seed method within the data package's MinimalDataset struct handles the actual seeding logic
	s.Logf("----> Calling Seed method on MinimalDataset runner")
	err := minimalSeederRunner.Seed(ctx)
	if err != nil {
		return fmt.Errorf("minimal dataset runner failed to seed: %w", err)
	}
	s.Logf("Minimal dataset seeding completed via data package.")
	return nil
}

// SeedFullDataset seeds a complete dataset for development using the data package
func (s *Seeder) SeedFullDataset(ctx context.Context) error {
	s.Logf("Seeding full dataset using data package")

	// *** CORRECTED FUNCTION NAME ***
	// Pass the current Seeder (s) which should implement the data.Seeder interface
	// NOTE: Similar type compatibility assumption as in SeedMinimalDataset.
	fullSeederRunner := data.NewFullDataset(s.DB, s) // Pass DB and the seeder instance

	if fullSeederRunner == nil {
		return fmt.Errorf("failed to create full dataset runner (NewFullDataset returned nil)")
	}

	// The Seed method within the data package's FullDataset struct handles the actual seeding logic
	s.Logf("----> Calling Seed method on FullDataset runner")
	err := fullSeederRunner.Seed(ctx)
	if err != nil {
		return fmt.Errorf("full dataset runner failed to seed: %w", err)
	}
	s.Logf("Full dataset seeding completed via data package.")
	return nil
}

// --- Placeholder functions for tables potentially seeded by generators or data packages ---
// You might not need explicit SeedUsers, SeedRefreshTokens etc. methods here if
// the generators or data packages handle them internally. Keep them if you need
// separate control over seeding these tables from the main `seed` package.

func (s *Seeder) SeedUsers(_ context.Context) error {
	s.Logf("--> Seeding Users table (Placeholder - Implement if needed)")
	// Example:
	// gen := generators.NewUserGenerator(s.DB, s.RandomSeed)
	// if gen == nil { return fmt.Errorf("failed to create user generator") }
	// return gen.Generate(ctx, 10)
	return nil // Return nil if users are seeded elsewhere or not needed here
}

func (s *Seeder) SeedRefreshTokens(_ context.Context) error {
	s.Logf("--> Seeding Refresh Tokens table (Placeholder - Implement if needed)")
	// Add actual refresh token seeding logic here, likely dependent on users
	return nil
}

func (s *Seeder) SeedTelefonos(_ context.Context) error {
	s.Logf("--> Seeding Telefonos table (Placeholder - Implement if needed)")
	// Add actual telefono seeding logic here, likely dependent on miembros
	return nil
}

func (s *Seeder) SeedHistorialMembresia(ctx context.Context) error {
	s.Logf("--> Seeding Historial Membresia table (Placeholder - Implement if needed)")
	// Add actual historial membresia seeding logic here, likely dependent on miembros
	return nil
}

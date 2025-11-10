package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Script de limpieza de base de datos para producción
// IMPORTANTE: Este script eliminará TODOS los datos excepto los usuarios especificados
// Uso: go run cmd/cleanup-db/main.go -env production -keep-users "Babacar,javi"

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

var (
	env       = flag.String("env", "", "Environment: production (required)")
	keepUsers = flag.String("keep-users", "", "Comma-separated usernames to keep (e.g., 'Babacar,javi')")
	dryRun    = flag.Bool("dry-run", false, "Simulate cleanup without making changes")
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type CleanupStats struct {
	MembersDeleted          int
	FamiliesDeleted         int
	FamiliarsDeleted        int
	TelephonesDeleted       int
	PaymentsDeleted         int
	CashFlowsDeleted        int
	MembershipFeesDeleted   int
	UsersDeleted            int
	RefreshTokensDeleted    int
	VerificationTokensDeleted int
	UsersKept               []string
}

func main() {
	flag.Parse()

	if *env == "" {
		log.Fatal(colorRed + "Error: -env flag is required. Use: -env production" + colorReset)
	}

	if *env != "production" {
		log.Fatal(colorRed + "Error: Only 'production' environment is supported for this script" + colorReset)
	}

	if *keepUsers == "" {
		log.Fatal(colorRed + "Error: -keep-users flag is required. Example: -keep-users 'Babacar,javi'" + colorReset)
	}

	usersToKeep := parseKeepUsers(*keepUsers)
	if len(usersToKeep) == 0 {
		log.Fatal(colorRed + "Error: At least one user must be specified to keep" + colorReset)
	}

	config := getDBConfig(*env)

	fmt.Println(colorYellow + "========================================" + colorReset)
	fmt.Println(colorRed + "⚠️  DATABASE CLEANUP SCRIPT" + colorReset)
	fmt.Println(colorYellow + "========================================" + colorReset)
	fmt.Printf("Environment: %s%s%s\n", colorBlue, *env, colorReset)
	fmt.Printf("Database: %s%s@%s:%s/%s%s\n", colorBlue, config.User, config.Host, config.Port, config.DBName, colorReset)
	fmt.Printf("Users to KEEP: %s%v%s\n", colorGreen, usersToKeep, colorReset)
	if *dryRun {
		fmt.Printf("Mode: %sDRY RUN (no changes will be made)%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("Mode: %sPRODUCTION (REAL DELETION)%s\n", colorRed, colorReset)
	}
	fmt.Println(colorYellow + "========================================" + colorReset)
	fmt.Println()
	fmt.Println(colorRed + "This will DELETE ALL DATA except the specified users!" + colorReset)
	fmt.Println("The following will be deleted:")
	fmt.Println("  - All members (and associated data)")
	fmt.Println("  - All families and familiars")
	fmt.Println("  - All payments and cash flows")
	fmt.Println("  - All membership fees")
	fmt.Println("  - All users except: " + strings.Join(usersToKeep, ", "))
	fmt.Println("  - All tokens except for kept users")
	fmt.Println()

	if !*dryRun {
		if !confirmAction("Do you want to continue? Type 'DELETE ALL' to proceed: ", "DELETE ALL") {
			fmt.Println(colorYellow + "Operation cancelled by user" + colorReset)
			return
		}
	}

	// Connect to database
	db, err := connectDB(config)
	if err != nil {
		log.Fatalf(colorRed+"Failed to connect to database: %v"+colorReset, err)
	}
	defer db.Close()

	// Verify users exist
	if err := verifyUsersExist(db, usersToKeep); err != nil {
		log.Fatalf(colorRed+"Error: %v"+colorReset, err)
	}

	// Create backup recommendation
	fmt.Println(colorYellow + "\n⚠️  IMPORTANT: Make sure you have a recent backup!" + colorReset)
	if !*dryRun {
		if !confirmAction("Have you made a backup? Type 'YES' to continue: ", "YES") {
			fmt.Println(colorYellow + "Operation cancelled. Please create a backup first." + colorReset)
			return
		}
	}

	// Perform cleanup
	stats, err := performCleanup(db, usersToKeep, *dryRun)
	if err != nil {
		log.Fatalf(colorRed+"Cleanup failed: %v"+colorReset, err)
	}

	// Display results
	displayResults(stats, *dryRun)
}

func parseKeepUsers(userStr string) []string {
	users := strings.Split(userStr, ",")
	var result []string
	for _, u := range users {
		trimmed := strings.TrimSpace(u)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getDBConfig(env string) DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", ""),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", ""),
		SSLMode:  getEnv("DB_SSL_MODE", "require"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func connectDB(config DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func verifyUsersExist(db *sql.DB, usernames []string) error {
	placeholders := make([]string, len(usernames))
	args := make([]interface{}, len(usernames))
	for i, username := range usernames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = username
	}

	query := fmt.Sprintf("SELECT username FROM users WHERE username IN (%s)", strings.Join(placeholders, ","))
	rows, err := db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to verify users: %w", err)
	}
	defer rows.Close()

	foundUsers := make(map[string]bool)
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return fmt.Errorf("failed to scan username: %w", err)
		}
		foundUsers[username] = true
	}

	var missingUsers []string
	for _, username := range usernames {
		if !foundUsers[username] {
			missingUsers = append(missingUsers, username)
		}
	}

	if len(missingUsers) > 0 {
		return fmt.Errorf("the following users do not exist in the database: %v", missingUsers)
	}

	return nil
}

func performCleanup(db *sql.DB, keepUsers []string, dryRun bool) (*CleanupStats, error) {
	stats := &CleanupStats{UsersKept: keepUsers}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	fmt.Println(colorBlue + "\nStarting cleanup process..." + colorReset)

	// Get user IDs to keep
	placeholders := make([]string, len(keepUsers))
	args := make([]interface{}, len(keepUsers))
	for i, username := range keepUsers {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = username
	}
	keepUsersClause := strings.Join(placeholders, ",")

	// 1. Delete verification tokens for users to be deleted
	fmt.Println("→ Deleting verification tokens...")
	query := fmt.Sprintf("DELETE FROM verification_tokens WHERE user_id NOT IN (SELECT id FROM users WHERE username IN (%s))", keepUsersClause)
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM verification_tokens WHERE user_id NOT IN (SELECT id FROM users WHERE username IN ("+keepUsersClause+"))", args...).Scan(&stats.VerificationTokensDeleted)
	} else {
		result, err := tx.Exec(query, args...)
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.VerificationTokensDeleted = int(rows)
		}
	}

	// 2. Delete refresh tokens for users to be deleted
	fmt.Println("→ Deleting refresh tokens...")
	query = fmt.Sprintf("DELETE FROM refresh_tokens WHERE user_id NOT IN (SELECT id FROM users WHERE username IN (%s))", keepUsersClause)
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM refresh_tokens WHERE user_id NOT IN (SELECT id FROM users WHERE username IN ("+keepUsersClause+"))", args...).Scan(&stats.RefreshTokensDeleted)
	} else {
		result, err := tx.Exec(query, args...)
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.RefreshTokensDeleted = int(rows)
		}
	}

	// 3. Delete cash flows (no FK to users, so delete all)
	fmt.Println("→ Deleting cash flows...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM cash_flows").Scan(&stats.CashFlowsDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM cash_flows")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.CashFlowsDeleted = int(rows)
		}
	}

	// 4. Delete payments (no FK to users, so delete all)
	fmt.Println("→ Deleting payments...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM payments").Scan(&stats.PaymentsDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM payments")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.PaymentsDeleted = int(rows)
		}
	}

	// 5. Delete membership fees
	fmt.Println("→ Deleting membership fees...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM membership_fees").Scan(&stats.MembershipFeesDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM membership_fees")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.MembershipFeesDeleted = int(rows)
		}
	}

	// 6. Delete telephones
	fmt.Println("→ Deleting telephones...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM telephones").Scan(&stats.TelephonesDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM telephones")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.TelephonesDeleted = int(rows)
		}
	}

	// 7. Delete familiars
	fmt.Println("→ Deleting familiars...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM familiars").Scan(&stats.FamiliarsDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM familiars")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.FamiliarsDeleted = int(rows)
		}
	}

	// 8. Delete families
	fmt.Println("→ Deleting families...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM families").Scan(&stats.FamiliesDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM families")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.FamiliesDeleted = int(rows)
		}
	}

	// 9. Update users to remove member_id references before deleting members
	fmt.Println("→ Clearing member references from kept users...")
	query = fmt.Sprintf("UPDATE users SET member_id = NULL WHERE username IN (%s)", keepUsersClause)
	if !dryRun {
		_, err = tx.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to clear member_id from users: %w", err)
		}
	}

	// 10. Delete members
	fmt.Println("→ Deleting members...")
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM members").Scan(&stats.MembersDeleted)
	} else {
		result, err := tx.Exec("DELETE FROM members")
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.MembersDeleted = int(rows)
		}
	}

	// 11. Delete users (except the ones to keep)
	fmt.Println("→ Deleting users (except " + strings.Join(keepUsers, ", ") + ")...")
	query = fmt.Sprintf("DELETE FROM users WHERE username NOT IN (%s)", keepUsersClause)
	if dryRun {
		err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE username NOT IN ("+keepUsersClause+")", args...).Scan(&stats.UsersDeleted)
	} else {
		result, err := tx.Exec(query, args...)
		if err == nil {
			rows, _ := result.RowsAffected()
			stats.UsersDeleted = int(rows)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("cleanup failed: %w", err)
	}

	if dryRun {
		fmt.Println(colorYellow + "\nDry run completed - rolling back transaction (no changes made)" + colorReset)
		tx.Rollback()
	} else {
		fmt.Println(colorGreen + "\nCommitting changes..." + colorReset)
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return stats, nil
}

func displayResults(stats *CleanupStats, dryRun bool) {
	fmt.Println()
	fmt.Println(colorGreen + "========================================" + colorReset)
	if dryRun {
		fmt.Println(colorYellow + "DRY RUN RESULTS (No changes made)" + colorReset)
	} else {
		fmt.Println(colorGreen + "✓ CLEANUP COMPLETED SUCCESSFULLY" + colorReset)
	}
	fmt.Println(colorGreen + "========================================" + colorReset)
	fmt.Printf("Members deleted:           %d\n", stats.MembersDeleted)
	fmt.Printf("Families deleted:          %d\n", stats.FamiliesDeleted)
	fmt.Printf("Familiars deleted:         %d\n", stats.FamiliarsDeleted)
	fmt.Printf("Telephones deleted:        %d\n", stats.TelephonesDeleted)
	fmt.Printf("Payments deleted:          %d\n", stats.PaymentsDeleted)
	fmt.Printf("Cash flows deleted:        %d\n", stats.CashFlowsDeleted)
	fmt.Printf("Membership fees deleted:   %d\n", stats.MembershipFeesDeleted)
	fmt.Printf("Users deleted:             %d\n", stats.UsersDeleted)
	fmt.Printf("Refresh tokens deleted:    %d\n", stats.RefreshTokensDeleted)
	fmt.Printf("Verification tokens deleted: %d\n", stats.VerificationTokensDeleted)
	fmt.Println(colorGreen + "----------------------------------------" + colorReset)
	fmt.Printf(colorGreen+"Users kept: %v\n"+colorReset, stats.UsersKept)
	fmt.Println(colorGreen + "========================================" + colorReset)
}

func confirmAction(prompt, expectedAnswer string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(colorYellow + prompt + colorReset)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(response)
	return response == expectedAnswer
}

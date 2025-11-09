package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (db *DB) ensureMigrationsTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP NOT NULL DEFAULT NOW()
        )
    `
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

func (db *DB) isMigrationApplied(version string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`
	err := db.conn.QueryRow(query, version).Scan(&exists)
	return exists, err
}

func (db *DB) Up() error {
	migrationsPath := "./migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	if err := db.ensureMigrationsTable(); err != nil {
		return err
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}

	sort.Strings(upFiles)

	for _, file := range upFiles {
		version := strings.Split(file, "_")[0]

		applied, err := db.isMigrationApplied(version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			log.Printf("⊘ Skipping %s (already applied)\n", file)
			continue
		}

		log.Printf("→ Applying migration: %s\n", file)

		fullPath := filepath.Join(migrationsPath, file)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to mark migration as applied: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Printf("✓ Applied: %s\n", file)
	}

	log.Println("✓ All migrations completed successfully")
	return nil
}

func (db *DB) Down() error {
	migrationsPath := "./migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	if err := db.ensureMigrationsTable(); err != nil {
		return err
	}

	var lastVersion string
	query := `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`
	err := db.conn.QueryRow(query).Scan(&lastVersion)
	if err == sql.ErrNoRows {
		log.Println("No migrations to rollback")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var downFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), lastVersion) && strings.HasSuffix(entry.Name(), ".down.sql") {
			downFile = entry.Name()
			break
		}
	}

	if downFile == "" {
		return fmt.Errorf("down migration file not found for version %s", lastVersion)
	}

	log.Printf("→ Rolling back: %s\n", downFile)

	fullPath := filepath.Join(migrationsPath, downFile)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", downFile, err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute rollback %s: %w", downFile, err)
	}

	if _, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = $1`, lastVersion); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unmark migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✓ Rolled back: %s\n", downFile)
	return nil
}

// var migrationsFS embed.FS

// func (db *DB) RunMigrations() error {
// 	entries, err := migrationsFS.ReadDir("./migrations")
// 	if err != nil {
// 		return fmt.Errorf("failed to read migrations directory: %w", err)
// 	}

// 	var upFiles []string
// 	for _, entry := range entries {
// 		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
// 			upFiles = append(upFiles, entry.Name())
// 		}
// 	}

// 	sort.Strings(upFiles)

// 	for _, file := range upFiles {
// 		log.Printf("Running migration: %s\n", file)
// 		content, err := migrationsFS.ReadFile(filepath.Join("migrations", file))
// 		if err != nil {
// 			return fmt.Errorf("failed to read migration file %s: %w", file, err)
// 		}

// 		if _, err := db.conn.Exec(string(content)); err != nil {
// 			return fmt.Errorf("failed to execute migration %s: %w", file, err)
// 		}
// 	}

// 	log.Println("Migrations completed successfully")
// 	return nil
// }

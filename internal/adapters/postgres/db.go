package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}

func (db *DB) RunMigrations() error {
	migrationsPath := "./migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
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
		log.Printf("Running migration: %s\n", file)

		fullPath := filepath.Join(migrationsPath, file)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.conn.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	log.Println("Migrations completed successfully")
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

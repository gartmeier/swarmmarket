package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations runs all SQL migrations in order.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create migrations tracking table if not exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort by filename
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Run each migration
	for _, file := range files {
		version := strings.TrimSuffix(file, filepath.Ext(file))

		// Check if already applied
		var count int
		err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("Migration %s already applied, skipping", version)
			continue
		}

		// Read and execute migration
		content, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		log.Printf("Running migration: %s", version)
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		// Record migration
		_, err = pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		log.Printf("Migration %s applied successfully", version)
	}

	return nil
}

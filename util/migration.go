package util

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/aronreisx/bubblebank/db/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunDBMigration runs database migrations using goose
func RunDBMigration(migrationPath string, dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migration: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Configure goose to use the migration directory
	goose.SetBaseFS(nil) // Reset any previous FS setting

	// Run migrations - this will run both SQL and Go migrations
	if err := goose.Up(db, migrationPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("DB migration completed successfully")
	return nil
}

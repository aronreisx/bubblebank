package util

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/aronreisx/bubblebank/db/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunDBMigration runs database migrations using goose with in-code Go migrations
func RunDBMigration(migrationPath string, dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migration: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Since we're using Go-based migrations, we can run migrations directly without
	// needing the physical migration files. The Go files from db/migrations are imported
	// and registered with Goose when the program starts.
	if err := goose.Up(db, ""); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("DB migration completed successfully")
	return nil
}

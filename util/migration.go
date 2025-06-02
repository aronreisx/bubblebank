package util

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/aronreisx/bubblebank/db/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunDBMigration runs database migrations using goose with in-code Go migrations
func RunDBMigration(migrationPath string, dbURL string) error {
	var db *sql.DB
	var err error

	log.Println("Attempting to connect to database for migrations...")

	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := range maxRetries {
		db, err = sql.Open("pgx", dbURL)
		if err != nil {
			log.Printf("Database connection attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		// Check if connection is actually working
		err = db.Ping()
		if err != nil {
			log.Printf("Database ping attempt %d failed: %v", i+1, err)
			if err := db.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			}
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Successfully connected to database on attempt %d", i+1)
		break
	}

	if err != nil {
		log.Printf("Database connection failed after %d attempts: %v", maxRetries, err)
		return fmt.Errorf("database connection unavailable")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if migrationPath == "" || !directoryExists(migrationPath) {
		migrationPath = "."
		log.Printf("Using current directory for migrations as fallback since the migrations path was not provided")
	}

	if err := goose.Up(db, migrationPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("DB migration completed successfully")
	return nil
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

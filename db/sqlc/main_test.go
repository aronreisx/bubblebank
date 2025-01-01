package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aronreisx/bubblebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testQueries  *Queries
	testConnPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatalf("Error during configuration loading: %v", err)
	}

	// Start the PostgreSQL container
	dbContainer, dbURL, err := util.StartPostgresContainer(config)
	if err != nil {
		log.Fatalf("Could not start container: %v", err)
	}
	defer func() {
		if err := dbContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()

	// Run migrations
	if err := util.RunMigrations(dbURL); err != nil {
		log.Fatalf("Could not run migrations: %v", err)
	}

	// Create a connection pool
	testConnPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	// Initialize test queries
	testQueries = New(testConnPool)

	// Run tests
	os.Exit(m.Run())
}

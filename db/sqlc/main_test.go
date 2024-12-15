package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testConnPool *pgxpool.Pool

func init() {
	err := godotenv.Load("../../env/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func TestMain(m *testing.M) {
	var err error

	testConnPool, err = pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	testQueries = New(testConnPool)
	os.Exit(m.Run())
}

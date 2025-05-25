package main

import (
	"context"
	"log"
	"net"

	api "github.com/aronreisx/bubblebank/api"
	db "github.com/aronreisx/bubblebank/db/sqlc"
	"github.com/aronreisx/bubblebank/util"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/aronreisx/bubblebank/db/migrations"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	// Print connection details for debugging
	log.Printf("Connecting to PostgreSQL with Host: '%s', Port: '%s', User: '%s', Database: '%s'",
		config.DBHost, config.DBPort, config.DBUser, config.DBName)

	// Construct connection string with explicit parameters
	connString := util.ConstructDBConnectionString(
		config.DBUser,
		config.DBPass,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	// Run database migrations using in-code Go migrations
	log.Printf("Running database migrations")
	if err := util.RunDBMigration(config.MigrationsFolder, connString); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// Parse the connection string into a pgxpool config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("unable to parse config: %v\n", err)
	}

	// Force TCP connection by setting the dial function
	poolConfig.ConnConfig.Config.DialFunc = (&net.Dialer{}).DialContext

	// Create new connection pool with our custom configuration
	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	server.SetReady()
	log.Println("Server is ready to receive traffic")

	err = server.Start(":" + config.ServerPort)
	if err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}

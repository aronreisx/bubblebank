package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	ctx := context.Background()

	// Initialize telemetry
	telemetryManager, err := util.InitTelemetry(ctx, config)
	if err != nil {
		log.Fatalf("failed to initialize telemetry: %v", err)
	}
	defer telemetryManager.Shutdown()

	// Use structured logger from telemetry
	logger := telemetryManager.Logger

	logger.Info("Starting BubbleBank service",
		"db_host", config.DBHost,
		"db_port", config.DBPort,
		"db_user", config.DBUser,
		"db_name", config.DBName,
	)

	// Construct connection string with explicit parameters
	connString := util.ConstructDBConnectionString(
		config.DBUser,
		config.DBPass,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	// Run database migrations using in-code Go migrations
	logger.Info("Running database migrations")
	if err := util.RunDBMigration(config.MigrationsFolder, connString); err != nil {
		logger.Error("migration failed", "error", err)
		log.Fatalf("migration failed: %v", err)
	}

	// Parse the connection string into a pgxpool config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Error("unable to parse database config", "error", err)
		log.Fatalf("unable to parse config: %v\n", err)
	}

	// Force TCP connection by setting the dial function
	poolConfig.ConnConfig.Config.DialFunc = (&net.Dialer{}).DialContext

	// Create new connection pool with our custom configuration
	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		log.Fatalf("Database service unavailable")
	}

	// Initialize instrumented store
	store := db.NewInstrumentedStore(conn, telemetryManager)
	server := api.NewServer(store, telemetryManager)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		server.SetReady()
		logger.Info("Server is ready to receive traffic", "port", config.ServerPort)

		if err := server.Start(":" + config.ServerPort); err != nil {
			logger.Error("server failed to start", "error", err)
			log.Fatalf("cannot start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down server...")

	// Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}

	conn.Close()
	logger.Info("Server stopped")
}

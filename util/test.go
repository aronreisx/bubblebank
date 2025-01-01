package util

import (
	"context"
	"database/sql"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Required for migrations
	_ "github.com/jackc/pgx/v5/stdlib"                   // Import the pgx driver
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgresContainer starts a PostgreSQL container using testcontainers-go.
func StartPostgresContainer(config Config) (testcontainers.Container, string, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        config.DBImage + ":" + config.DBVersion,
		ExposedPorts: []string{config.DBPort + "/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     config.DBUser,
			"POSTGRES_PASSWORD": config.DBPass,
			"POSTGRES_DB":       config.DBName,
		},
		WaitingFor: wait.ForSQL(nat.Port(config.DBPort+"/tcp"), "pgx", func(host string, port nat.Port) string {
			return ConstructDBUrl(config.DBUser, config.DBPass, host, port.Port(), config.DBName)
		}).WithStartupTimeout(60 * time.Second),
	}
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := dbContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	port, err := dbContainer.MappedPort(ctx, nat.Port(config.DBPort+"/tcp"))
	if err != nil {
		return nil, "", err
	}

	dbURL := ConstructDBUrl(config.DBUser, config.DBPass, host, port.Port(), config.DBName)
	return dbContainer, dbURL, nil
}

// RunMigrations runs database migrations using golang-migrate.
func RunMigrations(dbURL string) error {
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}
	migrateInstance, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migration",
		"postgres", driver)
	if err != nil {
		return err
	}
	if err := migrateInstance.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

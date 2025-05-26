package util

import (
	"context"
	"time"

	_ "github.com/aronreisx/bubblebank/db/migrations"
	"github.com/docker/go-connections/nat"
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
			return ConstructDBConnectionString(config.DBUser, config.DBPass, host, port.Port(), config.DBName)
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

	dbURL := ConstructDBConnectionString(config.DBUser, config.DBPass, host, port.Port(), config.DBName)
	return dbContainer, dbURL, nil
}



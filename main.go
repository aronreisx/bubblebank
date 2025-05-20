package main

import (
	"context"
	"log"

	api "github.com/aronreisx/bubblebank/api"
	db "github.com/aronreisx/bubblebank/db/sqlc"
	"github.com/aronreisx/bubblebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	dbSource := util.ConstructDBUrl(
		config.DBUser,
		config.DBPass,
		"localhost",
		config.DBPort,
		config.DBName,
	)

	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(":" + config.ServerPort)
	if err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}

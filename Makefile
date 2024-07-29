ifndef ENV
	ENV_FILE_PATH = $(CURDIR)/env/.env
	COMPOSE_FILE_PATH = $(CURDIR)/docker/compose.yaml
else
	ENV_FILE_PATH = $(CURDIR)/env/.env.$(ENV)
	COMPOSE_FILE_PATH = $(CURDIR)/docker/compose.$(ENV).yaml
endif

include $(ENV_FILE_PATH)

COMPOSE_BASE_COMMAND=docker-compose -p $(PROJECT_NAME) -f $(COMPOSE_FILE_PATH) --env-file $(ENV_FILE_PATH)

.PHONY: run-database
run-database:
	docker run -d \
	--name $(DB_CONTAINER_NAME) \
	-p $(DB_PORT):$(DB_PORT) \
	-e PGDATA=/var/lib/postgresql/data/pgdata \
	-e POSTGRES_USER=$(DB_USER) \
	-e POSTGRES_PASSWORD=$(DB_PASS) \
	-e POSTGRES_DB=$(DB_NAME) \
	$(DB_TYPE):$(DB_VERSION)

.PHONY: start-database
start-database:
	docker start $(DB_CONTAINER_NAME)

.PHONY: stop-database
stop-database:
	docker stop $(DB_CONTAINER_NAME)

.PHONY: logs-database
logs-database:
	docker logs $(DB_CONTAINER_NAME)

.PHONY: restart-database
restart-database:
	docker restart $(DB_CONTAINER_NAME)

.PHONY: database-bash
database-bash:
	docker exec -it $(DB_CONTAINER_NAME) bash

.PHONY: remove-database
remove-database:
	docker rm -fv $(DB_CONTAINER_NAME)

.PHONY: remove-modules
remove-modules:
	rm -rf node_modules

.PHONY: compose-up
compose-up:
	$(COMPOSE_BASE_COMMAND) up -d

.PHONY: compose-down
compose-down:
	$(COMPOSE_BASE_COMMAND) down

.PHONY: sqlc-migrate
sqlc-migrate:
	migrate -path db/migration -database "$(DB_PROTOCOL)://$(DB_USER):$(DB_PASS)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SQL_MODE)" -verbose up

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

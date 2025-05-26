ENV_FILE_PATH = $(CURDIR)/.env

COMPOSE_BASE_COMMAND = docker-compose -p "$(PROJECT_NAME)" --env-file $(ENV_FILE_PATH)

ifneq ($(wildcard $(ENV_FILE_PATH)),)
	include $(ENV_FILE_PATH)
endif

.PHONY: remove-volumes db-migrate-up db-migrate-down sqlc-generate test-run \
 test-coverage test-coverage-html compose-up compose-down ci-test server mock \
 docker-build db-migrate-status db-migration lint-fix lint-check

remove-volumes:
	rm -rf volumes

compose-up:
	$(COMPOSE_BASE_COMMAND) up -d

compose-down:
	$(COMPOSE_BASE_COMMAND) down

db-migrate-up:
	goose -dir db/migrations postgres "$(DB_URL)" up

db-migrate-down:
	goose -dir db/migrations postgres "$(DB_URL)" down

db-migrate-status:
	goose -dir db/migrations postgres "$(DB_URL)" status

db-migration:
	goose -dir db/migrations create $(name) go

sqlc-generate:
	sqlc generate

ci-test:
	go test ./... -v -race # TODO: Add coverage

test-run:
	go test -v -cover ./...

test-coverage:
	go test -v -cover ./... -coverprofile=./coverage/coverage.log; \
	go tool cover -func=./coverage/coverage.log

test-coverage-html:
	go test -v -cover ./... -coverprofile=./coverage/coverage.log; \
	go tool cover -html=./coverage/coverage.log

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/aronreisx/bubblebank/db/sqlc Store

docker-build:
	docker build -t $(CONTAINER_REGISTRY)/$(PROJECT_NAME):latest .

lint-fix:
	golangci-lint run --fix

lint-check:
	golangci-lint run --verbose

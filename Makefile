ENV_FILE_PATH = $(CURDIR)/.env

COVERAGE_DIR ?= ./coverage
MIGRATIONS_FOLDER ?= db/migrations
IMAGE_TAG ?= $(shell git describe --tags --always --dirty)

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
	goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_URL)" up

db-migrate-down:
	goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_URL)" down

db-migrate-status:
	goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_URL)" status

db-migration:
	goose -dir $(MIGRATIONS_FOLDER) create $(name) go

sqlc-generate:
	sqlc generate

ci-test:
	go test ./... -v -race # TODO: Add coverage

test-run:
	go test -v -cover ./...

$(COVERAGE_DIR):
	mkdir -p $(COVERAGE_DIR)

test-coverage: $(COVERAGE_DIR)
	go test -v -cover ./... -coverprofile=$(COVERAGE_DIR)/coverage.log && \
	go tool cover -func=$(COVERAGE_DIR)/coverage.log

test-coverage-html: test-coverage
	go tool cover -html=$(COVERAGE_DIR)/coverage.log

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/aronreisx/bubblebank/db/sqlc Store

docker-build:
	docker build --cache-from $(CONTAINER_REGISTRY)/$(PROJECT_NAME):latest -t $(CONTAINER_REGISTRY)/$(PROJECT_NAME):latest -t $(CONTAINER_REGISTRY)/$(PROJECT_NAME):$(IMAGE_TAG) .

lint-fix:
	golangci-lint run --fix

lint-check:
	golangci-lint run --verbose

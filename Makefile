ENV_FILE_PATH = $(CURDIR)/.env

COMPOSE_BASE_COMMAND = docker-compose -p "$(PROJECT_NAME)"

ifneq ($(wildcard $(ENV_FILE_PATH)),)
	include $(ENV_FILE_PATH)
endif

.PHONY: remove-volumes db-migrate-up db-migrate-down sqlc-generate test-run \
 test-coverage test-coverage-html compose-up compose-down

remove-volumes:
	rm -rf volumes

compose-up:
	$(COMPOSE_BASE_COMMAND) up -d

compose-down:
	$(COMPOSE_BASE_COMMAND) down

db-migrate-up:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

db-migrate-down:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc-generate:
	sqlc generate

test-run:
	go test -v -cover ./...

test-coverage:
	go test -v -cover ./... -coverprofile=./coverage/coverage.log; \
	go tool cover -func=./coverage/coverage.log

test-coverage-html:
	go test -v -cover ./... -coverprofile=./coverage/coverage.log; \
	go tool cover -html=./coverage/coverage.log

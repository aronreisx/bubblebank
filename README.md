# Bubblebank

Minimal banking API in Go with PostgreSQL, SQLC, and in-code Goose migrations. Production-friendly container image and CI/CD with linting and security scans.

## Features

- Health and readiness: `GET /health`, `GET /ready` (ready after migrations)
- Accounts API:
  - `POST /accounts` create account (`owner`, `currency=USD|EUR`)
  - `GET /accounts/:id` fetch account by id
  - `GET /accounts?page_id=&page_size=` paginated list
- Transactional money transfer implemented in store layer (`TransferTx`)—not exposed via HTTP yet
- Typed data access generated with SQLC (`accounts`, `entries`, `transfers`)

## Quick Start

### Prerequisites

- Go `1.24+` (toolchain pinned via `mise.toml`)
- PostgreSQL (local or via Docker)

### Configure

1. Copy `.env.example` to `.env` and set variables:
   - `DB_IMAGE`, `DB_VERSION`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`
   - `DB_HOST` (e.g., `localhost` or `db` when using compose)
   - `SERVER_PORT` (e.g., `8080`)
   - `MIGRATIONS_FOLDER` (default `db/migrations`)

### Run

- Local server (migrations run automatically):
  ```sh
  make server
  ```
- Optional: start Postgres + pgAdmin via Docker Compose (requires `.env`):
  ```sh
  make compose-up
  ```

## API Summary

- `POST /accounts`
  ```sh
  curl -X POST http://localhost:$SERVER_PORT/accounts \
    -H 'Content-Type: application/json' \
    -d '{"owner":"alice","currency":"USD"}'
  ```
- `GET /accounts/:id`
  ```sh
  curl http://localhost:$SERVER_PORT/accounts/1
  ```
- `GET /accounts?page_id=1&page_size=5`
  ```sh
  curl 'http://localhost:$SERVER_PORT/accounts?page_id=1&page_size=5'
  ```
- `GET /health` and `GET /ready`

## Development

- Generate SQLC code: `make sqlc-generate`
- Create migration: `make db-migration name=<migration_name>`
- Apply migrations (CLI): `make db-migrate-up` / `make db-migrate-down`
- Lint: `make lint-check` (fix with `make lint-fix`)
- Tests: `make test-run`

## Architecture

- `api/` Gin HTTP server and handlers
- `db/sqlc/` SQLC-generated queries, models, and transactional store (`TransferTx`)
- `db/migrations/` in-code Goose migrations
- `util/` configuration, migrations runner, random test data, testcontainers helpers

## Non-Functional

- Readiness gating via `SetReady()` after migrations
- Distroless, non-root runtime image
- Static analysis: `golangci-lint` (`errcheck`, `goimports`, `govet`, `staticcheck`)
- Security: `gosec` for code, `trivy` for images
- Tests use `testcontainers` (requires proper DB env variables)

## License

MIT — see [LICENSE](LICENSE)

## Contact

- Email: `aronreis2@gmail.com`
- Issues: https://github.com/aronreisx/bubblebank/issues

Note: Transfers and entries are implemented in the store layer but have no HTTP endpoints yet.

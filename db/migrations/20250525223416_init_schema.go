package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitSchema, downInitSchema)
}

func upInitSchema(ctx context.Context, tx *sql.Tx) error {
	// Create accounts table
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "accounts" (
		  "id" bigserial PRIMARY KEY,
		  "owner" varchar NOT NULL,
		  "balance" bigint NOT NULL,
		  "currency" varchar NOT NULL,
		  "created_at" timestamptz NOT NULL DEFAULT (now())
		);
	`)
	if err != nil {
		return err
	}

	// Create entries table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "entries" (
		  "id" bigserial PRIMARY KEY,
		  "account_id" bigint NOT NULL,
		  "amount" bigint NOT NULL,
		  "created_at" timestamptz NOT NULL DEFAULT (now())
		);
	`)
	if err != nil {
		return err
	}

	// Create transfers table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "transfers" (
		  "id" bigserial PRIMARY KEY,
		  "from_account_id" bigint NOT NULL,
		  "to_account_id" bigint NOT NULL,
		  "amount" bigint NOT NULL,
		  "created_at" timestamptz NOT NULL DEFAULT (now())
		);
	`)
	if err != nil {
		return err
	}

	// Add foreign keys
	_, err = tx.ExecContext(ctx, `
		ALTER TABLE IF EXISTS "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");
		ALTER TABLE IF EXISTS "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
		ALTER TABLE IF EXISTS "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
	`)
	if err != nil {
		return err
	}

	// Create indexes
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX ON "accounts" ("owner");
		CREATE INDEX ON "entries" ("account_id");
		CREATE INDEX ON "transfers" ("from_account_id");
		CREATE INDEX ON "transfers" ("to_account_id");
		CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");
	`)
	if err != nil {
		return err
	}

	// Add comments
	_, err = tx.ExecContext(ctx, `
		COMMENT ON COLUMN "entries"."amount" IS 'Can be negative or positive';
		COMMENT ON COLUMN "transfers"."amount" IS 'Must be positive';
	`)
	if err != nil {
		return err
	}

	return nil
}

func downInitSchema(ctx context.Context, tx *sql.Tx) error {
	// Drop tables in reverse order of creation
	_, err := tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS entries;
		DROP TABLE IF EXISTS transfers;
		DROP TABLE IF EXISTS accounts;
	`)
	return err
}

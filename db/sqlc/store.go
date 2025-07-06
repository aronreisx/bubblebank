package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

// InstrumentedStore wraps SQLStore with telemetry
type InstrumentedStore struct {
	*SQLStore
	tracer trace.Tracer
}

// NewInstrumentedStore creates a new instrumented store with telemetry
func NewInstrumentedStore(connPool *pgxpool.Pool, telemetryManager interface{}) Store {
	sqlStore := &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}

	return &InstrumentedStore{
		SQLStore: sqlStore,
		tracer:   otel.Tracer("bubblebank-db"),
	}
}

// CreateAccount creates a new account with tracing
func (s *InstrumentedStore) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	ctx, span := s.tracer.Start(ctx, "CreateAccount")
	defer span.End()

	span.SetAttributes(
		attribute.String("account.owner", arg.Owner),
		attribute.String("account.currency", arg.Currency),
		attribute.Int64("account.balance", arg.Balance),
	)

	start := time.Now()
	account, err := s.SQLStore.CreateAccount(ctx, arg)
	duration := time.Since(start)

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		span.RecordError(err)
	} else {
		span.SetAttributes(
			attribute.Int64("account.id", account.ID),
			attribute.String("account.created_at", account.CreatedAt.String()),
		)
	}

	span.SetAttributes(attribute.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))

	return account, err
}

// GetAccount gets an account by ID with tracing
func (s *InstrumentedStore) GetAccount(ctx context.Context, id int64) (Account, error) {
	ctx, span := s.tracer.Start(ctx, "GetAccount")
	defer span.End()

	span.SetAttributes(attribute.Int64("account.id", id))

	start := time.Now()
	account, err := s.SQLStore.GetAccount(ctx, id)
	duration := time.Since(start)

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		span.RecordError(err)
	} else {
		span.SetAttributes(
			attribute.String("account.owner", account.Owner),
			attribute.String("account.currency", account.Currency),
			attribute.Int64("account.balance", account.Balance),
		)
	}

	span.SetAttributes(attribute.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))

	return account, err
}

// ListAccounts lists accounts with tracing
func (s *InstrumentedStore) ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error) {
	ctx, span := s.tracer.Start(ctx, "ListAccounts")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("limit", int64(arg.Limit)),
		attribute.Int64("offset", int64(arg.Offset)),
	)

	start := time.Now()
	accounts, err := s.SQLStore.ListAccounts(ctx, arg)
	duration := time.Since(start)

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		span.RecordError(err)
	} else {
		span.SetAttributes(attribute.Int("accounts_count", len(accounts)))
	}

	span.SetAttributes(attribute.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))

	return accounts, err
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		/* if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		} */
		return err
	}

	return tx.Commit(ctx)
}

// TransferTx performs a money transfer from one account to the other with tracing
func (store *InstrumentedStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	ctx, span := store.tracer.Start(ctx, "TransferTx")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("transfer.from_account_id", arg.FromAccountID),
		attribute.Int64("transfer.to_account_id", arg.ToAccountID),
		attribute.Int64("transfer.amount", arg.Amount),
	)

	start := time.Now()
	result, err := store.SQLStore.TransferTx(ctx, arg)
	duration := time.Since(start)

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		span.RecordError(err)
	} else {
		span.SetAttributes(
			attribute.Int64("transfer.id", result.Transfer.ID),
			attribute.Int64("from_account.final_balance", result.FromAccount.Balance),
			attribute.Int64("to_account.final_balance", result.ToAccount.Balance),
		)
	}

	span.SetAttributes(attribute.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))

	return result, err
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

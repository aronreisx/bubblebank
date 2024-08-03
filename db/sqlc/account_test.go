package db

import (
	"context"
	"testing"

	"github.com/aronreisx/bubblebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	newAccount := createRandomAccount(t)

	retrievedAccount, err := testQueries.GetAccount(
		context.Background(),
		newAccount.ID,
	)

	require.NoError(t, err)
	require.NotEmpty(t, retrievedAccount)

	require.Equal(t, newAccount.ID, retrievedAccount.ID)
	require.Equal(t, newAccount.Owner, retrievedAccount.Owner)
	require.Equal(t, newAccount.Balance, retrievedAccount.Balance)
	require.Equal(t, newAccount.Currency, retrievedAccount.Currency)

	require.WithinDuration(
		t,
		newAccount.CreatedAt,
		retrievedAccount.CreatedAt,
		time.Second,
	)
}

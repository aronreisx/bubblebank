package db

import (
	"context"
	"testing"
	"time"

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

func TestUpdateAccount(t *testing.T) {
	createdAccount := createRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      createdAccount.ID,
		Balance: util.RandomMoney(),
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)

	require.Equal(t, createdAccount.ID, updatedAccount.ID)
	require.Equal(t, createdAccount.Owner, updatedAccount.Owner)
	require.Equal(t, arg.Balance, updatedAccount.Balance)
	require.Equal(t, createdAccount.Currency, updatedAccount.Currency)

	require.WithinDuration(
		t,
		createdAccount.CreatedAt,
		updatedAccount.CreatedAt,
		time.Second,
	)
}

func TestDeleteAccount(t *testing.T) {
	createdAccount := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), createdAccount.ID)
	require.NoError(t, err)

	unavailableAccount, err := testQueries.GetAccount(
		context.Background(),
		createdAccount.ID,
	)

	require.Error(t, err)
	require.Empty(t, unavailableAccount)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, accounts, int(arg.Limit))

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}

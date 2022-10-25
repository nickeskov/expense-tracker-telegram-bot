package inmemory

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

type repoFn func(t *testing.T) *Repository

var defaultUser = models.NewUser(1, "test")

func newRepo(t *testing.T) *Repository {
	r, err := New()
	require.NoError(t, err)
	return r
}

func newRepoWithUser(t *testing.T, u models.User) *Repository {
	r, err := New()
	require.NoError(t, err)
	r.storage[u.ID] = u
	return r
}

func TestRepository_CreateUser(t *testing.T) {
	ctx := context.Background()
	u := defaultUser

	tests := []struct {
		repoFn      repoFn
		expectedErr error
	}{
		{
			repoFn:      newRepo,
			expectedErr: nil,
		},
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			expectedErr: user.ErrAlreadyExists,
		},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		createdU, err := r.CreateUser(ctx, u)
		if test.expectedErr != nil {
			require.ErrorIs(t, err, user.ErrAlreadyExists)
		} else {
			require.NoError(t, err)
			require.Equal(t, u, createdU)
			require.Equal(t, u, r.storage[u.ID])

		}
	}
}

func TestRepository_IsUserExists(t *testing.T) {
	ctx := context.Background()
	u := defaultUser

	tests := []struct {
		exists bool
		repoFn repoFn
	}{
		{exists: false, repoFn: newRepo},
		{exists: true, repoFn: func(t *testing.T) *Repository { return newRepoWithUser(t, u) }},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		exists, err := r.IsUserExists(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, test.exists, exists)
	}
}

func TestRepository_ChangeUserCurrency(t *testing.T) {
	ctx := context.Background()
	u := defaultUser
	newCurr := models.CurrencyCode("new")

	tests := []struct {
		repoFn      repoFn
		curr        models.CurrencyCode
		expectedErr error
	}{
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			curr:        newCurr,
			expectedErr: nil,
		},
		{
			repoFn:      newRepo,
			curr:        "",
			expectedErr: user.ErrDoesNotExist,
		},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		err := r.ChangeUserCurrency(ctx, u.ID, test.curr)
		if test.expectedErr != nil {
			require.ErrorIs(t, err, test.expectedErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.curr, r.storage[u.ID].SelectedCurrency)
		}
	}
}

func TestRepository_GetUserCurrency(t *testing.T) {
	ctx := context.Background()
	u := defaultUser

	tests := []struct {
		repoFn      repoFn
		expectedErr error
	}{
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			expectedErr: nil,
		},
		{
			repoFn:      newRepo,
			expectedErr: user.ErrDoesNotExist,
		},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		storedCurr, err := r.GetUserCurrency(ctx, u.ID)
		if test.expectedErr != nil {
			require.Equal(t, err, test.expectedErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, u.SelectedCurrency, storedCurr)
		}
	}
}

func TestRepository_GetUserMonthlyLimit(t *testing.T) {
	ctx := context.Background()
	u := defaultUser
	expectedLimit := decimal.NewFromInt(42)

	tests := []struct {
		repoFn      repoFn
		limit       *decimal.Decimal
		expectedErr error
	}{
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			limit:       nil,
			expectedErr: nil,
		},
		{
			repoFn: func(t *testing.T) *Repository {
				u := u
				u.MonthlyLimit = &expectedLimit
				return newRepoWithUser(t, u)
			},
			limit:       &expectedLimit,
			expectedErr: nil,
		},
		{
			repoFn:      newRepo,
			limit:       nil,
			expectedErr: user.ErrDoesNotExist,
		},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		limit, err := r.GetUserMonthlyLimit(ctx, u.ID)
		if test.expectedErr != nil {
			require.Equal(t, err, test.expectedErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, test.limit, limit)
		}
	}
}

func TestRepository_SetUserMonthlyLimit(t *testing.T) {
	ctx := context.Background()
	u := defaultUser
	expectedLimit := decimal.NewFromInt(42)

	tests := []struct {
		repoFn      repoFn
		limit       *decimal.Decimal
		expectedErr error
	}{
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			limit:       nil,
			expectedErr: nil,
		},
		{
			repoFn:      func(t *testing.T) *Repository { return newRepoWithUser(t, u) },
			limit:       &expectedLimit,
			expectedErr: nil,
		},
		{
			repoFn:      newRepo,
			limit:       nil,
			expectedErr: user.ErrDoesNotExist,
		},
	}
	for _, test := range tests {
		r := test.repoFn(t)
		err := r.SetUserMonthlyLimit(ctx, u.ID, test.limit)
		if test.expectedErr != nil {
			require.Equal(t, err, test.expectedErr)
		} else {
			require.NoError(t, err)
			limit, err := r.GetUserMonthlyLimit(ctx, u.ID)
			require.NoError(t, err)
			require.Equal(t, test.limit, limit)
		}
	}
}

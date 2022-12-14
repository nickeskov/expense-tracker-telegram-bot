package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	expenseInMemRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/inmemory"
	exrateInMemRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/inmemory"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	userInMemRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/repository/inmemory"
)

func newUC(t *testing.T, baseCurrency models.CurrencyCode, u models.User, rates ...models.ExchangeRate) *UseCase {
	ctx := context.TODO()

	expRepo, err := expenseInMemRepo.New()
	require.NoError(t, err)

	userRepo, err := userInMemRepo.New()
	require.NoError(t, err)
	_, err = userRepo.CreateUser(ctx, u)
	require.NoError(t, err)

	ratesRepo, err := exrateInMemRepo.New()
	require.NoError(t, err)
	err = ratesRepo.AddOrUpdateRates(ctx, rates...)
	require.NoError(t, err)

	uc, err := New(baseCurrency, expRepo, userRepo, ratesRepo)
	require.NoError(t, err)
	return uc
}

func TestUseCase_ExpensesSummaryByCategorySince(t *testing.T) {
	const (
		userID       = models.UserID(10)
		baseCurr     = models.CurrencyCode("RUB")
		selectedCurr = models.CurrencyCode("USD")
	)
	var (
		fromSelectedToBaseRate = decimal.NewFromInt(50)
		fromBaseToSelectedRate = decimal.NewFromInt(1).Div(fromSelectedToBaseRate)
	)
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	timeDelta := 32 * time.Hour

	var (
		user  = models.NewUser(userID, selectedCurr)
		rates = []models.ExchangeRate{
			models.NewExchangeRate(selectedCurr, fromBaseToSelectedRate, midnight),
			models.NewExchangeRate(selectedCurr, fromBaseToSelectedRate, midnight.Add(timeDelta/2)),
			models.NewExchangeRate(selectedCurr, fromBaseToSelectedRate, midnight.Add(timeDelta)),
		}
		expenses = []models.Expense{
			{
				ID:       1,
				Category: "cat1",
				Amount:   decimal.NewFromInt(111),
				Date:     midnight,
				Comment:  "comment",
			},
			{
				ID:       2,
				Category: "cat1",
				Amount:   decimal.NewFromInt(222),
				Date:     midnight.Add(timeDelta / 2),
				Comment:  "comment",
			},
			{
				ID:       3,
				Category: "cat2",
				Amount:   decimal.NewFromInt(999),
				Date:     midnight.Add(timeDelta),
				Comment:  "comment",
			},
		}
	)

	tests := []struct {
		since               time.Time
		till                time.Time
		expenses            []models.Expense
		selectedCurr        models.CurrencyCode
		summaryByCategories expense.SummaryReport
	}{
		{expenses: nil, summaryByCategories: expense.SummaryReport{}},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta),
			expenses:            expenses,
			selectedCurr:        selectedCurr,
			summaryByCategories: expense.SummaryReport{"cat1": decimal.NewFromInt(333), "cat2": decimal.NewFromInt(999)},
		},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta / 2),
			expenses:            expenses,
			selectedCurr:        baseCurr,
			summaryByCategories: expense.SummaryReport{"cat1": decimal.NewFromInt(333).Mul(fromSelectedToBaseRate)},
		},
		{
			since:               midnight.Add(timeDelta),
			till:                midnight,
			expenses:            expenses,
			summaryByCategories: expense.SummaryReport{},
		},
	}
	for i, test := range tests {
		testCase := test
		ctx := context.Background()
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			uc := newUC(t, baseCurr, user, rates...)
			for _, exp := range testCase.expenses {
				_, err := uc.AddExpense(ctx, userID, exp)
				require.NoError(t, err)
			}
			err := uc.userRepo.ChangeUserCurrency(ctx, userID, testCase.selectedCurr)
			require.NoError(t, err)
			summary, err := uc.GetExpensesSummaryByCategorySince(ctx, userID, testCase.since, testCase.till)
			require.NoError(t, err)
			for category, d := range summary {
				expected := testCase.summaryByCategories[category]
				actual := d
				assert.Truef(t, expected.Equal(actual), "category %q: want %v, got %v", category, expected, actual)
			}
		})
	}
}

package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	expenseInMemRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/inmemory"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

func newUC(t *testing.T) *UseCase {
	repo, err := expenseInMemRepo.New()
	require.NoError(t, err)
	uc, err := New(repo)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, repo.Close())
	})
	return uc
}

func TestUseCase_ExpensesSummaryByCategorySince(t *testing.T) {
	const userID = models.UserID(10)
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	timeDelta := 32 * time.Hour
	expenses := []models.Expense{
		{
			ID:       1,
			Category: "cat1",
			Amount:   111,
			Date:     midnight,
			Comment:  "comment",
		},
		{
			ID:       2,
			Category: "cat1",
			Amount:   222,
			Date:     midnight.Add(timeDelta / 2),
			Comment:  "comment",
		},
		{
			ID:       3,
			Category: "cat2",
			Amount:   999,
			Date:     midnight.Add(timeDelta),
			Comment:  "comment",
		},
	}
	tests := []struct {
		since               time.Time
		till                time.Time
		expenses            []models.Expense
		summaryByCategories expense.SummaryReport
	}{
		{expenses: nil, summaryByCategories: expense.SummaryReport{}},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta),
			expenses:            expenses,
			summaryByCategories: expense.SummaryReport{"cat1": 333, "cat2": 999},
		},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta / 2),
			expenses:            expenses,
			summaryByCategories: expense.SummaryReport{"cat1": 333},
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
			uc := newUC(t)
			for _, exp := range testCase.expenses {
				_, err := uc.AddExpense(ctx, userID, exp)
				require.NoError(t, err)
			}
			summary, err := uc.ExpensesSummaryByCategorySince(ctx, userID, testCase.since, testCase.till)
			require.NoError(t, err)
			require.Equal(t, testCase.summaryByCategories, summary)
		})
	}
}

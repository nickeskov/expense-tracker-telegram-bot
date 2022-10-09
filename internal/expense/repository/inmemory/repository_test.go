package inmemory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

func newRepo(t *testing.T) *Repository {
	r, err := New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Close())
	})
	return r
}

func Test_ExpenseRoundTrip(t *testing.T) {
	const userID = models.UserID(10)
	ctx := context.Background()

	r := newRepo(t)
	expected := models.Expense{
		ID:       1,
		Category: "test",
		Amount:   42,
		Date:     time.Now(),
		Comment:  "test comment",
	}
	_, err := r.AddExpense(ctx, userID, expected)
	require.NoError(t, err)

	actual, err := r.GetExpense(ctx, userID, expected.ID)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	expensesByDate, err := r.ExpensesByDate(ctx, userID, expected.Date)
	require.NoError(t, err)
	require.Len(t, expensesByDate, 1)
	require.Equal(t, expected, expensesByDate[0])
}

func TestRepository_ExpensesAscendSinceTill(t *testing.T) {
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
		since            time.Time
		till             time.Time
		expenses         []models.Expense
		expectedExpenses []models.Expense
	}{
		{expenses: nil, expectedExpenses: nil},
		{
			since:            midnight,
			till:             midnight.Add(timeDelta),
			expenses:         expenses,
			expectedExpenses: expenses,
		},
		{
			since:            midnight,
			till:             midnight.Add(timeDelta / 2),
			expenses:         expenses,
			expectedExpenses: expenses[:2],
		},
		{
			since:            midnight.Add(timeDelta),
			till:             midnight,
			expenses:         expenses,
			expectedExpenses: nil,
		},
	}
	for i, test := range tests {
		testCase := test
		ctx := context.Background()
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			r := newRepo(t)
			for _, expense := range testCase.expenses {
				_, err := r.AddExpense(ctx, userID, expense)
				require.NoError(t, err)
			}
			var actualExpenses []models.Expense
			err := r.ExpensesAscendSinceTill(ctx, userID, testCase.since, testCase.till, func(e *models.Expense) bool {
				actualExpenses = append(actualExpenses, *e)
				return true
			})
			require.NoError(t, err)
			require.Equal(t, testCase.expectedExpenses, actualExpenses)
		})
	}

}

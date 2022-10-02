package inmemory

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

func newRepo(t *testing.T) *Repository {
	r := New()
	t.Cleanup(func() {
		require.NoError(t, r.Close())
	})
	return r
}

func Test_ExpenseRoundTrip(t *testing.T) {
	const userID = models.UserID(10)

	r := newRepo(t)
	expected := models.Expense{
		ID:       1,
		Category: "test",
		Amount:   42,
		Date:     time.Now(),
		Comment:  "test comment",
	}
	_, err := r.AddExpense(userID, expected)
	require.NoError(t, err)

	actual, err := r.GetExpense(userID, expected.ID)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	expensesByDate, err := r.ExpensesByDate(userID, expected.Date)
	require.NoError(t, err)
	require.Len(t, expensesByDate, 1)
	require.Equal(t, expected, expensesByDate[0])
}

func TestRepository_ExpensesSummaryByCategorySince(t *testing.T) {
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
		summaryByCategories map[models.ExpenseCategory]float64
	}{
		{expenses: nil, summaryByCategories: map[models.ExpenseCategory]float64{}},
		{expenses: []models.Expense{}, summaryByCategories: map[models.ExpenseCategory]float64{}},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta),
			expenses:            expenses,
			summaryByCategories: map[models.ExpenseCategory]float64{"cat1": 333, "cat2": 999},
		},
		{
			since:               midnight,
			till:                midnight.Add(timeDelta / 2),
			expenses:            expenses,
			summaryByCategories: map[models.ExpenseCategory]float64{"cat1": 333},
		},
		{
			since:               midnight.Add(timeDelta),
			till:                midnight,
			expenses:            expenses,
			summaryByCategories: map[models.ExpenseCategory]float64{},
		},
	}
	for i, test := range tests {
		testCase := test
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			r := newRepo(t)
			for _, expense := range testCase.expenses {
				_, err := r.AddExpense(userID, expense)
				require.NoError(t, err)
			}
			summary, err := r.ExpensesSummaryByCategorySince(userID, testCase.since, testCase.till)
			require.NoError(t, err)
			require.Equal(t, testCase.summaryByCategories, summary)
		})
	}
}

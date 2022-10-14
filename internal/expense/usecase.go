package expense

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type SummaryReport map[models.ExpenseCategory]decimal.Decimal

func (r SummaryReport) Text() (string, error) {
	if len(r) == 0 {
		return "", nil
	}
	sortedKeys := make([]string, 0, len(r))
	for category := range r {
		sortedKeys = append(sortedKeys, string(category))
	}
	sort.Strings(sortedKeys)

	sb := new(strings.Builder)
	for _, key := range sortedKeys {
		category := models.ExpenseCategory(key)
		_, err := fmt.Fprintf(sb, "%s=%v\n", category, r[category])
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

type UseCase interface {
	AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error)
	ExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (SummaryReport, error)
	ExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error)
}

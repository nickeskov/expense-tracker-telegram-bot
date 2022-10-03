package expense

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type SummaryReport map[models.ExpenseCategory]float64

func (r SummaryReport) Text() string {
	if len(r) == 0 {
		return ""
	}
	sortedKeys := make([]string, 0, len(r))
	for category := range r {
		sortedKeys = append(sortedKeys, string(category))
	}
	sort.Strings(sortedKeys)

	sb := new(strings.Builder)
	for _, key := range sortedKeys {
		category := models.ExpenseCategory(key)
		_, err := fmt.Fprintf(sb, "%s=%f\n", category, r[category])
		if err != nil { // panic here because it's an impossible situation
			panic(err.Error())
		}
	}
	return sb.String()
}

type UseCase interface {
	AddExpense(userID models.UserID, expense models.Expense) (models.Expense, error)
	ExpensesSummaryByCategorySince(userID models.UserID, since, till time.Time) (SummaryReport, error)
	ExpensesAscendSinceTill(userID models.UserID, since, till time.Time, max int) ([]models.Expense, error)
}

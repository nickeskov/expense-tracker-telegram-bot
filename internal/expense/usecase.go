package expense

import (
	"time"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type UseCase interface {
	AddExpense(userID models.UserID, expense models.Expense) (models.Expense, error)
	ExpensesSummaryByCategorySince(userID models.UserID, since, till time.Time) (map[models.ExpenseCategory]float64, error)
	Close() error
}

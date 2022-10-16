package expense

import (
	"context"
	"time"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type Repository interface {
	AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error)
	GetExpensesByDate(ctx context.Context, userID models.UserID, date time.Time) ([]models.Expense, error)
	GetExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, iter func(expense *models.Expense) bool) error
}

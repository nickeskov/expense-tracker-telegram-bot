package expense

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

var (
	ErrExpenseDoesNotExist = errors.New("expense does not exist")
)

type Repository interface {
	AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error)
	GetExpense(ctx context.Context, userID models.UserID, expenseID models.ExpenseID) (models.Expense, error)
	ExpensesByDate(ctx context.Context, userID models.UserID, date time.Time) ([]models.Expense, error)
	ExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, iter func(expense *models.Expense) bool) error
}

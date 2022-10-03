package expense

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

var (
	ErrExpenseDoesNotExist = errors.New("expense does not exist")
)

type Repository interface {
	AddExpense(userID models.UserID, expense models.Expense) (models.Expense, error)
	GetExpense(userID models.UserID, expenseID models.ExpenseID) (models.Expense, error)
	ExpensesByDate(userID models.UserID, date time.Time) ([]models.Expense, error)
	ExpensesAscendSinceTill(userID models.UserID, since, till time.Time, iter func(expense *models.Expense) bool) error
	Close() error
}

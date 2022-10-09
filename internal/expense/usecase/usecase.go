package usecase

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type UseCase struct {
	repo expense.Repository
}

func New(repo expense.Repository) (*UseCase, error) {
	return &UseCase{repo: repo}, nil
}

func (u *UseCase) AddExpense(ctx context.Context, id models.UserID, e models.Expense) (models.Expense, error) {
	return u.repo.AddExpense(ctx, id, e)
}

func (u *UseCase) ExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, error) {
	out := make(expense.SummaryReport)
	err := u.repo.ExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		out[expense.Category] += expense.Amount
		return true
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to iterate through expenses of userID=%d ans split by categories", userID)
	}
	return out, nil
}

func (u *UseCase) ExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error) {
	var out []models.Expense
	err := u.repo.ExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		out = append(out, *expense)
		return len(out) < max
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to iterate through expenses of userID=%d", userID)
	}
	return out, nil
}

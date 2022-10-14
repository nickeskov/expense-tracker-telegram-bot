package usecase

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

type UseCase struct {
	baseCurrency models.CurrencyCode
	expRepo      expense.Repository
	userRepo     user.Repository
	exrateRepo   exrate.Repository
}

func New(baseCurrency models.CurrencyCode, expRepo expense.Repository, userRepo user.Repository, exrateRepo exrate.Repository) (*UseCase, error) {
	return &UseCase{baseCurrency: baseCurrency, expRepo: expRepo, userRepo: userRepo, exrateRepo: exrateRepo}, nil
}

func (u *UseCase) AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error) {
	curr, err := u.userRepo.GetUserCurrency(ctx, userID)
	if err != nil {
		return models.Expense{}, errors.Wrapf(err, "failed to get selected user currency by userID=%d", userID)
	}
	if curr != u.baseCurrency {
		rate, err := u.exrateRepo.GetRate(ctx, curr, expense.Date)
		if err != nil {
			return models.Expense{}, errors.Wrapf(err, "failed to get exchange rate for currency=%q at time=%v", curr, expense.Date)
		}
		expense.Amount = rate.ConvertToBase(expense.Amount)
	}
	return u.expRepo.AddExpense(ctx, userID, expense)
}

func (u *UseCase) GetExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, error) {
	out := make(expense.SummaryReport)
	err := u.handleExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		categoryAmount := out[expense.Category]
		out[expense.Category] = categoryAmount.Add(expense.Amount)
		return true
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to iterate through expenses of userID=%d and split by categories", userID)
	}
	return out, nil
}

func (u *UseCase) GetExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error) {
	var out []models.Expense
	err := u.handleExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		out = append(out, *expense)
		return len(out) < max
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to iterate through expenses of userID=%d", userID)
	}
	return out, nil
}

func (u *UseCase) handleExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, handler func(expense *models.Expense) bool) error {
	curr, err := u.userRepo.GetUserCurrency(ctx, userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get selected user currency by userID=%d", userID)
	}
	var (
		iterErr error
		iter    = handler
	)
	if curr != u.baseCurrency {
		iter = func(expense *models.Expense) bool {
			rate, err := u.exrateRepo.GetRate(ctx, curr, expense.Date)
			if err != nil {
				iterErr = errors.Wrapf(err, "failed to get exchange rate for currency=%q at time=%v", curr, expense.Date)
				return false
			}
			exp := *expense
			exp.Amount = rate.ConvertFromBase(exp.Amount)
			return handler(&exp)
		}
	}
	err = u.expRepo.GetExpensesAscendSinceTill(ctx, userID, since, till, iter)
	if err != nil {
		return err
	}
	if iterErr != nil {
		return iterErr
	}
	return nil
}

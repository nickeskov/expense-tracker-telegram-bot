package usecase

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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

func (u *UseCase) AddExpense(ctx context.Context, userID models.UserID, exp models.Expense) (models.Expense, error) {
	if err := exp.Validate(); err != nil {
		return models.Expense{}, errors.Wrap(err, "expense validation failed")
	}
	curr, err := u.userRepo.GetUserCurrency(ctx, userID)
	if err != nil {
		return models.Expense{}, errors.Wrapf(err, "failed to get selected user currency by userID=%d", userID)
	}
	if curr != u.baseCurrency {
		rate, err := u.exrateRepo.GetRate(ctx, curr, exp.Date)
		if err != nil {
			return models.Expense{}, errors.Wrapf(err, "failed to get exchange rate for currency=%q at time=%v", curr, exp.Date)
		}
		exp.Amount = rate.ConvertToBase(exp.Amount)
	}
	nowYear, nowMonth, _ := time.Now().UTC().Date()
	expenseYear, expenseMonth, _ := exp.Date.UTC().Date()
	// we don't check limit if expense month is not current month
	if expenseMonth != nowMonth || expenseYear != nowYear {
		return u.expRepo.AddExpense(ctx, userID, exp)
	}
	// expense happened in the current month
	var out models.Expense
	err = u.expRepo.Isolated(ctx, func(ctx context.Context) error {
		limit, err := u.userRepo.GetUserMonthlyLimit(ctx, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get user montly limit by userID=%q", userID)
		}
		if limit != nil {
			spentByMonth, err := u.getUserExpensesSumByMonth(ctx, userID, expenseYear, expenseMonth)
			if err != nil {
				return errors.Wrap(err, "failed to get user expenses sum by month")
			}
			newSum := spentByMonth.Add(exp.Amount)
			if newSum.GreaterThan(*limit) {
				return expense.ErrExpensesMonthlyLimitExcess
			}
		}
		out, err = u.expRepo.AddExpense(ctx, userID, exp)
		if err != nil {
			return errors.Wrap(err, "failed to add expense to expenses repository")
		}
		return nil
	})
	if err != nil {
		return models.Expense{}, errors.Wrapf(err, "error occured in expenses repo isolated environment")
	}
	return out, nil
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

func (u *UseCase) getUserExpensesSumByMonth(ctx context.Context, userID models.UserID, year int, month time.Month) (decimal.Decimal, error) {
	var (
		since = time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
		till  = since.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
		sum   decimal.Decimal
	)
	err := u.handleExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		sum = sum.Add(expense.Amount)
		return true
	})
	if err != nil {
		return decimal.Decimal{}, errors.Wrapf(err, "failed to get userID=%d expenses sum by %s", userID, since.Format("2006/January"))
	}
	return sum, nil
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

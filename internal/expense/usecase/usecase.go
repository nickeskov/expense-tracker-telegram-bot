package usecase

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

const (
	userIDSpanTagKey            = "user_id"
	sinceUnixMillisSpanTagKey   = "since_unix_ms"
	tillUnixMillisSpanTagKey    = "till_unix_ms"
	dateUnixMillisSpanTagKey    = "date_unix_ms"
	handlerCallsCountSpanTagKey = "handler_call_count"
	currencyCodeSpanTagKey      = "currency_code"
)

type UseCase struct {
	baseCurrency models.CurrencyCode
	expRepo      expense.Repository
	userRepo     user.Repository
	exrateRepo   exrate.Repository
	reportsCache expense.ReportsCache
}

func New(baseCurrency models.CurrencyCode, expRepo expense.Repository, userRepo user.Repository, exrateRepo exrate.Repository) (*UseCase, error) {
	return NewWithCache(baseCurrency, expRepo, userRepo, exrateRepo, nil)
}

func NewWithCache(
	baseCurrency models.CurrencyCode,
	expRepo expense.Repository, userRepo user.Repository, exrateRepo exrate.Repository,
	reportsCache expense.ReportsCache,
) (*UseCase, error) {
	if reportsCache == nil {
		reportsCache = &noopCache{}
	}
	return &UseCase{
		baseCurrency: baseCurrency,
		expRepo:      expRepo,
		userRepo:     userRepo,
		exrateRepo:   exrateRepo,
		reportsCache: reportsCache,
	}, nil
}

func (u *UseCase) AddExpense(ctx context.Context, userID models.UserID, exp models.Expense) (_ models.Expense, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AddExpense")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, userID)
	defer func() {
		if err != nil {
			return
		}
		err = u.reportsCache.DropCacheForUserID(ctx, userID)
	}()

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
	err = u.expRepo.Isolated(ctx, func(ctx context.Context) (err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "expRepo.Isolated")
		defer func() {
			ext.Error.Set(span, err != nil)
			span.Finish()
		}()
		span.SetTag(userIDSpanTagKey, userID)

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
	cached, ok, err := u.reportsCache.GetFromCache(ctx, userID, since, till)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get report from cache")
	}
	if ok {
		return cached, nil
	}
	out := make(expense.SummaryReport)
	err = u.handleExpensesAscendSinceTill(ctx, userID, since, till, func(expense *models.Expense) bool {
		categoryAmount := out[expense.Category]
		out[expense.Category] = categoryAmount.Add(expense.Amount)
		return true
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to iterate through expenses of userID=%d and split by categories", userID)
	}
	if err := u.reportsCache.AddToCache(ctx, userID, since, till, out); err != nil {
		return nil, errors.Wrap(err, "failed to set report to cache")
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

func (u *UseCase) handleExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, handler func(expense *models.Expense) bool) (err error) {
	var handlerCalls int
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleExpensesAscendSinceTill")
	defer func() {
		span.SetTag(handlerCallsCountSpanTagKey, handlerCalls)
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, userID)
	span.SetTag(sinceUnixMillisSpanTagKey, since.UnixMilli())
	span.SetTag(tillUnixMillisSpanTagKey, till.UnixMilli())

	curr, err := u.userRepo.GetUserCurrency(ctx, userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get selected user currency by userID=%d", userID)
	}
	var iterErr error
	iter := func(expense *models.Expense) bool {
		handlerCalls++
		return handler(expense)
	}
	if curr != u.baseCurrency {
		inner := iter
		iter = func(expense *models.Expense) bool {
			span, ctx := opentracing.StartSpanFromContext(ctx, "exrateRepo.GetRate")
			defer func() {
				ext.Error.Set(span, iterErr != nil)
				span.Finish()
			}()
			span.SetTag(dateUnixMillisSpanTagKey, expense.Date.UnixMilli())
			span.SetTag(currencyCodeSpanTagKey, curr)

			rate, err := u.exrateRepo.GetRate(ctx, curr, expense.Date)
			if err != nil {
				iterErr = errors.Wrapf(err, "failed to get exchange rate for currency=%q at time=%v", curr, expense.Date)
				return false
			}
			exp := *expense
			exp.Amount = rate.ConvertFromBase(exp.Amount)
			return inner(&exp)
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

type noopCache struct{}

func (n noopCache) AddToCache(_ context.Context, _ models.UserID, _, _ time.Time, _ expense.SummaryReport) error {
	return nil
}

func (n noopCache) GetFromCache(_ context.Context, _ models.UserID, _, _ time.Time) (expense.SummaryReport, bool, error) {
	return nil, false, nil
}

func (n noopCache) DropCacheForUserID(_ context.Context, _ models.UserID) error {
	return nil
}

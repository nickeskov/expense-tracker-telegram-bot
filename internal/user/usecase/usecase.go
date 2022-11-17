package usecase

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

const (
	userIDSpanTagKey       = "user_id"
	currencyCodeSpanTagKey = "currency_code"
	monthlyLimitSpanTagKey = "monthly_limit"
)

type UseCase struct {
	repo user.Repository
}

func New(repo user.Repository) (*UseCase, error) {
	return &UseCase{repo: repo}, nil
}

func (u *UseCase) CreateUser(ctx context.Context, um models.User) (_ models.User, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUser")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, um.ID)
	span.SetTag(currencyCodeSpanTagKey, um.SelectedCurrency)

	if err := um.Validate(); err != nil {
		return models.User{}, errors.Wrap(err, "user model validation failed")
	}
	return u.repo.CreateUser(ctx, um)
}

func (u *UseCase) IsUserExists(ctx context.Context, id models.UserID) (_ bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "IsUserExists")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, id)

	return u.repo.IsUserExists(ctx, id)
}

func (u *UseCase) ChangeUserCurrency(ctx context.Context, id models.UserID, currency models.CurrencyCode) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ChangeUserCurrency")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, id)
	span.SetTag(currencyCodeSpanTagKey, currency)

	return u.repo.ChangeUserCurrency(ctx, id, currency)
}

func (u *UseCase) GetUserCurrency(ctx context.Context, id models.UserID) (_ models.CurrencyCode, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetUserCurrency")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, id)

	return u.repo.GetUserCurrency(ctx, id)
}

func (u *UseCase) SetUserMonthlyLimit(ctx context.Context, id models.UserID, limit *decimal.Decimal) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SetUserMonthlyLimit")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, id)
	if limit != nil {
		span.SetTag(monthlyLimitSpanTagKey, limit.String())
	}

	if err := models.ValidateUserMonthlyLimit(limit); err != nil {
		return errors.Wrap(err, "user monthly limit validation failed")
	}
	return u.repo.SetUserMonthlyLimit(ctx, id, limit)
}

func (u *UseCase) GetUserMonthlyLimit(ctx context.Context, id models.UserID) (_ *decimal.Decimal, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetUserMonthlyLimit")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, id)

	return u.repo.GetUserMonthlyLimit(ctx, id)
}

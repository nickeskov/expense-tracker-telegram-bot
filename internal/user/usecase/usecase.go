package usecase

import (
	"context"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

type UseCase struct {
	repo user.Repository
}

func New(repo user.Repository) (*UseCase, error) {
	return &UseCase{repo: repo}, nil
}

func (u *UseCase) CreateUser(ctx context.Context, um models.User) (models.User, error) {
	return u.repo.CreateUser(ctx, um)
}

func (u *UseCase) IsUserExists(ctx context.Context, id models.UserID) (bool, error) {
	return u.repo.IsUserExists(ctx, id)
}

func (u *UseCase) ChangeUserCurrency(ctx context.Context, id models.UserID, currency models.CurrencyCode) error {
	return u.repo.ChangeUserCurrency(ctx, id, currency)
}

func (u *UseCase) GetUserCurrency(ctx context.Context, id models.UserID) (models.CurrencyCode, error) {
	return u.repo.GetUserCurrency(ctx, id)
}

func (u *UseCase) SetUserMonthlyLimit(ctx context.Context, id models.UserID, limit *decimal.Decimal) error {
	return u.repo.SetUserMonthlyLimit(ctx, id, limit)
}

func (u *UseCase) GetUserMonthlyLimit(ctx context.Context, id models.UserID) (*decimal.Decimal, error) {
	return u.repo.GetUserMonthlyLimit(ctx, id)
}

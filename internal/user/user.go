package user

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

var (
	ErrAlreadyExists = errors.New("user already exists")
	ErrDoesNotExist  = errors.New("user does not exist")
)

type Repository interface {
	CreateUser(ctx context.Context, u models.User) (models.User, error)
	IsUserExists(ctx context.Context, id models.UserID) (bool, error)
	ChangeUserCurrency(ctx context.Context, id models.UserID, currency models.CurrencyCode) error
	GetUserCurrency(ctx context.Context, id models.UserID) (models.CurrencyCode, error)
}

type UseCase interface {
	Repository
}

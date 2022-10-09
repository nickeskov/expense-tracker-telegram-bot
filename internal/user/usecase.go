package user

import (
	"context"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type UseCase interface {
	Repository
	CreateUserIfNotExists(ctx context.Context, u models.User) error
}

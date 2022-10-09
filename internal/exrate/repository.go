package exrate

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

var (
	ErrDoesNotExist = errors.New("exchange rate does not exist")
)

type Repository interface {
	GetRate(ctx context.Context, curr models.CurrencyCode, date time.Time) (models.ExchangeRate, error)
	AddOrUpdateRates(ctx context.Context, rates ...models.ExchangeRate) error
}

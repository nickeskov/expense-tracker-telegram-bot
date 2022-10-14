package providers

import (
	"context"
	"time"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type ExchangeRatesProvider interface {
	FetchExchangeRates(ctx context.Context, date time.Time) ([]models.ExchangeRate, error)
}

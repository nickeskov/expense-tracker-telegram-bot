package usecase

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
	"go.uber.org/zap"
)

const (
	currencyCodeSpanTagKey   = "currency_code"
	ratesCountSpanTagKey     = "rates_count"
	dateUnixMillisSpanTagKey = "date_unix_ms"
)

type UseCase struct {
	repo     exrate.Repository
	provider providers.ExchangeRatesProvider
}

func New(repo exrate.Repository, provider providers.ExchangeRatesProvider) (*UseCase, error) {
	return &UseCase{
		repo:     repo,
		provider: provider,
	}, nil
}

func (u *UseCase) GetRate(ctx context.Context, curr models.CurrencyCode, date time.Time) (_ models.ExchangeRate, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetRate")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(currencyCodeSpanTagKey, curr)
	span.SetTag(dateUnixMillisSpanTagKey, date.UnixMilli())

	rate, err := u.repo.GetRate(ctx, curr, date)
	if err != nil {
		if !errors.Is(err, exrate.ErrDoesNotExist) {
			return models.ExchangeRate{}, errors.Wrapf(err, "failed to get rate with curr=%q at time=%v", curr, date)
		}
		if _, err := u.fetchAndUpdateRates(ctx, date); err != nil {
			return models.ExchangeRate{}, errors.Wrap(err, "failed to fetch and update rates")
		}
		rate, err = u.repo.GetRate(ctx, curr, date)
		if err != nil {
			return models.ExchangeRate{}, errors.Wrapf(err, "failed to get rate after with curr=%q at time=%v even after update", curr, date)
		}
	}
	return rate, nil
}

func (u *UseCase) AddOrUpdateRates(ctx context.Context, rates ...models.ExchangeRate) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AddOrUpdateRates")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(ratesCountSpanTagKey, len(rates))

	return u.repo.AddOrUpdateRates(ctx, rates...)
}

func (u *UseCase) fetchAndUpdateRates(ctx context.Context, date time.Time) (_ []models.ExchangeRate, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fetchAndUpdateRates")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(dateUnixMillisSpanTagKey, date.UnixMilli())

	rates, err := u.provider.FetchExchangeRates(ctx, date)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch exchange rates at time=%v", date)
	}
	if err := u.AddOrUpdateRates(ctx, rates...); err != nil {
		return nil, errors.Wrapf(err, "failed to save fetched rates at time=%v", date)
	}
	return rates, nil
}

func (u *UseCase) RunAutoUpdater(ctx context.Context, logger *zap.Logger, interval time.Duration) (<-chan struct{}, error) {
	if interval <= 0 {
		return nil, errors.New("negative or zero auto update interval duration")
	}
	worker := func(done chan<- struct{}) {
		ticker := time.NewTicker(interval)
		defer func() {
			ticker.Stop()
			close(done)
			logger.Info("Exchange rates auto updater successfully stopped")
		}()
		logger.Info("Staring exchange rates auto updater with specific interval", zap.Duration("interval", interval))
		for {
			select {
			case tick := <-ticker.C:
				rates, err := u.fetchAndUpdateRates(ctx, tick.UTC())
				if err != nil {
					logger.Error("Error occurred in rates auto updater", zap.Error(err))
				} else {
					logger.Info("Fetched and saved new exchange rates", zap.Int("count", len(rates)))
				}
			case <-ctx.Done():
				return
			}
		}
	}
	done := make(chan struct{})
	go worker(done)
	return done, nil
}

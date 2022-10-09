package usecase

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
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

func (u *UseCase) GetRate(ctx context.Context, curr models.CurrencyCode, date time.Time) (models.ExchangeRate, error) {
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

func (u *UseCase) AddOrUpdateRates(ctx context.Context, rates ...models.ExchangeRate) error {
	return u.repo.AddOrUpdateRates(ctx, rates...)
}

func (u *UseCase) fetchAndUpdateRates(ctx context.Context, date time.Time) ([]models.ExchangeRate, error) {
	rates, err := u.provider.FetchExchangeRates(ctx, date)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch exchange rates at time=%v", date)
	}
	if err := u.repo.AddOrUpdateRates(ctx, rates...); err != nil {
		return nil, errors.Wrapf(err, "failed to save fetched rates at time=%v", date)
	}
	return rates, nil
}

func (u *UseCase) RunAutoUpdater(ctx context.Context, logger *log.Logger, interval time.Duration) (<-chan struct{}, error) {
	if interval <= 0 {
		return nil, errors.New("negative or zero auto update interval duration")
	}
	worker := func(done chan<- struct{}) {
		ticker := time.NewTicker(interval)
		defer func() {
			ticker.Stop()
			close(done)
			logger.Printf("Exchange rates auto updater successfully stopped")
		}()
		logger.Printf("Staring exchange rates auto updater with interval=%v", interval)
		for {
			select {
			case tick := <-ticker.C:
				select {
				case <-ctx.Done():
					return
				default:
					rates, err := u.fetchAndUpdateRates(ctx, tick.In(time.UTC))
					if err != nil {
						logger.Printf("Rates auto updater: %v", err)
					} else {
						logger.Printf("Fetched and saved new %d exchange rates", len(rates))
					}
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

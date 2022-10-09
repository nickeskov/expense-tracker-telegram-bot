package inmemory

import (
	"context"
	"sync"
	"time"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type (
	currencyRatesByDate       map[time.Time]models.ExchangeRate
	ratesByCurrencyAndDateMap map[models.CurrencyCode]currencyRatesByDate
)

func (r ratesByCurrencyAndDateMap) Get(curr models.CurrencyCode, date time.Time) (models.ExchangeRate, bool) {
	ratesByDate := r.getCurrencyRatesByDate(curr)
	rate, ok := ratesByDate[extractDate(date)]
	return rate, ok
}

func (r ratesByCurrencyAndDateMap) getCurrencyRatesByDate(curr models.CurrencyCode) currencyRatesByDate {
	byCurrencyRates, ok := r[curr]
	if !ok {
		byCurrencyRates = make(map[time.Time]models.ExchangeRate)
		r[curr] = byCurrencyRates
	}
	return byCurrencyRates
}

func (r ratesByCurrencyAndDateMap) Set(rates ...models.ExchangeRate) {
	for _, rate := range rates {
		bucket := r.getCurrencyRatesByDate(rate.Code)
		bucket[extractDate(rate.Date)] = rate
	}
}

func extractDate(t time.Time) time.Time {
	y, m, d := t.In(time.UTC).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type Repository struct {
	mu      *sync.RWMutex
	storage ratesByCurrencyAndDateMap
}

func New() (*Repository, error) {
	return &Repository{
		mu:      &sync.RWMutex{},
		storage: make(ratesByCurrencyAndDateMap),
	}, nil
}

func (r *Repository) GetRate(ctx context.Context, curr models.CurrencyCode, date time.Time) (models.ExchangeRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rate, ok := r.storage.Get(curr, date)
	if !ok {
		return models.ExchangeRate{}, exrate.ErrDoesNotExist
	}
	return rate, nil
}

func (r *Repository) AddOrUpdateRates(ctx context.Context, rates ...models.ExchangeRate) error {
	if len(rates) == 0 { // don't lock mutex if there's nothing to do
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.storage.Set(rates...)
	return nil
}

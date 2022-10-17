package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/database/postgres"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type Repository struct {
	db postgres.DBDoer
}

func New(db postgres.DBDoer) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (r *Repository) GetRate(ctx context.Context, curr models.CurrencyCode, date time.Time) (models.ExchangeRate, error) {
	var rate models.ExchangeRate
	err := r.db.Do(ctx).QueryRowContext(ctx,
		"SELECT currency, date, rate FROM exchange_rates WHERE currency = $1 AND date = $2",
		curr, date.UTC(),
	).Scan(&rate.Code, &rate.Date, &rate.Rate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ExchangeRate{}, exrate.ErrDoesNotExist
		}
		return models.ExchangeRate{}, errors.Wrapf(err, "failed to get rate for currency %q and date %v", curr, date)
	}
	return rate, nil
}

func (r *Repository) AddOrUpdateRates(ctx context.Context, rates ...models.ExchangeRate) (err error) {
	if len(rates) == 0 {
		return nil
	}
	tx, err := r.db.Do(ctx).BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for add or update exchange rates")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrap(err, rollbackErr.Error())
			}
		} else {
			err = tx.Commit()
		}
	}()
	for _, rate := range rates {
		_, err := tx.ExecContext(ctx, `
					INSERT INTO exchange_rates (currency, date, rate)
					VALUES ($1, $2, $3)
					ON CONFLICT ON CONSTRAINT exchange_rates_currency_date_key DO UPDATE SET rate = $3`,
			rate.Code, rate.Date.UTC(), rate.Rate,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to update rate in tx for currency %q at date %v", rate.Code, rate.Date)
		}
	}
	return nil
}

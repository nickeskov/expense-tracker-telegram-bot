package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (r *Repository) CreateUser(ctx context.Context, u models.User) (models.User, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users(id, currency, monthly_limit) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		u.ID, u.SelectedCurrency, u.MonthlyLimit,
	)
	if err != nil {
		return models.User{}, errors.Wrapf(err, "failed to create userID=%d", u.ID)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return models.User{}, errors.Wrapf(err, "failed to create userID=%d", u.ID)
	}
	if affected == 0 {
		return models.User{}, user.ErrAlreadyExists
	}
	return u, nil
}

func (r *Repository) IsUserExists(ctx context.Context, id models.UserID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check whether userID=%d exists", id)
	}
	return exists, nil
}

func (r *Repository) ChangeUserCurrency(ctx context.Context, id models.UserID, currency models.CurrencyCode) error {
	res, err := r.db.ExecContext(ctx, "UPDATE users SET currency = $1 WHERE id = $2", currency, id)
	if err != nil {
		return errors.Wrapf(err, "failed to change currency to %q for userID=%d", currency, id)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to change currency to %q for userID=%d", currency, id)
	}
	if affected == 0 {
		return user.ErrDoesNotExist
	}
	return nil
}

func (r *Repository) GetUserCurrency(ctx context.Context, id models.UserID) (models.CurrencyCode, error) {
	var curr models.CurrencyCode
	err := r.db.QueryRowContext(ctx, "SELECT currency FROM users WHERE id = $1", id).Scan(&curr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return curr, user.ErrDoesNotExist
		}
		return curr, errors.Wrapf(err, "failed to get currency for userID=%d", id)
	}
	return curr, nil
}

func (r *Repository) SetUserMonthlyLimit(ctx context.Context, id models.UserID, limit *decimal.Decimal) error {
	res, err := r.db.ExecContext(ctx, "UPDATE users SET monthly_limit = $1 WHERE id = $2", limit, id)
	if err != nil {
		return errors.Wrapf(err, "failed to set montyly limit %v for userID=%d", limit, id)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to set montyly limit %v for userID=%d", limit, id)
	}
	if affected == 0 {
		return user.ErrDoesNotExist
	}
	return nil
}

func (r *Repository) GetUserMonthlyLimit(ctx context.Context, id models.UserID) (*decimal.Decimal, error) {
	var limit *decimal.Decimal
	err := r.db.QueryRowContext(ctx, "SELECT monthly_limit FROM users WHERE id = $1", id).Scan(&limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrDoesNotExist
		}
		return nil, errors.Wrapf(err, "failed to get monthly limit for userID=%d", id)
	}
	return limit, nil
}

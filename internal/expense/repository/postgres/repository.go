package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (r *Repository) AddExpense(ctx context.Context, userID models.UserID, exp models.Expense) (models.Expense, error) {
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO expenses (user_id, category, amount, date, comment) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, exp.Category, exp.Amount, exp.Date.UTC(), exp.Comment,
	).Scan(&exp.ID)
	if err != nil {
		return models.Expense{}, errors.Wrap(err, "failed to add expense to db")
	}
	return exp, nil
}

func (r *Repository) GetExpensesByDate(ctx context.Context, userID models.UserID, date time.Time) ([]models.Expense, error) {
	var out []models.Expense
	err := r.GetExpensesAscendSinceTill(ctx, userID, date, date, func(e *models.Expense) bool {
		out = append(out, *e)
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get expenses by date")
	}
	return out, nil
}

func (r *Repository) GetExpensesAscendSinceTill(
	ctx context.Context,
	userID models.UserID,
	since, till time.Time,
	iter func(expense *models.Expense) bool,
) (err error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, category, amount, date, comment FROM expenses WHERE user_id = $1 AND date BETWEEN $2 AND $3 ORDER BY date",
		userID, since.UTC(), till.UTC(),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create db query and get expenses since/till")
	}
	defer rows.Close()
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Category, &e.Amount, &e.Date, &e.Comment); err != nil {
			return errors.Wrap(err, "failed to scan expenses since/till")
		}
		if !iter(&e) {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "error occurred after scanning expenses since/till")
	}
	return nil
}

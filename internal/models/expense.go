package models

import (
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var decimalValueLimit = decimal.NewFromInt(10).Shift(21)

var (
	ErrExpenseAmountTooBig        = errors.New("too big expense amount")
	ErrExpenseAmountIsNotPositive = errors.New("expense amount is not positive")
)

type (
	ExpenseID       int64
	ExpenseCategory string
)

type Expense struct {
	ID       ExpenseID
	Category ExpenseCategory
	Amount   decimal.Decimal
	Date     time.Time
	Comment  string
}

func (e *Expense) Validate() error {
	switch {
	case !e.Amount.IsPositive():
		return ErrExpenseAmountIsNotPositive
	case e.Amount.GreaterThanOrEqual(decimalValueLimit):
		return ErrExpenseAmountTooBig
	default:
		return nil
	}
}

package models

import (
	"time"

	"github.com/shopspring/decimal"
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

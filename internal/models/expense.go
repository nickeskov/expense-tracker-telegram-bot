package models

import "time"

type (
	ExpenseID       int64
	ExpenseCategory string
)

type Expense struct {
	ID       ExpenseID
	Category ExpenseCategory
	Amount   float64
	Date     time.Time
	Comment  string
}

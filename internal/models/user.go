package models

import "github.com/shopspring/decimal"

type UserID int64

type User struct {
	ID               UserID
	SelectedCurrency CurrencyCode
	MonthlyLimit     *decimal.Decimal // nil value means no limit
}

func NewUser(id UserID, curr CurrencyCode) User {
	return User{ID: id, SelectedCurrency: curr}
}

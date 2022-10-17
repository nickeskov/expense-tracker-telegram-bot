package models

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrUserMonthlyLimitTooBig     = errors.New("user monthly limit is too big")
	ErrUserMonthlyLimitIsNegative = errors.New("user monthly limit is negative")
)

type UserID int64

type User struct {
	ID               UserID
	SelectedCurrency CurrencyCode
	MonthlyLimit     *decimal.Decimal // nil value means no limit
}

func NewUser(id UserID, curr CurrencyCode) User {
	return User{ID: id, SelectedCurrency: curr}
}

func (u *User) Validate() error {
	return ValidateUserMonthlyLimit(u.MonthlyLimit)
}

func ValidateUserMonthlyLimit(monthlyLimit *decimal.Decimal) error {
	if monthlyLimit == nil {
		return nil
	}
	switch {
	case monthlyLimit.IsNegative():
		return ErrUserMonthlyLimitIsNegative
	case monthlyLimit.GreaterThanOrEqual(decimalValueLimit):
		return ErrUserMonthlyLimitTooBig
	default:
		return nil
	}
}

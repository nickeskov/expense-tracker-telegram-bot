package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CurrencyCode string

type ExchangeRate struct {
	Code CurrencyCode
	Rate decimal.Decimal
	Date time.Time
}

func NewExchangeRate(code CurrencyCode, value decimal.Decimal, date time.Time) ExchangeRate {
	return ExchangeRate{
		Code: code,
		Rate: value,
		Date: date,
	}
}

func (e *ExchangeRate) ConvertFromBase(amountInBaseCurrency decimal.Decimal) decimal.Decimal {
	amount := amountInBaseCurrency.Mul(e.Rate)
	return amount
}

func (e *ExchangeRate) ConvertToBase(amountInSelectedCurrency decimal.Decimal) decimal.Decimal {
	reverseRate := decimal.NewFromInt(1).Div(e.Rate)
	amount := reverseRate.Mul(amountInSelectedCurrency)
	return amount
}

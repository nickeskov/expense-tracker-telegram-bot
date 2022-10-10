package models

import "time"

type (
	CurrencyCode   string
	ExchangeRageID int64
)

type ExchangeRate struct {
	ID   ExchangeRageID
	Code CurrencyCode
	Rate float64
	Date time.Time
}

func NewExchangeRate(code CurrencyCode, value float64, date time.Time) ExchangeRate {
	return ExchangeRate{
		Code: code,
		Rate: value,
		Date: date,
	}
}

func (e *ExchangeRate) ConvertFromBase(amountInBaseCurrency float64) float64 {
	amount := amountInBaseCurrency * e.Rate
	return amount
}

func (e *ExchangeRate) ConvertToBase(amountInSelectedCurrency float64) float64 {
	reverseRate := 1.0 / e.Rate
	amount := reverseRate * amountInSelectedCurrency
	return amount
}

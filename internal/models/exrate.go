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

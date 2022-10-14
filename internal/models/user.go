package models

type UserID int64

type User struct {
	ID               UserID
	SelectedCurrency CurrencyCode
}

func NewUser(id UserID, curr CurrencyCode) User {
	return User{ID: id, SelectedCurrency: curr}
}

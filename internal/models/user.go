package models

import "time"

type (
	UserID      int64
	UserStateID int64
)
type User struct {
	ID    UserID
	State State
}

type State struct {
	ID              UserStateID
	LastInteraction time.Time
}

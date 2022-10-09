package exrate

import (
	"context"
	"time"
)

type UseCase interface {
	Repository
	RunAutoUpdater(ctx context.Context, interval time.Duration) (chan<- struct{}, error)
}

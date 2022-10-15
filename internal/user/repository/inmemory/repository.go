package inmemory

import (
	"context"
	"sync"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
)

type Repository struct {
	mu      *sync.RWMutex
	storage map[models.UserID]models.User
}

func New() (*Repository, error) {
	return &Repository{
		mu:      &sync.RWMutex{},
		storage: make(map[models.UserID]models.User),
	}, nil
}

func (r *Repository) CreateUser(ctx context.Context, u models.User) (models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.storage[u.ID]; ok {
		return models.User{}, user.ErrAlreadyExists
	}
	r.storage[u.ID] = u
	return u, nil
}

func (r *Repository) IsUserExists(ctx context.Context, id models.UserID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.storage[id]
	return ok, nil
}

func (r *Repository) ChangeUserCurrency(ctx context.Context, id models.UserID, currency models.CurrencyCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.storage[id]
	if !ok {
		return user.ErrDoesNotExist
	}
	u.SelectedCurrency = currency
	r.storage[id] = u
	return nil
}

func (r *Repository) GetUserCurrency(ctx context.Context, id models.UserID) (models.CurrencyCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.storage[id]
	if !ok {
		var zero models.CurrencyCode
		return zero, user.ErrDoesNotExist
	}
	return u.SelectedCurrency, nil
}

package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/google/btree"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type Repository struct {
	mu           *sync.RWMutex
	isolatedMu   *sync.Mutex
	userExpenses map[models.UserID]*userExpenses
}
type expensesAtOneDate struct {
	date     time.Time
	expenses []*models.Expense
}

func newExpensesAtOneDate(date time.Time) *expensesAtOneDate {
	y, m, d := date.Date()
	midnight := time.Date(y, m, d, 0, 0, 0, 0, date.Location())
	return &expensesAtOneDate{date: midnight}
}

func (e *expensesAtOneDate) cloneExpenses() []models.Expense {
	if len(e.expenses) == 0 {
		return nil
	}
	out := make([]models.Expense, 0, len(e.expenses))
	for _, exp := range e.expenses {
		out = append(out, *exp)
	}
	return out
}

type userExpenses struct {
	*sync.Mutex
	byDate *btree.BTreeG[*expensesAtOneDate]
	byID   map[models.ExpenseID]*models.Expense
}

const newUserExpensesByDateBTreeDegree = 3

func newUserExpensesByDate(btreeDegree int) *userExpenses {
	less := func(a, b *expensesAtOneDate) bool {
		return a.date.Before(b.date)
	}
	return &userExpenses{
		&sync.Mutex{},
		btree.NewG(btreeDegree, less),
		map[models.ExpenseID]*models.Expense{},
	}
}

func New() (*Repository, error) {
	return &Repository{
		mu:           &sync.RWMutex{},
		isolatedMu:   &sync.Mutex{},
		userExpenses: map[models.UserID]*userExpenses{},
	}, nil
}

func (r *Repository) tryGetUserExpenses(userID models.UserID) (*userExpenses, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	expenses, ok := r.userExpenses[userID]
	return expenses, ok
}

func (r *Repository) tryInitUserExpenses(userID models.UserID) *userExpenses {
	newExpenses := newUserExpensesByDate(newUserExpensesByDateBTreeDegree)
	r.mu.Lock()
	defer r.mu.Unlock()
	expenses, ok := r.userExpenses[userID]
	if !ok {
		r.userExpenses[userID] = newExpenses
		expenses = newExpenses
	}
	return expenses
}

func (r *Repository) getUserExpenses(userID models.UserID) *userExpenses {
	expenses, ok := r.tryGetUserExpenses(userID)
	if !ok {
		expenses = r.tryInitUserExpenses(userID)
	}
	return expenses
}

func (r *Repository) Isolated(ctx context.Context, callback func(ctx context.Context) error) error {
	r.isolatedMu.Lock()
	defer r.isolatedMu.Unlock()
	return callback(ctx)
}

func (r *Repository) AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error) {
	expenses := r.getUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	keyVal := newExpensesAtOneDate(expense.Date)
	expensesAtOneDay, ok := expenses.byDate.Get(keyVal)
	if !ok {
		expensesAtOneDay = keyVal
		expenses.byDate.ReplaceOrInsert(expensesAtOneDay)
	}
	expensesAtOneDay.expenses = append(expensesAtOneDay.expenses, &expense)
	expenses.byID[expense.ID] = &expense
	return expense, nil
}

func (r *Repository) GetExpensesByDate(ctx context.Context, userID models.UserID, date time.Time) ([]models.Expense, error) {
	expenses := r.getUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	expensesAtOneDay, ok := expenses.byDate.Get(newExpensesAtOneDate(date))
	if !ok {
		return nil, nil
	}
	return expensesAtOneDay.cloneExpenses(), nil
}

func (r *Repository) GetExpensesAscendSinceTill(
	ctx context.Context,
	userID models.UserID,
	since, till time.Time,
	iter func(expense *models.Expense) bool,
) error {
	expenses := r.getUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	if since.Equal(till) {
		key := newExpensesAtOneDate(since)
		if atOneDate, ok := expenses.byDate.Get(key); ok {
			for _, e := range atOneDate.expenses {
				if !iter(e) {
					return nil
				}
			}
		}
	} else {
		var (
			greaterOrEqual = newExpensesAtOneDate(since)
			lessThan       = newExpensesAtOneDate(till)
		)
		expenses.byDate.AscendGreaterOrEqual(greaterOrEqual, func(atOneDate *expensesAtOneDate) bool {
			if atOneDate.date.After(lessThan.date) {
				return false
			}
			for _, e := range atOneDate.expenses {
				if !iter(e) {
					return false
				}
			}
			return true
		})
	}
	return nil
}

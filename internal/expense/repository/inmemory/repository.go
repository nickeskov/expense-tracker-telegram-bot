package inmemory

import (
	"sync"
	"time"

	"github.com/google/btree"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type Repository struct {
	mu           *sync.Mutex
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

func New() *Repository {
	return &Repository{
		mu:           &sync.Mutex{},
		userExpenses: map[models.UserID]*userExpenses{},
	}
}

func (r *Repository) getOrInitUserExpenses(userID models.UserID) *userExpenses {
	r.mu.Lock()
	defer r.mu.Unlock()
	expenses, ok := r.userExpenses[userID]
	if !ok {
		expenses = newUserExpensesByDate(newUserExpensesByDateBTreeDegree)
		r.userExpenses[userID] = expenses
	}
	return expenses
}

func (r *Repository) AddExpense(id models.UserID, e models.Expense) (models.Expense, error) {
	expenses := r.getOrInitUserExpenses(id)
	expenses.Lock()
	defer expenses.Unlock()

	keyVal := newExpensesAtOneDate(e.Date)
	expensesAtOneDay, ok := expenses.byDate.Get(keyVal)
	if !ok {
		expensesAtOneDay = keyVal
		expenses.byDate.ReplaceOrInsert(expensesAtOneDay)
	}
	expensesAtOneDay.expenses = append(expensesAtOneDay.expenses, &e)
	expenses.byID[e.ID] = &e
	return e, nil
}

func (r *Repository) GetExpense(userID models.UserID, expenseID models.ExpenseID) (models.Expense, error) {
	expenses := r.getOrInitUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	e, ok := expenses.byID[expenseID]
	if !ok {
		return models.Expense{}, expense.ErrExpenseDoesNotExist
	}
	return *e, nil
}

func (r *Repository) ExpensesByDate(userID models.UserID, date time.Time) ([]models.Expense, error) {
	expenses := r.getOrInitUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	expensesAtOneDay, ok := expenses.byDate.Get(newExpensesAtOneDate(date))
	if !ok {
		return nil, nil
	}
	return expensesAtOneDay.cloneExpenses(), nil
}

func (r *Repository) ExpensesSummaryByCategorySince(userID models.UserID, since, till time.Time) (map[models.ExpenseCategory]float64, error) {
	out := make(map[models.ExpenseCategory]float64)
	expenses := r.getOrInitUserExpenses(userID)
	expenses.Lock()
	defer expenses.Unlock()

	if since.Equal(till) {
		key := newExpensesAtOneDate(since)
		if atOneDate, ok := expenses.byDate.Get(key); ok {
			for _, e := range atOneDate.expenses {
				out[e.Category] += e.Amount
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
				out[e.Category] += e.Amount
			}
			return true
		})
	}
	return out, nil
}

func (r *Repository) Close() error {
	return nil // no-op
}

package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Statementer interface {
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

type Transactioner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type DBInteractor interface {
	Transactioner
	Querier
	Execer
}

type Tx interface {
	driver.Tx
	Transactioner
	Statementer
	DBInteractor
}

type DBDoer interface {
	Do(ctx context.Context) DBInteractor
	DoIsolated(ctx context.Context, opts *sql.TxOptions, callback func(ctx context.Context) error) (err error)
}

type txKey struct{}

func extractTxFromCtx(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(Tx)
	return tx, ok
}
func storeTxToCtx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

type doer struct {
	db *db
}

func NewDBDoer(sqlDB *sql.DB) DBDoer {
	innerDB := &db{DB: sqlDB}
	return &doer{db: innerDB}
}

func (d *doer) Do(ctx context.Context) DBInteractor {
	tx, ok := extractTxFromCtx(ctx)
	if !ok {
		return d.db
	}
	return tx
}

func (d *doer) DoIsolated(ctx context.Context, opts *sql.TxOptions, callback func(ctx context.Context) error) (err error) {
	tx, err := d.Do(ctx).BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrap(err, rollbackErr.Error())
			}
		} else {
			err = tx.Commit()
		}
	}()
	return callback(storeTxToCtx(ctx, tx))
}

type db struct {
	*sql.DB
}

func (d *db) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	sqlTx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &transaction{ctx: ctx, Tx: sqlTx}, nil
}

type transaction struct {
	*sql.Tx
	savepoint string
	ctx       context.Context
}

func (tx *transaction) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	uuID, err := uuid.New().MarshalBinary()
	if err != nil {
		return nil, err
	}
	savepoint := hex.EncodeToString(uuID)
	_, err = tx.Tx.ExecContext(ctx, "SAVEPOINT $1", savepoint)
	if err != nil {
		return nil, err
	}
	return &transaction{
		Tx:        tx.Tx,
		savepoint: savepoint,
		ctx:       ctx,
	}, nil

}

func (tx *transaction) Commit() error {
	if tx.savepoint == "" {
		return tx.Tx.Commit()
	}
	_, err := tx.Tx.ExecContext(tx.ctx, "RELEASE SAVEPOINT $1", tx.savepoint)
	return err
}

func (tx *transaction) Rollback() error {
	if tx.savepoint == "" {
		return tx.Tx.Rollback()
	}
	_, err := tx.Tx.ExecContext(tx.ctx, "ROLLBACK TO SAVEPOINT $1", tx.savepoint)
	return err
}

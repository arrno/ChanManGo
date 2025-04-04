package datastore

import (
	"context"

	"github.com/arrno/chanman/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ctx = context.Background()
)

// SQLDatabase defines a generic interface for interacting with an SQL database.
type SQLDatabase interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	BeginTx(ctx context.Context) (SQLDatabase, error)
	Close()
	WithTx(ctx context.Context, f func(tx SQLDatabase) error) error
}

// PostgresDB implements SQLDatabase using pgxpool for connection pooling.
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresDB initializes a new database connection pool.
func NewPostgresDB(dsn string) (*PostgresDB, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

func (db *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

func (db *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

func (db *PostgresDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.pool, dest, query, args...)
}

func (db *PostgresDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.pool, dest, query, args...)
}

func (db *PostgresDB) BeginTx(ctx context.Context) (SQLDatabase, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &PostgresTx{tx: tx}, nil
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

func (db *PostgresDB) WithTx(ctx context.Context, f func(tx SQLDatabase) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		tx.(*PostgresTx).Rollback()
		return err
	}
	return tx.(*PostgresTx).Commit()
}

// PostgresTx implements SQLDatabase within a transaction.
type PostgresTx struct {
	tx pgx.Tx
}

func (tx *PostgresTx) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return tx.tx.Query(ctx, query, args...)
}

func (tx *PostgresTx) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return tx.tx.QueryRow(ctx, query, args...)
}

func (tx *PostgresTx) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return tx.tx.Exec(ctx, query, args...)
}

func (tx *PostgresTx) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, tx.tx, dest, query, args...)
}

func (tx *PostgresTx) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, tx.tx, dest, query, args...)
}

func (tx *PostgresTx) BeginTx(ctx context.Context) (SQLDatabase, error) {
	return nil, domain.ErrNotImplemented // Nested transactions not supported in pgx
}

func (tx *PostgresTx) Close() {
	// No-op for transactions
}

func (tx *PostgresTx) Commit() error {
	return tx.tx.Commit(context.Background())
}

func (tx *PostgresTx) Rollback() error {
	return tx.tx.Rollback(context.Background())
}

func (tx *PostgresTx) WithTx(ctx context.Context, f func(tx SQLDatabase) error) error {
	return domain.ErrNotImplemented // Nested transactions not supported
}

// // Example Usage
// type User struct {
// 	ID    int    `db:"id"`
// 	Name  string `db:"name"`
// 	Email string `db:"email"`
// }

// func GetUsers(ctx context.Context, db SQLDatabase) ([]User, error) {
// 	var users []User
// 	err := db.Select(ctx, &users, "SELECT id, name, email FROM users")
// 	return users, err
// }

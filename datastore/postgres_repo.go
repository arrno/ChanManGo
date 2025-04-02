package datastore

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// PostgresRepo implements the full RepoProvider interface by composing
// the SystemRepo and MessageRepo implementations.
type PostgresRepo struct {
	SystemRepo
	MessageRepo
	db SQLDatabase
}

// NewPostgresRepo creates a new PostgreSQL repository with a database connection
func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := NewPostgresDB(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &PostgresRepo{
		SystemRepo:  SystemRepo{db: db},
		MessageRepo: MessageRepo{db: db},
		db:          db,
	}, nil
}

// WithTx executes a function within a transaction
func (r *PostgresRepo) WithTx(f func(tx RepoProvider) error) error {
	return r.db.WithTx(ctx, func(txDb SQLDatabase) error {
		txRepo := &PostgresRepo{
			SystemRepo:  SystemRepo{db: txDb},
			MessageRepo: MessageRepo{db: txDb},
			db:          txDb,
		}
		return f(txRepo)
	})
}

// Close closes the underlying database connection
func (r *PostgresRepo) Close() {
	r.db.Close()
}

// NewRepoProvider creates a repository based on environment configuration
func NewRepoProvider() (RepoProvider, error) {
	pgDSN := "postgresql://myuser:password@localhost:5433/postgres"
	// pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN != "" {
		// Use PostgreSQL if DSN is provided
		repo, err := NewPostgresRepo(pgDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL repository: %w", err)
		}
		return repo, nil
	}

	log.Warn().Msg("Repo provider -> defaulting to in memory db.")
	// Default to in-memory repository
	return NewInMemoryDB(), nil
}

package datastore

import (
	"errors"
	"fmt"
	"time"

	"github.com/arrno/chanman/domain"
)

var (
	ErrNotFound = errors.New("Target record not found")
)

// RepoProvider is the main interface for data storage operations
type RepoProvider interface {
	MessageRepoProvider
	SystemRepoProvider
	Close()
	WithTx(f func(tx RepoProvider) error) error
}

// MessageRepoProvider handles channel, subscription, and message operations
type MessageRepoProvider interface {
	StoreChannel(*domain.Channel) error
	GetChannel(id string) (*domain.Channel, error)
	GetChannelByGroupAndName(groupId, channelName string) (*domain.Channel, error)
	UpdateChannel(*domain.Channel) error
	ArchiveChannel(id string) error
	//
	StoreTopicSubscriber(*domain.TopicSubscription) error
	GetTopicSubscriber(id string) (*domain.TopicSubscription, error)
	UpdateTopicSubscriber(*domain.TopicSubscription) error
	ArchiveTopicSubscriber(id string) error
	//
	StoreChannelSubscriber(*domain.ChannelSubscription) error
	GetChannelSubscriber(id string) (*domain.ChannelSubscription, error)
	UpdateChannelSubscriber(*domain.ChannelSubscription) error
	ArchiveChannelSubscriber(id string) error
	//
	StoreMessage(*domain.Message) error
	GetMessage(id string) (*domain.Message, error)
	UpdateMessage(*domain.Message) error
	ArchiveMessage(id string) error
	PruneMessages(retentionHours uint) error
}

// TODO all these queries should honor the caller's portal

// SystemRepoProvider handles client, portal, and group operations
type SystemRepoProvider interface {
	StoreClient(*domain.Client) error
	GetClient(id string) (*domain.Client, error)
	GetClients(pagination *domain.Pagination[*domain.Client], filter *domain.Filter) ([]*domain.Client, error)
	UpdateClient(*domain.Client) error
	ArchiveClient(id string) error
	AggregateClientMessageUsage(start, stop time.Time) ([]domain.ClientUsage, error)
	//
	GetClientAuthByToken(token domain.Token) (*domain.ClientGroup, *domain.ClientPermission, bool, bool, error)
	//
	StoreClientSecret(*domain.ClientSecret) error
	GetClientSecret(id string) (*domain.ClientSecret, error)
	GetClientSecretByClientId(clientId string) (*domain.ClientSecret, error)
	UpdateClientSecret(*domain.ClientSecret) error
	ArchiveClientSecret(id string) error
	//
	StoreClientPermission(*domain.ClientPermission) error
	GetClientPermission(id string) (*domain.ClientPermission, error)
	GetClientPermissionByClientId(clientId string) (*domain.ClientPermission, error)
	UpdateClientPermission(*domain.ClientPermission) error
	ArchiveClientPermission(id string) error
	//
	StoreGroup(*domain.Group) error
	GetGroup(id string) (*domain.Group, error)
	UpdateGroup(*domain.Group) error
	ArchiveGroup(id string) error
	//
	StoreClientGroup(*domain.ClientGroup) error
	GetClientGroupsByClientId(clientId string) ([]*domain.ClientGroup, error)
	GetClientGroup(id string) (*domain.ClientGroup, error)
	UpdateClientGroup(*domain.ClientGroup) error
	ArchiveClientGroup(id string) error
	//
	StorePortal(*domain.Portal) error
	GetPortal(id string) (*domain.Portal, error)
	UpdatePortal(*domain.Portal) error
	ArchivePortal(id string) error
}

// SQLRepo implements the full RepoProvider interface by composing
// the SystemRepo and MessageRepo implementations.
type SQLRepo struct {
	SystemRepo
	MessageRepo
	db SQLDatabase
}

// NewSQLRepo creates a new PostgreSQL repository with a database connection
func NewSQLRepo(dsn string) (*SQLRepo, error) {
	db, err := NewPostgresDB(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &SQLRepo{
		SystemRepo:  SystemRepo{db: db},
		MessageRepo: MessageRepo{db: db},
		db:          db,
	}, nil
}

// WithTx executes a function within a transaction
func (r *SQLRepo) WithTx(f func(tx RepoProvider) error) error {
	return r.db.WithTx(ctx, func(txDb SQLDatabase) error {
		txRepo := &SQLRepo{
			SystemRepo:  SystemRepo{db: txDb},
			MessageRepo: MessageRepo{db: txDb},
			db:          txDb,
		}
		return f(txRepo)
	})
}

// Close closes the underlying database connection
func (r *SQLRepo) Close() {
	r.db.Close()
}

// NewRepoProvider creates a repository based on environment configuration
func NewRepoProvider() (RepoProvider, error) {
	pgDSN := "postgresql://myuser:password@localhost:5433/postgres"
	// pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		return nil, errors.New("failed to create PostgreSQL repository: pgDSN not found.")
	}

	// Use PostgreSQL if DSN is provided
	repo, err := NewSQLRepo(pgDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL repository: %w", err)
	}
	return repo, nil
}

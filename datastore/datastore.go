package datastore

import (
	"errors"

	"github.com/arrno/chanman/domain"
)

// many messages from a subscriber
// many subscribers to a worker
// many workers to a channel
// many messages in a channel

var (
	ErrNotFound = errors.New("Target record not found")
)

type DatastoreProvider interface {
	MessageDatastoreProvider
	SystemDatastoreProvider
	WithTx(f func(tx DatastoreProvider) error) error
}

// Channel / Subscriptions /Messages
type MessageDatastoreProvider interface {
	StoreChannel(*domain.Channel) error
	GetChannel(id string) (*domain.Channel, error)
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
}

// TODO all these queries should honor the caller's portal

// Client / Portal / Groups
type SystemDatastoreProvider interface {
	StoreClient(*domain.Client) error
	GetClient(id string) (*domain.Client, error)
	GetClients(pagination domain.Pagination[*domain.Client]) ([]*domain.Client, error)
	UpdateClient(*domain.Client) error
	ArchiveClient(id string) error
	//
	GetClientGroupByToken(token string) (*domain.ClientGroup, error)
	//
	StoreClientSecret(*domain.ClientSecret) error
	GetClientSecret(id string) (*domain.ClientSecret, error)
	GetClientSecretByClientId(clientId string) (*domain.ClientSecret, error)
	UpdateClientSecret(*domain.ClientSecret) error
	ArchiveClientSecret(id string) error
	//
	StoreGroup(*domain.Group) error
	GetGroup(id string) (*domain.Group, error)
	UpdateGroup(*domain.Group) error
	ArchiveGroup(id string) error
	//
	StoreClientGroup(*domain.ClientGroup) error
	GetClientGroupByClientId(clientId string) (*domain.ClientGroup, error)
	GetClientGroup(id string) (*domain.ClientGroup, error)
	UpdateClientGroup(*domain.ClientGroup) error
	ArchiveClientGroup(id string) error
	//
	StorePortal(*domain.Portal) error
	GetPortal(id string) (*domain.Portal, error)
	UpdatePortal(*domain.Portal) error
	ArchivePortal(id string) error
}

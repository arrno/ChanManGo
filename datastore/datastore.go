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

// Client / Portal / Groups
type DatastoreProvider interface {
	StoreChannel(*domain.Channel) error
	GetChannel(id string) (*domain.Channel, error)
	UpdateChannel(*domain.Channel) error
	ArchiveChannel(id string) error
	//
	StoreWorker(*domain.Channel) error
	GetWorker(id string) (*domain.Channel, error)
	UpdateWorker(*domain.Channel) error
	ArchiveWorker(id string) error
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
	//
	WithTx(f func(tx DatastoreProvider) error) error
}

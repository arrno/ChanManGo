package datastore

import (
	"github.com/arrno/chanban/domain"
)

// many messages from a subscriber
// many subscribers to a worker
// many workers to a channel
// many messages in a channel

type IDatastore interface {
	StoreChannel(*domain.Channel) error
	GetChannel(id string) (*domain.Channel, error)
	UpdateChannel(*domain.Channel) error
	ArchiveChannel(id string) error
	//
	StoreWorker(*domain.Worker) error
	GetWorker(id string) (*domain.Worker, error)
	UpdateWorker(*domain.Worker) error
	ArchiveWorker(id string) error
	//
	StoreSubscriber(*domain.Subscription) error
	GetSubscriber(id string) (*domain.Subscription, error)
	UpdateSubscriber(*domain.Subscription) error
	ArchiveSubscriber(id string) error
	//
	StoreMessage(*domain.Message) error
	GetMessage(id string) (*domain.Message, error)
	UpdateMessage(*domain.Message) error
	ArchiveMessage(id string) error
}

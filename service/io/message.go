package io

import (
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type MessageProvider interface {
	CreateMessage(message *domain.Message) error
	GetMessage(id string) (*domain.Message, error)
	UpdateMessage(message *domain.Message) error
	ArchiveMessage(id string) error
}

type Message struct {
	db datastore.DatastoreProvider
}

func NewMessage(db datastore.DatastoreProvider) *Message {
	return &Message{db}
}

func (m *Message) CreateMessage(message *domain.Message) error {
	return m.db.StoreMessage(message)
}
func (m *Message) GetMessage(id string) (*domain.Message, error) {
	return m.db.GetMessage(id)
}
func (m *Message) UpdateMessage(message *domain.Message) error {
	return m.db.UpdateMessage(message)
}
func (m *Message) ArchiveMessage(id string) error {
	return m.db.ArchiveMessage(id)
}

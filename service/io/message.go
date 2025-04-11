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
	PruneMessages(retentionHours uint) error
}

type Message struct {
	repo datastore.RepoProvider
}

func NewMessage(repo datastore.RepoProvider) *Message {
	return &Message{repo}
}

func (m *Message) CreateMessage(message *domain.Message) error {
	return m.repo.StoreMessage(message)
}
func (m *Message) GetMessage(id string) (*domain.Message, error) {
	return m.repo.GetMessage(id)
}
func (m *Message) UpdateMessage(message *domain.Message) error {
	return m.repo.UpdateMessage(message)
}
func (m *Message) ArchiveMessage(id string) error {
	return m.repo.ArchiveMessage(id)
}
func (m *Message) PruneMessages(retentionHours uint) error {
	return m.repo.PruneMessages(retentionHours)
}

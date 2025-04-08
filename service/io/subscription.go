package io

import (
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type SubscriberProvider interface {
	CreateTopicSubscriber(subscriber *domain.TopicSubscription) error
	GetTopicSubscriber(id string) (*domain.TopicSubscription, error)
	UpdateTopicSubscriber(subscriber *domain.TopicSubscription) error
	ArchiveTopicSubscriber(id string) error
}

type Subscriber struct {
	repo datastore.RepoProvider
}

func NewSubscriber(repo datastore.RepoProvider) SubscriberProvider {
	return &Subscriber{repo}
}

func (m *Subscriber) CreateTopicSubscriber(subscriber *domain.TopicSubscription) error {
	return m.repo.StoreTopicSubscriber(subscriber)
}
func (m *Subscriber) GetTopicSubscriber(id string) (*domain.TopicSubscription, error) {
	return m.repo.GetTopicSubscriber(id)
}
func (m *Subscriber) UpdateTopicSubscriber(subscriber *domain.TopicSubscription) error {
	return m.repo.UpdateTopicSubscriber(subscriber)
}
func (m *Subscriber) ArchiveTopicSubscriber(id string) error {
	return m.repo.ArchiveTopicSubscriber(id)
}

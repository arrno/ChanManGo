package io

import (
	"errors"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

type ChannelProvider interface {
	CreateChannel(channel *domain.Channel) (bool, error)
	GetChannel(id string) (*domain.Channel, error)
	UpdateChannel(channel *domain.Channel) error
	ArchiveChannel(id string) error
}

type Channel struct {
	db datastore.DatastoreProvider
}

func NewChannel(db datastore.DatastoreProvider) *Channel {
	return &Channel{db}
}

func (m *Channel) CreateChannel(channel *domain.Channel) (created bool, err error) {
	err = m.db.WithTx(func(tx datastore.DatastoreProvider) error {
		channelExists, err := m.channelExistsTx(tx, channel)
		if err != nil {
			return err
		}
		created = !channelExists
		if channelExists {
			return nil
		}
		log.Info().Msg("Creating channel n DB: " + channel.Id)
		if err := tx.StoreChannel(channel); err != nil {
			return err
		}
		return nil
	})
	return
}
func (m *Channel) GetChannel(id string) (*domain.Channel, error) {
	return m.db.GetChannel(id)
}
func (m *Channel) UpdateChannel(channel *domain.Channel) error {
	return m.db.UpdateChannel(channel)
}
func (m *Channel) ArchiveChannel(id string) error {
	return m.db.ArchiveChannel(id)
}

func (m *Channel) channelExistsTx(tx datastore.DatastoreProvider, channel *domain.Channel) (bool, error) {
	_, err := tx.GetChannel(channel.Id)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}

	return false, err
}

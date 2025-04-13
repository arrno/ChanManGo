package io

import (
	"errors"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

type ChannelProvider interface {
	CreateChannelIfNotFound(channel *domain.Channel) (bool, error)
}

type Channel struct {
	repo datastore.RepoProvider
}

func NewChannel(repo datastore.RepoProvider) *Channel {
	return &Channel{repo}
}

func (m *Channel) CreateChannelIfNotFound(channel *domain.Channel) (created bool, err error) {
	err = m.repo.WithTx(func(tx datastore.RepoProvider) error {
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

func (m *Channel) channelExistsTx(tx datastore.RepoProvider, channel *domain.Channel) (bool, error) {
	_, err := tx.GetChannel(channel.Id)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}

	return false, err
}

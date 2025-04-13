package io

import (
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

type SystemProvider struct {
	datastore.RepoProvider
	ChannelProvider
	ClientProvider
	GroupProvider
}

func NewSystem(repo datastore.RepoProvider) SystemProvider {
	return SystemProvider{
		ClientProvider:  NewClient(repo),
		ChannelProvider: NewChannel(repo),
		GroupProvider:   NewGroup(repo),
	}
}

func (s *SystemProvider) InitSandbox(
	client *domain.Client,
	channel *domain.Channel,
	group *domain.Group,
	permission *domain.ClientPermission,
) (domain.Token, error) {
	var token domain.Token
	return token, s.WithTx(func(tx datastore.RepoProvider) error {
		// Create all resources in a single transaction

		// 1. Create group
		if err := tx.StoreGroup(group); err != nil {
			return err
		}

		// 2. Create channel
		if err := tx.StoreChannel(channel); err != nil {
			return err
		}

		// 3. Create client with permissions
		if err := tx.StoreClient(client); err != nil {
			return err
		}

		// 4. Create client permissions
		if err := tx.StoreClientPermission(permission); err != nil {
			return err
		}

		// 5. Create client token
		// var tokenErr error
		// token, tokenErr = domain.NewTokenWithExpiry(util.Nanoid(), client.Id, 24*time.Hour)
		// if tokenErr != nil {
		// 	return tokenErr
		// }
		token = *(domain.NewToken(util.Nanoid(), client.Id))
		token.Hash()

		secret := domain.NewClientSecretFromToken(token)
		if err := tx.StoreClientSecret(&secret); err != nil {
			return err
		}

		// 6. Create client-group association
		clientGroup := domain.NewClientGroup(client.Id, group.Id)
		if err := tx.StoreClientGroup(clientGroup); err != nil {
			return err
		}

		return nil
	})
}

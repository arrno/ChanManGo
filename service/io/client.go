package io

import (
	"errors"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

type ClientProvider interface {
	CreateClient(client *domain.Client, permission *domain.ClientPermission) (domain.Token, error)
	//
	RotateClientSecret(clientId string) (domain.Token, error)
	ValidateClientSecret(clientId, token string) error
	FetchActiveClaims(token domain.Token) (*domain.ClientGroup, *domain.ClientPermission, bool, bool, error)
}

type Client struct {
	repo datastore.RepoProvider
}

func NewClient(repo datastore.RepoProvider) *Client {
	return &Client{
		repo: repo,
	}
}

func (c *Client) CreateClient(client *domain.Client, permission *domain.ClientPermission) (domain.Token, error) {
	exists, err := c.clientExists(client.Id)
	if err != nil {
		return domain.Token{}, err
	} else if exists {
		return domain.Token{}, domain.ErrCreateOnExistingRecord
	}

	token := domain.NewToken(util.Nanoid(), client.Id).Hash()
	secret := domain.NewClientSecretFromToken(*token)
	return *token, c.repo.WithTx(func(tx datastore.RepoProvider) error {
		if err := tx.StoreClient(client); err != nil {
			return err
		}
		if err := tx.StoreClientSecret(&secret); err != nil {
			return err
		}
		if err = tx.StoreClientGroup(domain.NewClientGroup(client.Id, domain.MAIN_GROUP)); err != nil {
			return err
		}
		if err = tx.StoreClientPermission(permission); err != nil {
			return err
		}
		return nil
	})
}

func (c *Client) RotateClientSecret(clientId string) (domain.Token, error) {
	// First validate the client exists
	client, err := c.repo.GetClient(clientId)
	if err != nil {
		return domain.Token{}, err
	}

	// Create new token
	token := domain.NewToken(util.Nanoid(), client.Id).Hash()

	// Create new secret object
	secret := domain.NewClientSecretFromToken(*token)

	err = c.repo.StoreClientSecret(&secret)
	if err != nil {
		return domain.Token{}, err
	}

	return *token, nil
}

func (c *Client) ValidateClientSecret(clientId, password string) error {
	secret, err := c.repo.GetClientSecretByClientId(clientId)
	if err != nil {
		return err
	}
	if secret.ArchivedAt != nil {
		return domain.ErrTokenArchived
	}
	return domain.Token{
		Hashed: secret.EncryptedSecret,
		Salt:   secret.Salt,
	}.Validate(password)
}

func (c *Client) clientExists(clientId string) (bool, error) {
	_, err := c.repo.GetClient(clientId)
	if err == nil {
		return true, nil
	} else if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (c *Client) FetchActiveClaims(token domain.Token) (*domain.ClientGroup, *domain.ClientPermission, bool, bool, error) {
	clientGroup, permission, isAdmin, expired, err := c.repo.GetClientAuthByToken(token)
	if clientGroup.ArchivedAt != nil {
		clientGroup = nil
	}
	if permission.ArchivedAt != nil {
		permission = nil
	}
	return clientGroup, permission, isAdmin, expired, err
}

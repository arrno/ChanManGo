package io

import (
	"errors"
	"time"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

type ClientProvider interface {
	CreateClient(client *domain.Client, permission *domain.ClientPermission) (domain.Token, error)
	GetClient(id string) (*domain.Client, error)
	GetClients(pg domain.Pagination[*domain.Client]) ([]*domain.Client, error)
	UpdateClient(client *domain.Client) error
	ArchiveClient(id string) error
	//
	RotateClientSecret(clientId string) (domain.Token, error)
	ValidateClientSecret(clientId, token string) (bool, error)
}

type Client struct {
	db datastore.DatastoreProvider
}

func NewClient(db datastore.DatastoreProvider) *Client {
	return &Client{
		db: db,
	}
}

func (c *Client) CreateClient(client *domain.Client, permission *domain.ClientPermission) (domain.Token, error) {
	exists, err := c.clientExists(client.Id)
	if err != nil {
		return domain.Token{}, err
	} else if exists {
		return domain.Token{}, ErrCreateOnExistingRecord
	}
	localSecret := util.NewEnv().Expect("LOCAL_SECRET")
	token, err := domain.NewToken(util.Nanoid(), string(localSecret))
	if err != nil {
		return token, err
	}
	secret := domain.NewClientSecret(client.Id, token.Encrypted)
	return token, c.db.WithTx(func(tx datastore.DatastoreProvider) error {
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

func (c *Client) GetClient(id string) (*domain.Client, error) {
	client, err := c.db.GetClient(id)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) GetClients(pg domain.Pagination[*domain.Client]) ([]*domain.Client, error) {
	clients, err := c.db.GetClients(pg)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func (c *Client) UpdateClient(client *domain.Client) error {
	return c.db.UpdateClient(client)
}

func (c *Client) ArchiveClient(id string) error {
	client, err := c.GetClient(id)
	if err != nil {
		return err
	}
	now := time.Now()
	client.ArchivedAt = &now
	client.Active = false
	return c.db.UpdateClient(client)
}

func (c *Client) RotateClientSecret(clientId string) (domain.Token, error) {
	client, err := c.GetClient(clientId)
	if err != nil {
		return domain.Token{}, err
	}
	localSecret := util.NewEnv().Expect("LOCAL_SECRET")
	token, err := domain.NewToken(util.Nanoid(), localSecret)
	if err != nil {
		return domain.Token{}, err
	}
	secret := domain.NewClientSecret(client.Id, token.Encrypted)
	if err := c.db.StoreClientSecret(&secret); err != nil {
		return domain.Token{}, err
	}
	return token, nil
}

func (c *Client) ValidateClientSecret(clientId, password string) (bool, error) {
	secret, err := c.db.GetClientSecretByClientId(clientId)
	if err != nil {
		return false, err
	}
	return domain.Token{
		Encrypted: secret.EncryptedSecret,
	}.Validate(password), nil
}

func (c *Client) clientExists(clientId string) (bool, error) {
	_, err := c.db.GetClient(clientId)
	if err == nil {
		return true, nil
	} else if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

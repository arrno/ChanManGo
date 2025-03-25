package service

import (
	"errors"
	"time"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

type ClientProvider interface {
	CreateClient(client domain.Client) error
	GetClient(id string) (domain.Client, error)
	UpdateClient(client domain.Client) error
	ArchiveClient(id string) error
	//
	RotateClientSecret(clientId string) error
	ValidateClientSecret(clientId, token string) (bool, error)
}

// generate new token / salt / hash

type Client struct {
	db datastore.DatastoreProvider
}

func (c *Client) CreateClient(client domain.Client) error {
	exists, err := c.clientExists(client.Id)
	if err != nil {
		return err
	} else if exists {
		return ErrCreateOnExistingRecord
	}
	token, err := domain.NewToken(util.Nanoid())
	if err != nil {
		return err
	}
	secret := domain.NewClientSecret(client.Id, token.Hash)
	return c.db.WithTx(func(tx datastore.DatastoreProvider) error {
		if err := tx.StoreClient(&client); err != nil {
			return err
		}
		if err := tx.StoreClientSecret(&secret); err != nil {
			return err
		}
		return nil
	})
}

func (c *Client) GetClient(id string) (domain.Client, error) {
	client, err := c.db.GetClient(id)
	if err != nil {
		return domain.Client{}, err
	}
	return *client, nil
}

func (c *Client) UpdateClient(client domain.Client) error {
	return c.db.UpdateClient(&client)
}

func (c *Client) ArchiveClient(id string) error {
	client, err := c.GetClient(id)
	if err != nil {
		return err
	}
	now := time.Now()
	client.ArchivedAt = &now
	client.Active = false
	return c.db.UpdateClient(&client)
}

func (c *Client) RotateClientSecret(clientId string) error {
	client, err := c.GetClient(clientId)
	if err != nil {
		return err
	}
	token, err := domain.NewToken(util.Nanoid())
	if err != nil {
		return err
	}
	secret := domain.NewClientSecret(client.Id, token.Hash)
	return c.db.StoreClientSecret(&secret)
}

func (c *Client) ValidateClientSecret(clientId, password string) (bool, error) {
	// TODO get secret by clientId
	secretId := ""
	secret, err := c.db.GetClientSecret(secretId)
	if err != nil {
		return false, err
	}
	return domain.Token{
		Raw:  password,
		Hash: secret.Hash,
	}.Validate(), nil
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

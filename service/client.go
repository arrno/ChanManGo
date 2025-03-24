package service

import "github.com/arrno/chanman/domain"

type ClientProvider interface {
	CreateClient(secret_token string, client domain.Client) error
	GetClient(id string) (domain.Client, error)
	UpdateClient(client domain.Client) error
	ArchiveClient(id string) error
	//
	UpdateClientSecret(secret_token string) error
	ValidateClientSecret(secret_token string) (bool, error)
}

// TODO implement interface

// generate new token / salt / hash

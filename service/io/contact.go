package io

import (
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type ContactProvider interface {
	CreateContactRequest(contactRequest *domain.ContactRequest) error
}

type Contact struct {
	repo datastore.RepoProvider
}

func NewContact(repo datastore.RepoProvider) *Contact {
	return &Contact{repo}
}

func (c *Contact) CreateContactRequest(contactRequest *domain.ContactRequest) error {
	return c.repo.StoreContactRequest(contactRequest)
}

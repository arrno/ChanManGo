package io

import (
	"errors"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type GroupProvider interface {
	CreateGroup(group *domain.Group) error
	UpdateGroup(group *domain.Group) error
	GetGroup(id string) (*domain.Group, error)
	GetGroupByOwner(ownerId string) ([]*domain.Group, error)
	ArchiveGroup(id string) error
	//
	CreateClientGroup(clientId, groupId string) error
	ArchiveClientGroup(clientId, groupId string) error
	//
	InGroup(channelId, clientId string) (bool, error)
}

// TODO implement interface

type Group struct {
	db datastore.DatastoreProvider
}

func NewGroup(db datastore.DatastoreProvider) *Group {
	return &Group{
		db: db,
	}
}

func (g *Group) CreateGroup(group *domain.Group) error {
	exists, err := g.groupExists(group.Id)
	if err != nil {
		return err
	} else if exists {
		return ErrCreateOnExistingRecord
	}
	return g.db.StoreGroup(group)
}

func (g *Group) UpdateGroup(group *domain.Group) error {
	return g.db.UpdateGroup(group)
}
func (g *Group) GetGroup(id string) (*domain.Group, error) {
	return g.db.GetGroup(id)
}
func (g *Group) GetGroupByOwner(ownerId string) ([]*domain.Group, error) {
	return nil, domain.ErrNotImplemented
}
func (g *Group) ArchiveGroup(id string) error {
	return g.db.ArchiveGroup(id)
}

// For now, only one group
func (g *Group) CreateClientGroup(clientId, groupId string) error {
	inGroup, err := g.clientInGroup(clientId)
	if err != nil {
		return err
	} else if inGroup {
		return ErrCreateOnExistingRecord
	}
	return g.db.StoreClientGroup(domain.NewClientGroup(clientId, groupId))
}

func (g *Group) ArchiveClientGroup(clientId, groupId string) error {
	return domain.ErrNotImplemented
}

func (g *Group) InGroup(channelId, clientId string) (bool, error) {
	return false, domain.ErrNotImplemented
}

func (g *Group) groupExists(id string) (bool, error) {
	_, err := g.db.GetGroup(id)
	if err == nil {
		return true, nil
	} else if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (g *Group) clientInGroup(clientId string) (bool, error) {
	groups, err := g.db.GetClientGroupsByClientId(clientId)
	if err == nil {
		return true, nil
	} else if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return len(groups) == 0, err
}

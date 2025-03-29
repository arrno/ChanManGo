package io

import (
	"errors"
	"time"

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
	// Get the client groups
	clientGroups, err := g.db.GetClientGroupsByClientId(clientId)
	if err != nil {
		return err
	}
	
	// Find the specific client-group association
	for _, clientGroup := range clientGroups {
		if clientGroup.GroupId == groupId {
			// Archive the client-group association
			now := time.Now()
			clientGroup.ArchivedAt = &now
			clientGroup.State = domain.STATE_ARCHIVED
			return g.db.UpdateClientGroup(clientGroup)
		}
	}
	
	return datastore.ErrNotFound
}

func (g *Group) InGroup(channelId, clientId string) (bool, error) {
	// Get the channel to find its groupId
	channel, err := g.db.GetChannel(channelId)
	if err != nil {
		return false, err
	}
	
	// Get client groups
	clientGroups, err := g.db.GetClientGroupsByClientId(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	
	// Check if any of the client's groups match the channel's group
	channelGroupId := channel.GroupId
	for _, clientGroup := range clientGroups {
		if clientGroup.GroupId == channelGroupId {
			return true, nil
		}
	}
	
	return false, nil
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
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	
	// If we got groups, the client is in at least one group
	return len(groups) > 0, nil
}

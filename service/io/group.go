package io

import (
	"errors"
	"time"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type GroupProvider interface {
	CreateGroup(group *domain.Group) error
	//
	CreateClientGroup(clientId, groupId string) error
	ArchiveClientGroup(clientId, groupId string) error
	//
	InGroup(channelId, clientId string) (bool, error)
}

// TODO implement interface

type Group struct {
	repo datastore.RepoProvider
}

func NewGroup(repo datastore.RepoProvider) *Group {
	return &Group{
		repo: repo,
	}
}

func (g *Group) CreateGroup(group *domain.Group) error {
	exists, err := g.groupExists(group.Id)
	if err != nil {
		return err
	} else if exists {
		return domain.ErrCreateOnExistingRecord
	}
	return g.repo.StoreGroup(group)
}

// For now, only one group
func (g *Group) CreateClientGroup(clientId, groupId string) error {
	inGroup, err := g.clientInGroup(clientId)
	if err != nil {
		return err
	} else if inGroup {
		return domain.ErrCreateOnExistingRecord
	}
	return g.repo.StoreClientGroup(domain.NewClientGroup(clientId, groupId))
}

func (g *Group) ArchiveClientGroup(clientId, groupId string) error {
	// Get the client groups
	clientGroups, err := g.repo.GetClientGroupsByClientId(clientId)
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
			return g.repo.UpdateClientGroup(clientGroup)
		}
	}

	return datastore.ErrNotFound
}

func (g *Group) InGroup(channelId, clientId string) (bool, error) {
	// Get the channel to find its groupId
	channel, err := g.repo.GetChannel(channelId)
	if err != nil {
		return false, err
	}

	// Get client groups
	clientGroups, err := g.repo.GetClientGroupsByClientId(clientId)
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
	_, err := g.repo.GetGroup(id)
	if err == nil {
		return true, nil
	} else if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (g *Group) clientInGroup(clientId string) (bool, error) {
	groups, err := g.repo.GetClientGroupsByClientId(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	// If we got groups, the client is in at least one group
	return len(groups) > 0, nil
}

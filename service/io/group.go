package io

import "github.com/arrno/chanman/domain"

type GroupProvider interface {
	CreateGroup(ownerId string, group domain.Group) error
	UpdateGroup(group domain.Group) error
	GetGroup(id string) (domain.Group, error)
	GetGroupByOwner(ownerId string) ([]domain.Group, error)
	ArchiveGroup(id string) error
	//
	CreateClientGroup(clientId, groupId string) error
	ArchiveClientGroup(clientId, groupId string) error
	CreateChannelGroup(channelId, groupId string) error
	ArchiveChannelGroup(channelId, groupId string) error
	//
	InGroup(channelId, clientId string) (bool, error)
}

// TODO implement interface

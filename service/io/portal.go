package io

import "github.com/arrno/chanman/domain"

type PortalProvider interface {
	CreatePortal(ownerId string, portal domain.Portal) error
	UpdatePortal(portal domain.Portal) error
	GetPortal(id string) (domain.Portal, error)
	GetPortalByOwner(ownerId string) (domain.Portal, error)
	ArchivePortal(id string) error
}

// TODO implement interface

package io

import (
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

type PortalProvider interface {
	CreatePortal(ownerId string, portal domain.Portal) error
	UpdatePortal(portal domain.Portal) error
	GetPortal(id string) (domain.Portal, error)
	GetPortalByOwner(ownerId string) (domain.Portal, error)
	ArchivePortal(id string) error
}

type Portal struct {
	repo datastore.RepoProvider
}

func NewPortal(repo datastore.RepoProvider) *Portal {
	return &Portal{repo}
}

func (m *Portal) CreatePortal(portal *domain.Portal) error {
	return m.repo.StorePortal(portal)
}
func (m *Portal) GetPortal(id string) (*domain.Portal, error) {
	return m.repo.GetPortal(id)
}
func (m *Portal) UpdatePortal(portal *domain.Portal) error {
	return m.repo.UpdatePortal(portal)
}
func (m *Portal) ArchivePortal(id string) error {
	return m.repo.ArchivePortal(id)
}

package io

import (
	"errors"

	"github.com/arrno/chanman/datastore"
)

var (
	ErrCreateOnExistingRecord = errors.New("Create attempted on record that already exists")
)

type SystemProvider struct {
	ChannelProvider
	ClientProvider
	GroupProvider
	MessageProvider
	PortalProvider
}

func NewSystem(db datastore.DatastoreProvider) SystemProvider {
	return SystemProvider{
		ClientProvider:  NewClient(db),
		MessageProvider: NewMessage(db),
		ChannelProvider: NewChannel(db),
		GroupProvider:   NewGroup(db),
		// Note: PortalProvider is still TODO, as indicated in portal.go
	}
}

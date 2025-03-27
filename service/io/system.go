package io

import "errors"

var (
	ErrCreateOnExistingRecord = errors.New("Create attempted on record that already exists")
)

type SystemProvider interface {
	ClientProvider
	GroupProvider
	PortalProvider
}

func NewSystem() SystemProvider {
	return nil
}

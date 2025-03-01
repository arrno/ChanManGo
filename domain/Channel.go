package domain

import (
	"fmt"
)

type Channel struct {
	Id       string `json:"id"`
	PortalId string `json:"portalId"`
	Name     string `json:"name"`
	State    STATE  `json:"state"`
	Archived bool   `json:"archived"`
}

func NewChannel(portalId, name string) *Channel {
	return &Channel{
		Id:       fmt.Sprintf("%s__%s", portalId, name),
		PortalId: portalId,
		Name:     name,
		State:    STATE_ACTIVE,
	}
}

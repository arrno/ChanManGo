package domain

import (
	"fmt"
	"time"
)

type Channel struct {
	Id         string     `json:"id"`
	PortalId   string     `json:"portalId"`
	GroupId    string     `json:"groupId"`
	Name       string     `json:"name"`
	State      STATE      `json:"state"`
	Archived   bool       `json:"archived"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewChannel(portalId, name string) *Channel {
	return &Channel{
		Id:       fmt.Sprintf("%s__%s", portalId, name),
		PortalId: portalId,
		Name:     name,
		State:    STATE_ACTIVE,
	}
}

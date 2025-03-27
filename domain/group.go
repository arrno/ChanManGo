package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

type Group struct {
	Id            string     `json:"id"`
	ParentGroupId string     `json:"parentGroupId"`
	OwnerId       string     `json:"ownerId"`
	PortalId      string     `json:"portalId"`
	Name          string     `json:"name"`
	State         STATE      `json:"state"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updateAt"`
	ArchivedAt    *time.Time `json:"archivedAt"`
}

func NewGroup(ownerId, portalId, name string) *Group {
	return &Group{
		Id:       PortalId(name, portalId),
		OwnerId:  ownerId,
		PortalId: portalId,
		Name:     name,
		State:    STATE_ACTIVE,
	}
}

type ClientGroup struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	GroupId    string     `json:"groupId"`
	State      STATE      `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewClientGroup(clientId, groupId string) *ClientGroup {
	return &ClientGroup{
		Id:       util.Nanoid(),
		ClientId: clientId,
		GroupId:  groupId,
	}
}

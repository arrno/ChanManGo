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
	UpdatedAt     time.Time  `json:"updatedAt"`
	ArchivedAt    *time.Time `json:"archivedAt"`
}

func NewGroup(ownerId, portalId, name string) (*Group, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Group{
		Id:        CompositeId(name, portalId),
		OwnerId:   ownerId,
		PortalId:  portalId,
		Name:      name,
		State:     STATE_ACTIVE,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type ClientGroup struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	GroupId    string     `json:"groupId"`
	State      STATE      `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewClientGroup(clientId, groupId string) *ClientGroup {
	now := time.Now()
	return &ClientGroup{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		GroupId:   groupId,
		State:     STATE_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (g *Group) UpdateState(state STATE) {
	g.State = state
	g.UpdatedAt = time.Now()
}

func (g *Group) Archive() {
	now := time.Now()
	g.State = STATE_ARCHIVED
	g.ArchivedAt = &now
	g.UpdatedAt = now
}

func (cg *ClientGroup) UpdateState(state STATE) {
	cg.State = state
	cg.UpdatedAt = time.Now()
}

func (cg *ClientGroup) Archive() {
	now := time.Now()
	cg.State = STATE_ARCHIVED
	cg.ArchivedAt = &now
	cg.UpdatedAt = now
}

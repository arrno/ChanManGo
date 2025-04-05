package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

type Group struct {
	Id            string     `db:"id" json:"id"`
	ParentGroupId *string    `db:"parent_group_id" json:"parentGroupId"`
	OwnerId       string     `db:"owner_id" json:"ownerId"`
	PortalId      string     `db:"portal_id" json:"portalId"`
	Name          string     `db:"name" json:"name"`
	State         STATE      `db:"state" json:"state"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt    *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewGroup(id, ownerId, portalId, name string) (*Group, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	if id == "" {
		id = util.Nanoid()
	}
	return &Group{
		Id:        id,
		OwnerId:   ownerId,
		PortalId:  portalId,
		Name:      name,
		State:     STATE_ACTIVE,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type ClientGroup struct {
	Id         string     `db:"id" json:"id"`
	ClientId   string     `db:"client_id" json:"clientId"`
	GroupId    string     `db:"group_id" json:"groupId"`
	State      STATE      `db:"state" json:"state"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time `db:"archived_at" json:"archivedAt"`
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

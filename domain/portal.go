package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

// func CompositeId(selfId, parentId string) string {
// 	return fmt.Sprintf("%s::%s", parentId, selfId)
// }

type Portal struct {
	Id         string     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	OwnerId    string     `db:"owner_id" json:"ownerId"`
	Active     bool       `db:"active" json:"active"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewPortal(id, ownerId, name string) (*Portal, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	if id == "" {
		id = util.Nanoid()
	}
	now := time.Now()
	return &Portal{
		Id:        id,
		Name:      name,
		OwnerId:   ownerId,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Portal) UpdateState(active bool) {
	p.Active = active
	p.UpdatedAt = time.Now()
}

func (p *Portal) Archive() {
	now := time.Now()
	p.Active = false
	p.ArchivedAt = &now
	p.UpdatedAt = now
}

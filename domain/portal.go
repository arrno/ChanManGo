package domain

import (
	"fmt"
	"time"

	"github.com/arrno/chanman/util"
)

func CompositeId(selfId, parentId string) string {
	return fmt.Sprintf("%s::%s", parentId, selfId)
}

type Portal struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	OwnerId    string     `json:"ownerId"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewPortal(ownerId, name string) (*Portal, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	now := time.Now()
	return &Portal{
		Id:        util.Nanoid(),
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

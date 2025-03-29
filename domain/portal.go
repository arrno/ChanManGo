package domain

import (
	"fmt"
	"time"

	"github.com/arrno/chanman/util"
)

func PortalId(id, portalId string) string {
	return fmt.Sprintf("%s::%s", portalId, id)
}

func GroupId(id, groupId, portalId string) string {
	return fmt.Sprintf("%s::%s::%s", portalId, groupId, id)
}

type Portal struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	OwnerId    string     `json:"ownerId"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewPortal(ownerId, name string) (*Portal, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Portal{
		Id:        util.Nanoid(),
		Name:      name,
		OwnerId:   ownerId,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

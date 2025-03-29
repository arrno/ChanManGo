package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

type Client struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	Active     bool       `json:"active"`
	Admin      bool       `json:"admin"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewClient(name, groupId, portalId string, admin bool) (*Client, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Client{
		Id:        GroupId(name, groupId, portalId),
		Name:      name,
		Active:    true,
		Admin:     admin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type ClientSecret struct {
	Id              string     `json:"id"`
	ClientId        string     `json:"clientId"`
	EncryptedSecret string     `json:"encryptedSecret"`
	State           STATE      `json:"state"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	ArchivedAt      *time.Time `json:"archivedAt"`
}

func NewClientSecret(clientId, encryptedSecret string) ClientSecret {
	now := time.Now()
	return ClientSecret{
		Id:              util.Nanoid(),
		ClientId:        clientId,
		EncryptedSecret: encryptedSecret,
		State:           STATE_ACTIVE,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (cs *ClientSecret) UpdateState(state STATE) {
	cs.State = state
	cs.UpdatedAt = time.Now()
}

func (cs *ClientSecret) Archive() {
	now := time.Now()
	cs.State = STATE_ARCHIVED
	cs.ArchivedAt = &now
	cs.UpdatedAt = now
}

func (c *Client) UpdateState(active bool) {
	c.Active = active
	c.UpdatedAt = time.Now()
}

func (c *Client) Archive() {
	now := time.Now()
	c.Active = false
	c.ArchivedAt = &now
	c.UpdatedAt = now
}

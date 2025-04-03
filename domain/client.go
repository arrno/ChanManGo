package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

type Client struct {
	Id         string     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	Active     bool       `db:"active" json:"active"`
	Admin      bool       `db:"admin" json:"admin"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewClient(name, groupId string, admin bool) (*Client, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Client{
		Id:        CompositeId(name, groupId),
		Name:      name,
		Active:    true,
		Admin:     admin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type ClientSecret struct {
	Id              string     `db:"id" json:"id"`
	ClientId        string     `db:"client_id" json:"clientId"`
	EncryptedSecret string     `db:"encrypted_secret" json:"encryptedSecret"`
	Salt            string     `db:"salt" json:"salt"`
	State           STATE      `db:"state" json:"state"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt      *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewClientSecretFromToken(token Token) ClientSecret {
	now := time.Now()
	return ClientSecret{
		Id:              util.Nanoid(),
		ClientId:        token.Identity,
		EncryptedSecret: token.Hashed,
		Salt:            token.Salt,
		State:           STATE_ACTIVE,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// func NewClientSecret(clientId, encryptedSecret, salt string) ClientSecret {
// 	now := time.Now()
// 	return ClientSecret{
// 		Id:              util.Nanoid(),
// 		ClientId:        clientId,
// 		EncryptedSecret: encryptedSecret,
// 		Salt:            salt,
// 		State:           STATE_ACTIVE,
// 		CreatedAt:       now,
// 		UpdatedAt:       now,
// 	}
// }

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

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
	Email      string     `db:"email" json:"email"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewClient(id, name, groupId, email string, admin bool) (*Client, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	if id == "" {
		id = util.Nanoid()
	}
	return &Client{
		Id:        id,
		Name:      name,
		Active:    true,
		Email:     email,
		Admin:     admin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type ClientUsage struct {
	Id       string    `db:"id" json:"id"`
	ClientId string    `db:"client_id" json:"client_id"`
	Start    time.Time `db:"start" json:"start"`
	Stop     time.Time `db:"stop" json:"stop"`
	Count    uint      `db:"count" json:"count"`
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

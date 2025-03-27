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
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewClient(name, portalId string, admin bool) *Client {
	return &Client{
		Id:        PortalId(name, portalId),
		Name:      name,
		Active:    true,
		Admin:     admin,
		CreatedAt: time.Now(),
	}
}

type ClientSecret struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	Hash       string     `json:"hash"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewClientSecret(clientId, hash string) ClientSecret {
	return ClientSecret{
		Id:       util.Nanoid(),
		ClientId: clientId,
		Hash:     hash,
	}
}

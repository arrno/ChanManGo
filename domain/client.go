package domain

import "time"

type Client struct {
	Id         string     `json:"id"`
	Active     bool       `json:"active"`
	Admin      bool       `json:"admin"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

type ClientSecret struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	Hash       string     `json:"hash"`
	Salt       string     `json:"salt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

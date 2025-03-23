package domain

import "time"

type Client struct {
	Id         string     `json:"id"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

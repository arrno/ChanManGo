package domain

import "time"

type Portal struct {
	Id         string     `json:"id"`
	OwnerId    string     `json:"ownerId"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

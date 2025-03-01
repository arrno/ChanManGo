package domain

import "time"

type Group struct {
	Id         string     `json:"id"`
	OwnerId    string     `json:"ownerId"`
	PortalId   string     `json:"portalId"`
	Name       string     `json:"name"`
	State      STATE      `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

type ClientGroup struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	GroupId    string     `json:"groupId"`
	State      STATE      `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

type ChannelGroup struct {
	Id         string     `json:"id"`
	ChannelId  string     `json:"channelId"`
	GroupId    string     `json:"groupId"`
	State      STATE      `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updateAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

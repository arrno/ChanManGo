package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

// ChannelSubscription is a connection from a channel worker to a broker channel
type ChannelSubscription struct {
	Id         string                       `json:"id"`
	ChannelId  string                       `json:"channelId"`
	State      STATE                        `json:"state"`
	Archived   bool                         `json:"archived"`
	Callback   func(message *Message) error `json:"-"`
	CreatedAt  time.Time                    `json:"createdAt"`
	UpdatedAt  time.Time                    `json:"updateAt"`
	ArchivedAt *time.Time                   `json:"archivedAt"`
}

// TopicSubscription is a connection from a client to a channel worker over a subset of topics
type TopicSubscription struct {
	Id         string                       `json:"id"`
	ClientId   string                       `json:"clientId"`
	ChannelId  string                       `json:"channelId"`
	Topics     []string                     `json:"topicName"`
	State      STATE                        `json:"state"`
	Archived   bool                         `json:"archived"`
	Callback   func(message *Message) error `json:"-"`
	CreatedAt  time.Time                    `json:"createdAt"`
	UpdatedAt  time.Time                    `json:"updateAt"`
	ArchivedAt *time.Time                   `json:"archivedAt"`
}

func NewTopicSubscription(clientId, channelId string, topics []string) *TopicSubscription {
	return &TopicSubscription{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		ChannelId: channelId,
		Topics:    topics,
		State:     STATE_ACTIVE,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

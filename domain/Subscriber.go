package domain

import (
	"errors"
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
	UpdatedAt  time.Time                    `json:"updatedAt"`
	ArchivedAt *time.Time                   `json:"archivedAt"`
}

func NewChannelSubscription(channelId string, callback func(message *Message) error) *ChannelSubscription {
	now := time.Now()
	return &ChannelSubscription{
		Id:        util.Nanoid(),
		ChannelId: channelId,
		State:     STATE_ACTIVE,
		Archived:  false,
		Callback:  callback,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TopicSubscription is a connection from a client to a channel worker over a subset of topics
type TopicSubscription struct {
	Id         string                       `json:"id"`
	ClientId   string                       `json:"clientId"`
	ChannelId  string                       `json:"channelId"`
	Topics     []string                     `json:"topics"`
	State      STATE                        `json:"state"`
	Archived   bool                         `json:"archived"`
	Callback   func(message *Message) error `json:"-"`
	CreatedAt  time.Time                    `json:"createdAt"`
	UpdatedAt  time.Time                    `json:"updatedAt"`
	ArchivedAt *time.Time                   `json:"archivedAt"`
}

func NewTopicSubscription(clientId, channelId string, topics []string) (*TopicSubscription, error) {
	// Validate that we have at least one topic
	if len(topics) == 0 {
		return nil, errors.New("At least one topic must be specified")
	}
	
	// Validate each topic
	for _, topic := range topics {
		if topic == "" {
			return nil, errors.New("Empty topic names are not allowed")
		}
	}
	
	now := time.Now()
	return &TopicSubscription{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		ChannelId: channelId,
		Topics:    topics,
		State:     STATE_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

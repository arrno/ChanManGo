package domain

import (
	"sync"
	"time"

	"github.com/arrno/chanman/util"
)

// ChannelSubscription is a connection from a channel worker to a broker channel
type ChannelSubscription struct {
	Id         string                       `db:"id" json:"id"`
	ChannelId  string                       `db:"channel_id" json:"channelId"`
	State      STATE                        `db:"state" json:"state"`
	Archived   bool                         `db:"archived" json:"archived"`
	Callback   func(message *Message) error `json:"-"`
	CreatedAt  time.Time                    `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time                    `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time                   `db:"archived_at" json:"archivedAt"`
}

func NewChannelSubscription(id, channelId string, callback func(message *Message) error) *ChannelSubscription {
	now := time.Now()
	if id == "" {
		id = util.Nanoid()
	}
	return &ChannelSubscription{
		Id:        id,
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
	mu         sync.Mutex                   `json:"-"`
	Id         string                       `db:"id" json:"id"`
	ClientId   string                       `db:"client_id" json:"clientId"`
	ChannelId  string                       `db:"channel_id" json:"channelId"`
	Topics     []string                     `db:"topics" json:"topics"`
	State      STATE                        `db:"state" json:"state"`
	Archived   bool                         `db:"archived" json:"archived"`
	Callback   func(message *Message) error `json:"-"`
	CreatedAt  time.Time                    `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time                    `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time                   `db:"archived_at" json:"archivedAt"`
}

func NewTopicSubscription(id, clientId, channelId string, topics []string) (*TopicSubscription, error) {
	// Validate that we have at least one topic
	if len(topics) == 0 {
		return nil, ErrEmptyTopicList
	}

	// Validate each topic
	for _, topic := range topics {
		if topic == "" {
			return nil, ErrEmptyTopic
		}
	}

	if id == "" {
		id = util.Nanoid()
	}

	now := time.Now()
	return &TopicSubscription{
		Id:        id,
		ClientId:  clientId,
		ChannelId: channelId,
		Topics:    topics,
		State:     STATE_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (t *TopicSubscription) UpdateState(s STATE) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.State = s
	t.UpdatedAt = time.Now()
}

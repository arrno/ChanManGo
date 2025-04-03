package domain

import (
	"sync"
	"time"

	"github.com/arrno/chanman/util"
)

type Message struct {
	mu             sync.Mutex `json:"-"`
	Id             string     `db:"id" json:"id"`
	Topic          string     `db:"topic" json:"topic"`
	ClientId       string     `db:"client_id" json:"clientId"`
	ChannelId      string     `db:"channel_id" json:"channelId"`
	ProcessedAt    *time.Time `db:"processed_at" json:"processedAt"`
	AcknowledgedAt *time.Time `db:"acknowledged_at" json:"acknowledged"`
	State          STATE      `db:"state" json:"state"`
	Archived       bool       `db:"archived" json:"archived"`
	Payload        []byte     `db:"payload" json:"payload"`
	Attempts       uint       `db:"attempts" json:"attempts"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt     *time.Time `db:"archived_at" json:"archivedAt"`
}

const (
	MaxPayloadSizeBytes = 1024 * 1024 // 1MB maximum payload size
)

func NewMessage(clientId, channelId, topic string, payload []byte) (*Message, error) {
	// Validate payload size
	if len(payload) > MaxPayloadSizeBytes {
		return nil, ErrPayloadSizeExceeded
	}

	// Validate topic
	if topic == "" {
		return nil, ErrEmptyTopic
	}

	now := time.Now()
	return &Message{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		ChannelId: channelId,
		Topic:     topic,
		State:     STATE_PENDING,
		Payload:   payload,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (m *Message) Clone() Message {
	return Message{
		Id:             m.Id,
		Topic:          m.Topic,
		ClientId:       m.ClientId,
		ChannelId:      m.ChannelId,
		ProcessedAt:    m.ProcessedAt,
		AcknowledgedAt: m.AcknowledgedAt,
		State:          m.State,
		Archived:       m.Archived,
		Payload:        m.Payload,
		Attempts:       m.Attempts,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		ArchivedAt:     m.ArchivedAt,
	}
}

func (m *Message) UpdateState(s STATE) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = s
}

func (m *Message) Process() {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Now()
	m.State = STATE_PROCESSED
	m.ProcessedAt = &t
	m.Attempts++
}

func (m *Message) Acknowledge() {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Now()
	m.AcknowledgedAt = &t
}

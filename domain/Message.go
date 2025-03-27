package domain

import (
	"sync"
	"time"

	"github.com/arrno/chanman/util"
)

type Message struct {
	mu             sync.Mutex `json:"-"`
	Id             string     `json:"id"`
	Topic          string     `json:"topic"`
	ClientId       string     `json:"clientId"`
	ChannelId      string     `json:"channelId"`
	ProcessedAt    *time.Time `json:"processedAt"`
	AcknowledgedAt *time.Time `json:"acknowledged"`
	State          STATE      `json:"state"`
	Archived       bool       `json:"archived"`
	Payload        []byte     `json:"payload"`
	Attempts       uint       `json:"attempts"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updateAt"`
	ArchivedAt     *time.Time `json:"archivedAt"`
}

func NewMessage(clientId, channelId, topic string, payload []byte) *Message {
	return &Message{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		ChannelId: channelId,
		Topic:     topic,
		State:     STATE_PENDING,
		Payload:   payload,
	}
}

func (m *Message) Clone() Message {
	return Message{
		Id:             m.Id,
		Topic:          m.Topic,
		ClientId:       m.ClientId,
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

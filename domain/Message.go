package domain

import (
	"sync"
	"time"

	"github.com/arrno/chanban/util"
)

type Message struct {
	mu                     sync.Mutex `json:"-"`
	Id                     string     `json:"id"`
	Topic                  string     `json:"topic"`
	ClientId               string     `json:"clientId"`
	ProcessedByWorkerId    string     `json:"processedByWorkerId"`
	ProcessedAt            *time.Time `json:"processedAt"`
	AcknowledgedByWorkerId string     `json:"acknowledgedByWorkerId"`
	AcknowledgedAt         *time.Time `json:"acknowledged"`
	State                  STATE      `json:"state"`
	Archived               bool       `json:"archived"`
	Payload                []byte     `json:"payload"`
	Attempts               uint       `json:"attempts"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updateAt"`
	ArchivedAt             *time.Time `json:"archivedAt"`
}

func NewMessage(clientId, topic string, payload []byte) *Message {
	return &Message{
		Id:       util.Nanoid(),
		ClientId: clientId,
		Topic:    topic,
		State:    STATE_PENDING,
		Payload:  payload,
	}
}

func (m *Message) Clone() Message {
	return Message{
		Id:                     m.Id,
		ClientId:               m.ClientId,
		ProcessedByWorkerId:    m.ProcessedByWorkerId,
		ProcessedAt:            m.ProcessedAt,
		AcknowledgedByWorkerId: m.AcknowledgedByWorkerId,
		AcknowledgedAt:         m.AcknowledgedAt,
		State:                  m.State,
		Archived:               m.Archived,
		Payload:                m.Payload,
		Attempts:               m.Attempts,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		ArchivedAt:             m.ArchivedAt,
	}
}

func (m *Message) UpdateState(s STATE) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = s
}

func (m *Message) Process(w *Worker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Now()
	m.ProcessedByWorkerId = w.Id
	m.State = STATE_PROCESSED
	m.ProcessedAt = &t
	m.Attempts++
}

func (m *Message) Acknowledge(w *Worker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Now()
	m.AcknowledgedByWorkerId = w.Id
	m.AcknowledgedAt = &t
}

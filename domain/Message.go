package domain

import (
	"sync"
	"time"

	"github.com/arrno/chanban/util"
)

type Message struct {
	mu             sync.Mutex
	Id             string     `json:"id"`
	ClientId       string     `json:"clientId"`
	AcknowledgedAt *time.Time `json:"acknowledged"`
	State          STATE      `json:"state"`
	Archived       bool       `json:"archived"`
	Payload        []byte     `json:"payload"`
}

func NewMessage(clientId string, payload []byte) *Message {
	return &Message{
		Id:       util.Nanoid(),
		ClientId: clientId,
		State:    PENDING,
		Payload:  payload,
	}
}

func (m *Message) UpdateState(s STATE) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = s
}

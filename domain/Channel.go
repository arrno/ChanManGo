package domain

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanman/util"
)

type Operation int

const (
	SUBSCRIBE Operation = iota
	PUBLISH
	ARCHIVE
)

var (
	ErrChanOpNotPermitted = errors.New("Client action denied on channel.")
)

type Channel struct {
	mu              sync.RWMutex `json:"-"`
	Id              string       `json:"id"`
	PortalId        string       `json:"portalId"`
	GroupId         string       `json:"groupId"`
	Name            string       `json:"name"`
	State           STATE        `json:"state"`
	Archived        bool         `json:"archived"`
	IntervalSeconds uint         `json:"intervalSeconds"`
	Heartbeat       *time.Time   `json:"heartbeat"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updateAt"`
	ArchivedAt      *time.Time   `json:"archivedAt"`
}

func NewChannel(portalId, groupId, name string, hbIntervalSec uint) (*Channel, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Channel{
		Id:              GroupId(name, groupId, portalId),
		GroupId:         groupId,
		PortalId:        portalId,
		Name:            name,
		State:           STATE_RUNNING,
		IntervalSeconds: hbIntervalSec,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (w *Channel) UpdateHeartbeat() {
	w.mu.Lock()
	defer w.mu.Unlock()
	t := time.Now()
	w.Heartbeat = &t
}

func (w *Channel) Repr() string {
	return fmt.Sprintf("{ WORKER: %s, STATE: %s, ARCHIVED: %t, INTERVAL: %d }", w.Id, w.State, w.Archived, w.IntervalSeconds)
}

func (m *Channel) UpdateState(s STATE) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = s
}

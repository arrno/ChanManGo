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
	Id              string       `db:"id" json:"id"`
	PortalId        string       `db:"portal_id" json:"portalId"`
	GroupId         string       `db:"group_id" json:"groupId"`
	Name            string       `db:"name" json:"name"`
	State           STATE        `db:"state" json:"state"`
	Archived        bool         `db:"archived" json:"archived"`
	IntervalSeconds uint         `db:"interval_seconds" json:"intervalSeconds"`
	Heartbeat       *time.Time   `db:"heartbeat" json:"heartbeat"`
	CreatedAt       time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time    `db:"updated_at" json:"updatedAt"`
	ArchivedAt      *time.Time   `db:"archived_at" json:"archivedAt"`
}

func NewChannel(portalId, groupId, name string, hbIntervalSec uint) (*Channel, error) {
	if !util.IsValidName(name) {
		return nil, ErrInvalidName
	}
	return &Channel{
		Id:              CompositeId(name, groupId),
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

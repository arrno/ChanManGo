package domain

import (
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanban/util"
)

type Worker struct {
	mu              sync.RWMutex
	Id              string     `json:"id"`
	Channel         *Channel   `json:"channel"`
	State           STATE      `json:"state"`
	Archived        bool       `json:"archived"`
	IntervalSeconds uint       `json:"intervalSeconds"`
	Heartbeat       *time.Time `json:"heartbeat"`
}

func NewWorker(channel *Channel, hbIntervalSec uint) *Worker {
	return &Worker{
		Id:              util.Nanoid(),
		Channel:         channel,
		State:           STATE_RUNNING,
		IntervalSeconds: hbIntervalSec,
	}
}

func (w *Worker) UpdateHeartbeat() {
	w.mu.Lock()
	defer w.mu.Unlock()
	t := time.Now()
	w.Heartbeat = &t
}

func (w *Worker) Repr() string {
	return fmt.Sprintf("{ WORKER: %s, STATE: %s, ARCHIVED: %t, INTERVAL: %d }", w.Id, w.State, w.Archived, w.IntervalSeconds)
}

func (m *Worker) UpdateState(s STATE) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = s
}

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
		State:           PENDING,
		IntervalSeconds: hbIntervalSec,
	}
}
func (w *Worker) RLock() func() {
	w.mu.RLock()
	return func() {
		defer w.mu.RUnlock()
	}
}
func (w *Worker) Repr() string {
	return fmt.Sprintf("{ WORKER: %s, STATE: %s, ARCHIVED: %t, INTERVAL: %d }", w.Id, w.State, w.Archived, w.IntervalSeconds)
}

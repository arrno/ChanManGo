package service

import (
	"context"
	"sync"
	"time"

	"github.com/arrno/chanban/broker"
	"github.com/arrno/chanban/datastore"
	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/util"
)

type IWorker interface {
	Listen() error // open subscription... on message, broadcast
	Destroy() error
	Publish(*domain.Message) error
	Heartbeat(done chan any, f func(*domain.Worker))
	AddSubscription(subscriber *domain.Subscription) bool
	RemoveSubscription(id string) bool
}

// One worker per channel
type Worker struct {
	mu               sync.Mutex
	worker           *domain.Worker
	db               datastore.IDatastore
	broker           broker.IBroker
	rootSubscription *domain.Subscription
	subscribers      map[string]*domain.Subscription
}

func NewWorker(state *domain.Worker, db datastore.IDatastore, broker broker.IBroker) *Worker {
	w := Worker{
		worker: state,
		broker: broker,
		db:     db,
		rootSubscription: &domain.Subscription{
			Id:        util.Nanoid(),
			ClientId:  state.Channel.PortalId,
			TopicName: state.Channel.TopicName,
			State:     domain.PENDING,
			Archived:  false,
			Callback:  nil,
		},
		subscribers: map[string]*domain.Subscription{},
	}
	// Broadcast
	w.rootSubscription.Callback = func(message *domain.Message) (err error) {
		defer func() {
			if err != nil {
				message.UpdateState(domain.ERROR)
			} else {
				message.UpdateState(domain.COMPLETE)
			}
			w.db.UpdateMessage(message)

		}()
		var wg sync.WaitGroup
		wg.Add(len(w.subscribers))
		errChan := make(chan error, 1)
		done := make(chan any)
		for _, sub := range w.subscribers {
			go func() {
				defer wg.Done()
				if err := sub.Callback(message); err != nil {
					select {
					case errChan <- err:
						close(done)
					case <-done:
						return
					}
				}
			}()
		}
		wg.Wait()
		select {
		case err = <-errChan:
			return
		default:
		}
		return
	}
	return &w
}
func (w *Worker) State() *domain.Worker {
	return w.worker
}
func (w *Worker) Publish(message *domain.Message) error {
	return w.broker.PublishMessage(context.Background(), w.worker.Channel, message)
}

func (w *Worker) Listen() (err error) {
	defer func() {
		if err != nil {
			w.worker.State = domain.ERROR
		} else {
			w.worker.State = domain.RUNNING
		}
	}()
	err = w.broker.SubscribeForMessages(
		context.Background(),
		w.worker.Channel,
		w.rootSubscription,
	)
	return
}

func (w *Worker) Destroy() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.worker.State == domain.INACTIVE {
		return nil
	}
	defer func() {
		if err != nil {
			w.worker.State = domain.ERROR
		} else {
			w.worker.State = domain.INACTIVE
		}
	}()
	w.subscribers = nil
	err = w.broker.UnsubscribeForMessages(
		context.Background(),
		w.worker.Channel,
		w.rootSubscription,
	)
	return
}

func (w *Worker) AddSubscription(subscriber *domain.Subscription) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.worker.State != domain.RUNNING {
		return false
	}
	if _, ok := w.subscribers[subscriber.Id]; ok {
		return false
	}
	w.subscribers[subscriber.Id] = subscriber
	return true
}

func (w *Worker) RemoveSubscription(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.worker.State != domain.RUNNING {
		return false
	}
	if _, ok := w.subscribers[id]; !ok {
		return false
	}
	delete(w.subscribers, id)
	return true
}

func (w *Worker) Heartbeat(done chan any, f func(*domain.Worker)) {
	started := make(chan any)
	go func() {
		close(started)
		for {
			select {
			case <-done:
				return
			case <-time.After(time.Duration(w.worker.IntervalSeconds) * time.Second):
				func() {
					f(w.worker)
				}()
			}
		}
	}()
	<-started
}

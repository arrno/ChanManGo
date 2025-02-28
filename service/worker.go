package service

import (
	"context"
	"sync"
	"time"

	"github.com/arrno/chanban/broker"
	"github.com/arrno/chanban/domain"
)

type WorkerProvider interface {
	State() *domain.Worker
	Listen() error // open subscription... on message, broadcast
	Destroy() error
	Reboot() error
	Idle() bool
	Publish(*domain.Message) error
	Heartbeat(done chan any, f func(*domain.Worker))
	IncrPublishers() func()
	AddSubscription(subscriber *domain.TopicSubscription) bool
	RemoveSubscription(topics []string, id string)
}

// One worker per channel
type Worker struct {
	mu               sync.Mutex
	worker           *domain.Worker
	broker           broker.BrokerProvider
	rootSubscription *domain.ChannelSubscription
	subscriptionBook AddressBookProvider
	publishers       uint
}

func NewWorker(state *domain.Worker, broker broker.BrokerProvider, update func(message *domain.Message) error) *Worker {
	w := Worker{
		worker: state,
		broker: broker,
		rootSubscription: &domain.ChannelSubscription{
			Id:       state.Id,
			PortalId: state.Channel.PortalId,
			State:    domain.PENDING,
			Archived: false,
			Callback: nil,
		},
		subscriptionBook: NewAddressBook(state.Id),
	}
	// Broadcast
	w.rootSubscription.Callback = func(message *domain.Message) (err error) {

		message.Acknowledge(w.worker)
		defer update(message)

		var subscribers []*domain.TopicSubscription
		if message.Topic != "all" {
			subscribers = w.subscriptionBook.FetchTopicSubscriptions([]string{"all"})
		}
		for _, v := range w.subscriptionBook.FetchTopicSubscriptions([]string{message.Topic}) {
			subscribers = append(subscribers, v)
		}

		var wg sync.WaitGroup
		wg.Add(len(subscribers))

		errChan := make(chan error, 1)
		done := make(chan any)
		for _, sub := range subscribers {
			go func() {
				defer wg.Done()
				messageClone := message.Clone() // safe
				if err := sub.Callback(&messageClone); err != nil {
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
	message.Process(w.worker)
	return w.broker.PublishMessage(context.Background(), w.worker.Channel, message)
}

func (w *Worker) IncrPublishers() func() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.publishers++
	return func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.publishers--
	}
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
	w.subscriptionBook.Clear()
	err = w.broker.UnsubscribeForMessages(
		context.Background(),
		w.worker.Channel,
		w.rootSubscription,
	)
	return
}
func (w *Worker) Reboot() (err error) {
	return nil
}

func (w *Worker) Idle() bool {
	return w.publishers == 0 && w.subscriptionBook.Size() == 0
}

func (w *Worker) AddSubscription(subscriber *domain.TopicSubscription) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.worker.State != domain.RUNNING {
		return false
	}
	w.subscriptionBook.CacheSubscription(subscriber)
	return true
}

func (w *Worker) RemoveSubscription(topics []string, id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.worker.State != domain.RUNNING {
		return
	}
	w.subscriptionBook.UnCacheWorkerSubscription(topics, id)
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

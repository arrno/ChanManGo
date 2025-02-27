package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanban/broker"
	"github.com/arrno/chanban/datastore"
	"github.com/arrno/chanban/domain"
	"github.com/rs/zerolog/log"
)

type ManagerProvider interface {
	CreateChannel(channel domain.Channel)
	DestroyChannel(id string)
	CreateWorker(channel *domain.Channel) (*Worker, chan any, error)
	DestroyWorker(worker *domain.Worker) (err error)
	AddSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error)
	RemoveSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error)
	PublishMessage(channel *domain.Channel, message *domain.Message) (err error)
}

type Manager struct {
	cleanupIntervalSec uint
	db                 datastore.DatastoreProvider
	broker             broker.BrokerProvider
	state              controllerState
}

type controllerState struct {
	mu                sync.RWMutex
	state             domain.STATE
	workers           map[*domain.Channel]WorkerProvider
	destroyDependents map[*domain.Worker][]chan any
}

func NewManager() *Manager {
	// broker / dn
	return new(Manager)
}

func (c *Manager) cannelExistsTx(tx datastore.DatastoreProvider, channel *domain.Channel) (bool, error) {
	_, err := tx.GetChannel(channel.Id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (c *Manager) ExpectChannel(channel *domain.Channel) (WorkerProvider, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	if worker, ok := c.state.workers[channel]; ok {
		return worker, nil
	}

	var channelExists bool
	var err error
	err = c.db.WithTx(func(tx datastore.DatastoreProvider) error {
		channelExists, err = c.cannelExistsTx(tx, channel)
		if err != nil {
			return err
		}
		if channelExists {
			return nil
		}
		if err := c.db.StoreChannel(channel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !channelExists {
		if err := c.broker.CreateChannel(channel); err != nil {
			return nil, err
		}
		return nil, nil
	}

	worker, done, err := c.CreateWorker(channel)
	if err != nil {
		return nil, err
	}
	// add worker to cache
	c.state.workers[channel] = worker
	c.state.destroyDependents[worker.State()] = append(c.state.destroyDependents[worker.State()], done)
	return worker, nil
}

func (c *Manager) RemoveSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.RemoveSubscription(subscriber.Id), nil
}

func (c *Manager) AddSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.AddSubscription(subscriber), nil
}

func (c *Manager) PublishMessage(channel *domain.Channel, message *domain.Message) (err error) {
	defer func() {
		if err != nil {
			message.State = domain.ERROR
		}
		if storeErr := c.db.StoreMessage(message); storeErr != nil {
			err = fmt.Errorf("%w. %w", err, storeErr)
		}
	}()
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return err
	}
	return worker.Publish(message)
}

func (c *Manager) handleHeartbeat(worker *domain.Worker) {
	worker.UpdateHeartbeat()
	err := c.db.UpdateWorker(worker)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Worker heartbeat failed to update: %s. Err: %s ", worker.Repr(), err.Error()))
	}
}

func (c *Manager) SpawnWatcher() chan any {
	destroy := make(chan any)
	go func() {
		for {
			select {
			case <-destroy:
				return
			case <-time.After(time.Duration(c.cleanupIntervalSec) * time.Second):
				// check worker state/heartbeats, check message failures, check channel states
				// fix broken workers/messages
				// when there are no publishers and no subscribers to a channel, prune worker from memory
			}
		}
	}()
	return destroy
}

func (c *Manager) Destroy() {
	c.state.mu.Lock()
	defer func() {
		c.state.state = domain.INACTIVE
		c.state.mu.Unlock()
	}()
	for _, chans := range c.state.destroyDependents {
		for _, ch := range chans {
			close(ch)
		}
	}
}

func (c *Manager) CreateWorker(channel *domain.Channel) (*Worker, chan any, error) {
	ack := func(message *domain.Message) (err error) {
		if err != nil {
			message.UpdateState(domain.ERROR)
		} else {
			message.UpdateState(domain.COMPLETE)
		}
		return c.db.UpdateMessage(message)
	}
	worker := NewWorker(domain.NewWorker(channel, 60), c.broker, ack)
	if err := worker.Listen(); err != nil {
		return nil, nil, err
	}
	done := make(chan any)
	worker.Heartbeat(done, c.handleHeartbeat)
	// store channel + worker in db
	if err := c.db.StoreWorker(worker.State()); err != nil {
		close(done)
		return nil, nil, err
	}
	return worker, done, nil
}

func (c *Manager) DestroyWorker(worker *domain.Worker) (err error) {
	if worker == nil {
		return
	}
	c.state.mu.Lock()
	svc := c.state.workers[worker.Channel]
	chans := c.state.destroyDependents[worker]
	defer func() {
		c.state.mu.Unlock()
		// worker is dismounted from controller. Exit mutex
		// terminate worker
		for _, ch := range chans {
			close(ch)
		}
		err = svc.Destroy()
		if err != nil {
			worker.UpdateState(domain.ERROR)
		} else {
			worker.UpdateState(domain.INACTIVE)
		}
		if writeErr := c.db.UpdateWorker(worker); writeErr != nil {
			log.Error().Msg(fmt.Sprintf("Failed to write update to worker. Err: %s", writeErr.Error()))
		}
	}()
	// update state
	delete(c.state.workers, worker.Channel)
	delete(c.state.destroyDependents, worker)
	return
}

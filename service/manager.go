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
	ExpectChannel(channel *domain.Channel) (WorkerProvider, error)
	CreateChannel(channel domain.Channel)
	DestroyChannel(id string)
	CreateWorker(channel *domain.Channel) (*Worker, chan any, error)
	DestroyWorker(worker *domain.Worker) (err error)
	AddSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) (bool, error)
	RemoveSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) error
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

func (c *Manager) RemoveSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) error {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return err
	}
	worker.RemoveSubscription(subscriber.Topics, subscriber.Id)
	return nil
}

func (c *Manager) AddSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.AddSubscription(subscriber), nil
}

func (c *Manager) PublishMessage(channel *domain.Channel, message *domain.Message) (err error) {
	defer func() {
		if err != nil {
			message.State = domain.STATE_ERROR
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
				if err := c.HealChannels(); err != nil {
					log.Error().Msg("Failed to heal channels. Err: " + err.Error())
				}
				if err := c.HealWorkers(); err != nil {
					log.Error().Msg("Failed to heal workers. Err: " + err.Error())
				}
				if err := c.HealMessages(); err != nil {
					log.Error().Msg("Failed to heal messages. Err: " + err.Error())
				}
				if err := c.PruneWorkers(); err != nil {
					log.Error().Msg("Failed to prune workers. Err: " + err.Error())
				}
			}
		}
	}()
	return destroy
}

func (c *Manager) Destroy() {
	c.state.mu.Lock()
	defer func() {
		c.state.state = domain.STATE_INACTIVE
		c.state.mu.Unlock()
	}()
	for _, chans := range c.state.destroyDependents {
		for _, ch := range chans {
			close(ch)
		}
	}
}

func (c *Manager) CreateWorker(channel *domain.Channel) (*Worker, chan any, error) {
	update := func(message *domain.Message) (err error) {
		if err != nil {
			message.UpdateState(domain.STATE_ERROR)
		} else {
			message.UpdateState(domain.STATE_COMPLETE)
		}
		return c.db.UpdateMessage(message)
	}
	worker := NewWorker(domain.NewWorker(channel, 60), c.broker, update)
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
			worker.UpdateState(domain.STATE_ERROR)
		} else {
			worker.UpdateState(domain.STATE_INACTIVE)
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

// TODO
func (c *Manager) HealWorkers() error {
	// fetch all workers owned by this manager that errored or stopped heartbeat
	// reboot workers
	return nil
}

// TODO
func (c *Manager) HealMessages() error {
	// fetch all messages processed by owned workers that failed
	// reprocess all messages if attempts < max
	return nil
}
func (c *Manager) HealChannels() error {
	// channels are abstract... they can't really be broken
	return nil
}
func (c *Manager) PruneWorkers() (err error) {
	for worker := range c.workersReport() {
		if worker.Idle() {
			if desErr := c.DestroyWorker(worker.State()); desErr != nil {
				err = desErr
			}
		}
	}
	return
}

func (c *Manager) workersReport() <-chan WorkerProvider {
	workers := make(chan WorkerProvider)
	go func() {
		defer close(workers)
		for _, w := range c.state.workers {
			workers <- w
		}
	}()
	return workers
}

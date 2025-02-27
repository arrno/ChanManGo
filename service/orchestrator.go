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

type IOrchestrator interface {
	CreateChannel(channel domain.Channel)
	DestroyChannel(id string)
	AddSubscription(subscriber *domain.Subscription) bool
	PublishMessage(channel domain.Channel, message domain.Message) error
}

type Orchestrator struct {
	cleanupIntervalSec uint
	db                 datastore.IDatastore
	broker             broker.IBroker
	state              controllerState
}

type controllerState struct {
	mu                sync.RWMutex
	state             domain.STATE
	workers           map[*domain.Channel]IWorker
	destroyDependents map[*domain.Worker][]chan any
}

func NewOrchestrator() *Orchestrator {
	// broker / dn
	return new(Orchestrator)
}

func (c *Orchestrator) cannelExistsTx(tx datastore.IDatastore, channel *domain.Channel) (bool, error) {
	_, err := tx.GetChannel(channel.Id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, datastore.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (c *Orchestrator) ExpectChannel(channel *domain.Channel) (IWorker, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	if worker, ok := c.state.workers[channel]; ok {
		return worker, nil
	}

	var channelExists bool
	var err error
	err = c.db.WithTx(func(tx datastore.IDatastore) error {
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

	// spin up worker
	// worker listen
	// worker heartbeat
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
		return nil, err
	}
	done := make(chan any)
	worker.Heartbeat(done, c.handleHeartbeat)
	// store channel + worker in db
	c.db.StoreWorker(worker.State())
	// add worker to cache
	c.state.workers[channel] = worker
	c.state.destroyDependents[worker.State()] = append(c.state.destroyDependents[worker.State()], done)
	return worker, nil
}

func (c *Orchestrator) RemoveSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.RemoveSubscription(subscriber.Id), nil
}

func (c *Orchestrator) AddSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.AddSubscription(subscriber), nil
}

func (c *Orchestrator) PublishMessage(channel *domain.Channel, message *domain.Message) (err error) {
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

func (c *Orchestrator) handleHeartbeat(worker *domain.Worker) {
	worker.UpdateHeartbeat()
	err := c.db.UpdateWorker(worker)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Worker heartbeat failed to update: %s. Err: %s ", worker.Repr(), err.Error()))
	}
}

func (c *Orchestrator) SpawnWatcher() chan any {
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

func (c *Orchestrator) Destroy() {
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

func (c *Orchestrator) DestroyWorker(worker *domain.Worker) (err error) {
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

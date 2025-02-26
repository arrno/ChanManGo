package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanban/broker"
	"github.com/arrno/chanban/datastore"
	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/service"
	"github.com/arrno/chanban/util"
	"github.com/rs/zerolog/log"
)

type IController interface {
	CreateChannel(channel domain.Channel)
	DestroyChannel(id string)
	AddSubscription(subscriber *domain.Subscription) bool
	PublishMessage(channel domain.Channel, message domain.Message) error
}

type Controller struct {
	cleanupIntervalSec uint
	db                 datastore.IDatastore
	broker             broker.IBroker
	state              controllerState
}

type controllerState struct {
	mu                sync.RWMutex
	state             domain.STATE
	workers           map[*domain.Channel]service.IWorker
	destroyDependents []chan any
}

func NewController() *Controller {
	// broker / dn
	return new(Controller)
}

func (c *Controller) ExpectChannel(channel *domain.Channel) (service.IWorker, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	if worker, ok := c.state.workers[channel]; ok {
		return worker, nil
	}

	channelExists := false // TODO check datastore
	if !channelExists {
		// init channel in DB
		// init channel in broker
		_, err := util.InParallel(
			func() (any, error) {
				if err := c.broker.CreateChannel(channel); err != nil {
					return nil, err
				}
				return nil, nil
			},
			func() (any, error) {
				if err := c.db.StoreChannel(channel); err != nil {
					return nil, err
				}
				return nil, nil
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// spin up worker
	// worker listen
	// worker heartbeat
	worker := service.NewWorker(domain.NewWorker(channel, 60), c.broker)
	if err := worker.Listen(); err != nil {
		return nil, err
	}
	done := make(chan any)
	worker.Heartbeat(done, c.handleHeartbeat)
	// store channel + worker in db
	c.db.StoreWorker(worker.State())
	// add worker to cache
	c.state.workers[channel] = worker
	c.state.destroyDependents = append(c.state.destroyDependents, done)
	return worker, nil
}

func (c *Controller) DestroyChannel() {
	// destroy worker
	// remove channel in broker
	// close worker heartbeat
	// update channel + worker in db
	// remove worker/hb from cache
}

func (c *Controller) RemoveSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.RemoveSubscription(subscriber.Id), nil
}

func (c *Controller) AddSubscription(subscriber *domain.Subscription, channel *domain.Channel) (bool, error) {
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.AddSubscription(subscriber), nil
}

func (c *Controller) PublishMessage(channel *domain.Channel, message *domain.Message) (err error) {
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

func (c *Controller) handleHeartbeat(worker *domain.Worker) {
	err := c.db.UpdateWorker(worker)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Worker heartbeat failed to update: %s. Err: %s ", worker.Repr(), err.Error()))
	}
}

func (c *Controller) SpawnWatcher() chan any {
	destroy := make(chan any)
	go func() {
		for {
			select {
			case <-destroy:
				return
			case <-time.After(time.Duration(c.cleanupIntervalSec) * time.Second):
				// check worker state/heartbeats, check message failures, check channel states
				// when there are no publishers and no subscribers to a channel, archive the worker
			}
		}
	}()
	return destroy
}

func (c *Controller) Destroy() {
	c.state.mu.Lock()
	defer func() {
		c.state.state = domain.INACTIVE
		c.state.mu.Unlock()
	}()
	for _, dependent := range c.state.destroyDependents {
		close(dependent)
	}
}

func (c *Controller) HandleArchiveChannel() {}

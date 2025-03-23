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
	Destroy()
	ExpectChannel(channel *domain.Channel) (ChannelProvider, error)
	CreateChannel(channel *domain.Channel) (*Channel, chan any, error)
	DestroyChannel(worker *domain.Channel) (err error)
	AddSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) (bool, error)
	RemoveSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) error
	PublishMessage(clientId string, channel *domain.Channel, message *domain.Message) (err error)
}

type Manager struct {
	cleanupIntervalSec uint
	db                 datastore.DatastoreProvider
	broker             broker.BrokerProvider
	state              managerState
	destroy            chan any
}

type managerState struct {
	mu                sync.RWMutex
	state             domain.STATE
	channels          map[*domain.Channel]ChannelProvider
	destroyDependents map[*domain.Channel][]chan any
}

func NewManager() *Manager {
	m := Manager{
		broker:             broker.NewInMemoryBroker(),
		db:                 datastore.NewInMemoryDB(),
		cleanupIntervalSec: 600,
		state: managerState{
			state:             domain.STATE_RUNNING,
			channels:          map[*domain.Channel]ChannelProvider{},
			destroyDependents: map[*domain.Channel][]chan any{},
		},
	}
	m.destroy = m.SpawnWatcher()
	return &m
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

func (c *Manager) ExpectChannel(channel *domain.Channel) (ChannelProvider, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()

	if worker, ok := c.state.channels[channel]; ok {
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

	worker, done, err := c.CreateChannel(channel)
	if err != nil {
		return nil, err
	}

	// add worker to cache
	c.state.channels[channel] = worker
	c.state.destroyDependents[worker.Object()] = append(c.state.destroyDependents[worker.Object()], done)

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

	if !c.Permitted(channel.Id, subscriber.ClientId, domain.SUBSCRIBE) {
		return false, domain.ErrChanOpNotPermitted
	}

	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return false, err
	}
	return worker.AddSubscription(subscriber), nil
}

func (c *Manager) PublishMessage(clientId string, channel *domain.Channel, message *domain.Message) (err error) {

	if !c.Permitted(channel.Id, clientId, domain.PUBLISH) {
		return domain.ErrChanOpNotPermitted
	}

	defer func() {
		if err != nil {
			message.UpdateState(domain.STATE_ERROR)
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

func (c *Manager) handleHeartbeat(worker *domain.Channel) {
	worker.UpdateHeartbeat()
	err := c.db.UpdateChannel(worker)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Channel heartbeat failed to update: %s. Err: %s ", worker.Repr(), err.Error()))
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
				if err := c.HealMessages(); err != nil {
					log.Error().Msg("Failed to heal messages. Err: " + err.Error())
				}
				if err := c.PruneChannels(); err != nil {
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
	close(c.destroy)

}

func (c *Manager) CreateChannel(channel *domain.Channel) (*Channel, chan any, error) {
	update := func(message *domain.Message) (err error) {
		if err != nil {
			message.UpdateState(domain.STATE_ERROR)
		} else {
			message.UpdateState(domain.STATE_COMPLETE)
		}
		return c.db.UpdateMessage(message)
	}

	worker := NewChannel(channel, c.broker, update)
	if err := worker.Listen(); err != nil {
		return nil, nil, err
	}

	done := make(chan any)
	worker.Heartbeat(done, c.handleHeartbeat)
	// store channel + worker in db

	if err := c.db.StoreChannel(worker.Object()); err != nil {
		close(done)
		return nil, nil, err
	}
	return worker, done, nil
}

func (c *Manager) DestroyChannel(channel *domain.Channel) (err error) {
	if channel == nil {
		return
	}

	c.state.mu.Lock()
	svc := c.state.channels[channel]
	chans := c.state.destroyDependents[channel]

	defer func() {
		c.state.mu.Unlock()
		// worker is dismounted from controller. Exit mutex
		// terminate worker
		for _, ch := range chans {
			close(ch)
		}
		err = svc.Destroy()
		if err != nil {
			channel.UpdateState(domain.STATE_ERROR)
		} else {
			channel.UpdateState(domain.STATE_INACTIVE)
		}
		if writeErr := c.db.UpdateChannel(channel); writeErr != nil {
			log.Error().Msg(fmt.Sprintf("Failed to write update to worker. Err: %s", writeErr.Error()))
		}
	}()

	// update state
	delete(c.state.channels, channel)
	delete(c.state.destroyDependents, channel)
	return
}

// TODO
func (c *Manager) HealChannels() error {
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

func (c *Manager) PruneChannels() (err error) {
	for worker := range c.workersReport() {
		if worker.Idle() {
			if desErr := c.DestroyChannel(worker.Object()); desErr != nil {
				err = desErr
			}
		}
	}
	return
}

// TODO
func (c *Manager) Permitted(channelId, clientId string, operation domain.Operation) bool {
	return true
}

func (c *Manager) workersReport() <-chan ChannelProvider {
	workers := make(chan ChannelProvider)
	go func() {
		defer close(workers)
		for _, w := range c.state.channels {
			workers <- w
		}
	}()
	return workers
}

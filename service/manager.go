package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanman/broker"
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

type ManagerProvider interface {
	Info() managerSummary
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
	object             managerObject
	destroy            chan any
	debugMode          bool
}

type managerObject struct {
	mu                sync.RWMutex
	state             domain.STATE
	channels          map[string]ChannelProvider // map[channelId]ChannelProvider
	destroyDependents map[string][]chan any      // map[channel.Id][]chan any
}

func NewManager() *Manager {
	m := Manager{
		broker:             broker.NewInMemoryBroker(),
		db:                 datastore.NewInMemoryDB(),
		cleanupIntervalSec: 60,
		// cleanupIntervalSec: 600,
		debugMode: true,
		object: managerObject{
			state:             domain.STATE_RUNNING,
			channels:          map[string]ChannelProvider{},
			destroyDependents: map[string][]chan any{},
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
	c.object.mu.Lock()
	defer c.object.mu.Unlock()

	if worker, ok := c.object.channels[channel.Id]; ok {
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
		log.Info().Msg("Creating channel n DB: " + channel.Id)
		if err := c.db.StoreChannel(channel); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !channelExists {
		log.Info().Msg("Creating channel in broker: " + channel.Id)
		if err := c.broker.CreateChannel(channel); err != nil {
			return nil, err
		}
	}

	log.Info().Msg("Creating channel in memory: " + channel.Id)
	worker, done, err := c.CreateChannel(channel)
	if err != nil {
		return nil, err
	}

	// add worker to cache
	c.object.channels[channel.Id] = worker
	c.object.destroyDependents[worker.Object().Id] = append(c.object.destroyDependents[worker.Object().Id], done)

	return worker, nil
}

func (c *Manager) RemoveSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) error {
	log.Info().Msg(fmt.Sprintf("Removing subscription %s for client %s from channel %s", subscriber.Id, subscriber.ClientId, channel.Id))
	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return err
	}

	worker.RemoveSubscription(subscriber.Topics, subscriber.Id)
	return nil
}

func (c *Manager) AddSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) (bool, error) {
	log.Info().Msg(fmt.Sprintf("Adding subscription %s for client %s from channel %s", subscriber.Id, subscriber.ClientId, channel.Id))
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
	log.Info().Msg(fmt.Sprintf("Publishing message %s from client %s to channel %s", message.Id, clientId, channel.Id))
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
				if c.debugMode {
					c.Info().Print()
				}
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
	log.Info().Msg("Destroying ChanMan")
	c.object.mu.Lock()

	defer func() {
		c.object.state = domain.STATE_INACTIVE
		c.object.mu.Unlock()
	}()

	for _, chans := range c.object.destroyDependents {
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
	log.Info().Msg("destroying channel " + channel.Id)
	if channel == nil {
		return
	}

	c.object.mu.Lock()
	svc := c.object.channels[channel.Id]
	chans := c.object.destroyDependents[channel.Id]

	defer func() {
		c.object.mu.Unlock()
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
	delete(c.object.channels, channel.Id)
	delete(c.object.destroyDependents, channel.Id)
	return
}

// TODO
func (c *Manager) HealChannels() error {
	log.Info().Msg("haling channels")
	// fetch all workers owned by this manager that errored or stopped heartbeat
	// reboot workers
	return nil
}

// TODO
func (c *Manager) HealMessages() error {
	log.Info().Msg("haling messages")
	// fetch all messages processed by owned workers that failed
	// reprocess all messages if attempts < max
	return nil
}

func (c *Manager) PruneChannels() (err error) {
	log.Info().Msg("pruning channels")
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
	// channel and client must be in the same group
	return true
}

func (c *Manager) workersReport() <-chan ChannelProvider {
	workers := make(chan ChannelProvider)
	go func() {
		defer close(workers)
		for _, w := range c.object.channels {
			workers <- w
		}
	}()
	return workers
}

func (c *Manager) Info() managerSummary {
	var idleChannels uint = 0
	channelSummaries := make([]channelSummary, 0, len(c.object.channels))
	for _, channel := range c.object.channels {
		if channel.Idle() {
			idleChannels++
		}
		channelSummaries = append(channelSummaries, channel.Info())
	}
	return managerSummary{
		State:            c.object.state,
		AllChannels:      uint(len(c.object.channels)),
		IdleChannels:     idleChannels,
		ChannelSummaries: channelSummaries,
	}
}

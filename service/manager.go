package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arrno/chanman/broker"
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/service/io"
	"github.com/arrno/chanman/util"
	"github.com/rs/zerolog/log"
)

type ManagerProvider interface {
	Info() managerSummary
	Destroy()
	ExpectChannel(channel *domain.Channel) (ChannelProvider, error)
	DestroyChannel(worker *domain.Channel) (err error)
	AddSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) (bool, error)
	RemoveSubscription(subscriber *domain.TopicSubscription, channel *domain.Channel) error
	PublishMessage(clientId string, channel *domain.Channel, message *domain.Message) (err error)
}

type Manager struct {
	cleanupIntervalSec uint
	systemIo           io.SystemProvider
	broker             broker.BrokerProvider
	object             managerObject
	usageStart         time.Time
	destroy            chan any
	debugMode          bool
}

type managerObject struct {
	mu                sync.RWMutex
	state             domain.STATE
	channels          map[string]ChannelProvider // map[channelId]ChannelProvider
	destroyDependents map[string][]chan any      // map[channel.Id][]chan any
}

func NewManager(db datastore.RepoProvider, broker broker.BrokerProvider) *Manager {
	m := Manager{
		broker:             broker,
		systemIo:           io.NewSystem(db),
		cleanupIntervalSec: domain.CLEANUP_INTERVAL_SECONDS,
		debugMode:          util.Env().Expect("DEBUG") == "true",
		usageStart:         time.Now(),
		object: managerObject{
			state:             domain.STATE_RUNNING,
			channels:          map[string]ChannelProvider{},
			destroyDependents: map[string][]chan any{},
		},
	}
	m.destroy = m.SpawnWatcher()
	return &m
}

func (c *Manager) ExpectChannel(channel *domain.Channel) (ChannelProvider, error) {
	c.object.mu.Lock()
	defer c.object.mu.Unlock()

	if worker, ok := c.object.channels[channel.Id]; ok {
		return worker, nil
	}

	created, err := c.systemIo.CreateChannelIfNotFound(channel)

	if err != nil {
		return nil, err
	}

	if created {
		log.Info().Msg("Creating channel in broker: " + channel.Id)
		if err := c.broker.CreateChannel(channel); err != nil {
			return nil, err
		}
	}

	log.Info().Msg("Creating channel in memory: " + channel.Id)
	worker, done, err := c.bootChannel(channel)
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

	// no block
	go func() {
		subscriber.UpdateState(domain.STATE_INACTIVE)
		log.Info().
			Str("subscription", subscriber.Id).
			Str("channel", channel.Id).
			Str("client", subscriber.ClientId).
			Str("topics", strings.Join(subscriber.Topics, ",")).
			Msg("Setting topic subscriber inactive")
		if err := c.systemIo.UpdateTopicSubscriber(subscriber); err != nil {
			log.Error().
				Str("subscription", subscriber.Id).
				Str("channel", channel.Id).
				Str("client", subscriber.ClientId).
				Str("topics", strings.Join(subscriber.Topics, ",")).
				Msg("Failed to update topic subscriber")
		}
	}()

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

	added := worker.AddSubscription(subscriber)

	if added {
		// no block
		go func() {
			log.Info().
				Str("subscription", subscriber.Id).
				Str("channel", channel.Id).
				Str("client", subscriber.ClientId).
				Str("topics", strings.Join(subscriber.Topics, ",")).
				Msg("Storing topic subscriber")
			if err := c.systemIo.StoreTopicSubscriber(subscriber); err != nil {
				log.Error().
					Str("subscription", subscriber.Id).
					Str("channel", channel.Id).
					Str("client", subscriber.ClientId).
					Str("topics", strings.Join(subscriber.Topics, ",")).
					Msg("Failed to store topic subscriber")
			}
		}()
	}

	return added, nil
}

func (c *Manager) PublishMessage(clientId string, channel *domain.Channel, message *domain.Message) (err error) {
	log.Info().Msg(fmt.Sprintf("Publishing message %s from client %s to channel %s", message.Id, clientId, channel.Id))
	if !c.Permitted(channel.Id, clientId, domain.PUBLISH) {
		return domain.ErrChanOpNotPermitted
	}

	// TODO, this should just go into a separate log
	// defer func() {
	// 	if err != nil {
	// 		message.UpdateState(domain.STATE_ERROR)
	// 	}
	// 	if storeErr := c.messageIo.CreateMessage(message); storeErr != nil {
	// 		err = fmt.Errorf("%w. %w", err, storeErr)
	// 	}
	// }()

	worker, err := c.ExpectChannel(channel)
	if err != nil {
		return err
	}

	return worker.Publish(message)
}

func (c *Manager) handleHeartbeat(channel *domain.Channel) {
	channel.UpdateHeartbeat()
	err := c.systemIo.UpdateChannel(channel)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Channel heartbeat failed to update: %s. Err: %s ", channel.Repr(), err.Error()))
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
				if err := c.PruneMessages(); err != nil {
					log.Error().Msg("Failed to prune messages. Err: " + err.Error())
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-destroy:
				return
			case <-time.After(time.Duration(domain.USAGE_AGG_INTERVAL_SECONDS) * time.Second):
				if err := c.RecordUsage(); err != nil {
					log.Error().Msg("Failed to report usage. Err: " + err.Error())
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

func (c *Manager) bootChannel(channel *domain.Channel) (*Channel, chan any, error) {
	update := func(message *domain.Message, subs []*domain.MessageSubscriber) (err error) {
		if err := c.systemIo.StoreMessage(message); err != nil {
			return err
		}
		return c.systemIo.StoreMessageSubscribers(subs)
	}

	worker := NewChannel(channel, c.broker, update)
	if err := worker.Listen(); err != nil {
		return nil, nil, err
	}

	done := make(chan any)
	worker.Heartbeat(done, c.handleHeartbeat)

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
		if writeErr := c.systemIo.UpdateChannel(channel); writeErr != nil {
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
	log.Info().Msg("healing channels")
	// fetch all workers owned by this manager that errored or stopped heartbeat
	// reboot workers
	return nil
}

// TODO
func (c *Manager) HealMessages() error {
	log.Info().Msg("healing messages")
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

func (c *Manager) PruneMessages() error {
	log.Info().Msg("pruning messages")
	return c.systemIo.PruneMessages(domain.MESSAGE_RETENTION_HOURS)
}

func (c *Manager) RecordUsage() (err error) {
	log.Info().Msg(fmt.Sprintf("recording usage since last stop: %s", c.usageStart.Format(time.RFC3339)))
	stop := time.Now()
	usage, err := c.systemIo.AggregateClientMessageUsage(c.usageStart, stop)
	if err != nil {
		return err
	}
	err = c.systemIo.StoreClientUsage(usage)
	if err != nil {
		return err
	}
	c.usageStart = stop
	return nil
}

// This is sufficiently handled in the middleware but may be expanded in the future
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

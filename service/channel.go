package service

import (
	"context"
	"sync"
	"time"

	"github.com/arrno/chanman/broker"
	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

type ChannelProvider interface {
	Info() channelSummary
	Object() *domain.Channel
	Listen() error // open subscription... on message, broadcast
	Destroy() error
	Reboot() error
	Idle() bool
	Publish(*domain.Message) error
	Heartbeat(done chan any, f func(*domain.Channel))
	IncrPublishers() func()
	AddSubscription(subscriber *domain.TopicSubscription) bool
	RemoveSubscription(topics []string, id string)
}

// One worker per channel
type Channel struct {
	mu               sync.Mutex
	object           *domain.Channel
	broker           broker.BrokerProvider
	rootSubscription *domain.ChannelSubscription
	subscriptionBook AddressBookProvider
	publishers       uint
}

func NewChannel(
	chanObject *domain.Channel,
	broker broker.BrokerProvider,
	update func(message *domain.Message, subs []*domain.MessageSubscriber) error,
) *Channel {
	c := Channel{
		object: chanObject,
		broker: broker,
		rootSubscription: &domain.ChannelSubscription{
			Id:        chanObject.Id,
			ChannelId: chanObject.Id,
			State:     domain.STATE_PENDING,
			Archived:  false,
			Callback:  nil,
		},
		subscriptionBook: NewAddressBook(chanObject.Id),
	}

	// Broadcast
	c.rootSubscription.Callback = func(message *domain.Message) (err error) {

		var subscribers []*domain.TopicSubscription
		if message.Topic != domain.ALL_TOPIC {
			subscribers = c.subscriptionBook.FetchTopicSubscriptions([]string{domain.ALL_TOPIC})
		}
		for _, v := range c.subscriptionBook.FetchTopicSubscriptions([]string{message.Topic}) {
			subscribers = append(subscribers, v)
		}

		message.Acknowledge(uint(len(subscribers)))

		defer func() {
			messageSubscribers := make([]*domain.MessageSubscriber, len(subscribers))
			for i, sub := range subscribers {
				messageSubscribers[i] = domain.NewMessageSubscriber(
					message.Id, sub.ClientId,
				)
			}

			if err != nil {
				message.UpdateState(domain.STATE_ERROR)
			} else {
				message.UpdateState(domain.STATE_COMPLETE)
			}

			if err := update(message, messageSubscribers); err != nil {
				log.Error().
					Str("messageId", message.Id).
					Str("clientId", message.ClientId).
					Str("channelId", message.ChannelId).
					Msg("Failed to store message. Err: " + err.Error())
			}
		}()

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

	return &c
}

func (c *Channel) Object() *domain.Channel {
	return c.object
}

func (c *Channel) Publish(message *domain.Message) error {
	message.Process()
	return c.broker.PublishMessage(context.Background(), c.object, message)
}

func (c *Channel) IncrPublishers() func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishers++

	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.publishers--
	}
}

func (c *Channel) Listen() (err error) {
	// Set initial state to pending
	c.mu.Lock()
	c.object.State = domain.STATE_PENDING
	c.mu.Unlock()

	// Subscribe for messages
	err = c.broker.SubscribeForMessages(
		context.Background(),
		c.object,
		c.rootSubscription,
	)

	// Update state based on result
	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.object.State = domain.STATE_ERROR
	} else {
		c.object.State = domain.STATE_RUNNING
	}

	return
}

func (c *Channel) Destroy() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already inactive, nothing to do
	if c.object.State == domain.STATE_INACTIVE {
		return nil
	}

	// First mark as being pending
	c.object.State = domain.STATE_PENDING

	// Clear all subscriptions
	c.subscriptionBook.Clear()

	// Unsubscribe from broker
	err = c.broker.UnsubscribeForMessages(
		context.Background(),
		c.object,
		c.rootSubscription,
	)

	// Update final state based on result
	if err != nil {
		c.object.State = domain.STATE_ERROR
	} else {
		c.object.State = domain.STATE_INACTIVE
		// Reset publisher count
		c.publishers = 0
	}

	return
}

func (c *Channel) Reboot() (err error) {
	return nil
}

func (c *Channel) Idle() bool {
	return c.publishers == 0 && c.subscriptionBook.Size() == 0
}

func (c *Channel) AddSubscription(subscriber *domain.TopicSubscription) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.object.State != domain.STATE_RUNNING {
		return false
	}
	c.subscriptionBook.CacheSubscription(subscriber)
	return true
}

func (c *Channel) RemoveSubscription(topics []string, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.object.State != domain.STATE_RUNNING {
		return
	}

	c.subscriptionBook.UnCacheChannelSubscription(topics, id)
}

func (c *Channel) Heartbeat(done chan any, f func(*domain.Channel)) {
	started := make(chan any)

	go func() {
		close(started)
		for {
			select {
			case <-done:
				return
			case <-time.After(time.Duration(c.object.IntervalSeconds) * time.Second):
				func() {
					f(c.object)
				}()
			}
		}
	}()

	<-started
}

func (c *Channel) Info() channelSummary {
	return channelSummary{
		Name:             c.object.Name,
		State:            c.object.State,
		Publishers:       c.publishers,
		TopicSubscribers: c.subscriptionBook.Info(),
	}
}

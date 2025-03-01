package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/util"
)

type InMemoryBroker struct {
	channels    sync.Map // key: channel ID, value: *domain.Channel
	subscribers sync.Map // key: channel ID, value: []*domain.ChannelSubscription
	messages    sync.Map // key: message ID, value: *domain.Message
}

func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{}
}

func (b *InMemoryBroker) CreateChannel(channel *domain.Channel) error {
	if channel.Id == "" {
		channel.Id = util.Nanoid()
	}
	channel.CreatedAt = time.Now()
	b.channels.Store(channel.Id, channel)
	b.subscribers.Store(channel.Id, []*domain.ChannelSubscription{}) // Initialize empty subscriber list
	return nil
}

func (b *InMemoryBroker) DestroyChannel(channel *domain.Channel) error {
	b.channels.Delete(channel.Id)
	b.subscribers.Delete(channel.Id)
	return nil
}

func (b *InMemoryBroker) UpdateChannel(channel *domain.Channel) error {
	channel.UpdatedAt = time.Now()
	b.channels.Store(channel.Id, channel)
	return nil
}

func (b *InMemoryBroker) PublishMessage(ctx context.Context, channel *domain.Channel, m *domain.Message) error {
	if m.Id == "" {
		m.Id = util.Nanoid()
	}
	m.CreatedAt = time.Now()
	b.messages.Store(m.Id, m)

	// Fetch subscribers
	val, ok := b.subscribers.Load(channel.Id)
	if !ok {
		return fmt.Errorf("no subscribers for channel: %s", channel.Id)
	}
	subscribers := val.([]*domain.ChannelSubscription)

	// Drop message after retention time
	var wg sync.WaitGroup
	wg.Add(len(subscribers))
	go func() {
		wg.Wait()
		<-time.After(time.Duration(60) * time.Second)
		b.messages.Delete(m.Id)
	}()

	// Broadcast message to all subscribers
	for _, sub := range subscribers {
		go func(s *domain.ChannelSubscription) {
			defer wg.Done()
			if s.Callback != nil {
				err := s.Callback(m)
				if err != nil {
					fmt.Printf("Subscriber %s failed to process message: %v\n", s.Id, err)
				}
			}
		}(sub)
	}

	return nil
}

func (b *InMemoryBroker) SubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.ChannelSubscription) error {
	subscriber.CreatedAt = time.Now()

	val, _ := b.subscribers.LoadOrStore(channel.Id, []*domain.ChannelSubscription{})
	subscribers := val.([]*domain.ChannelSubscription)

	b.subscribers.Store(channel.Id, append(subscribers, subscriber))
	return nil
}

func (b *InMemoryBroker) UnsubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.ChannelSubscription) error {
	val, ok := b.subscribers.Load(channel.Id)
	if !ok {
		return fmt.Errorf("channel %s has no subscribers", channel.Id)
	}
	subscribers := val.([]*domain.ChannelSubscription)

	// Filter out the subscriber
	var updatedSubs []*domain.ChannelSubscription
	for _, sub := range subscribers {
		if sub.Id != subscriber.Id {
			updatedSubs = append(updatedSubs, sub)
		}
	}

	b.subscribers.Store(channel.Id, updatedSubs)
	return nil
}

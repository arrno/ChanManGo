package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

type InMemoryBroker struct {
	channels    sync.Map // key: channel ID, value: *domain.Channel
	subscribers sync.Map // key: channel ID, value: []*domain.ChannelSubscription
	messages    sync.Map // key: message ID, value: *domain.Message
	cleanup     chan string
	shutdown    chan any
}

func NewInMemoryBroker() *InMemoryBroker {
	b := InMemoryBroker{
		channels:    sync.Map{},
		subscribers: sync.Map{},
		messages:    sync.Map{},
		cleanup:     make(chan string),
		shutdown:    make(chan any),
	}
	// Drop message after retention time
	go func() {
		ticker := time.NewTicker(time.Duration(60) * time.Second)
		deleteIds := []string{}
		for {
			select {
			case <-b.shutdown:
				return
			case id := <-b.cleanup:
				deleteIds = append(deleteIds, id)
			case <-ticker.C:
				for _, id := range deleteIds {
					b.messages.Delete(id)
				}
				deleteIds = nil
			}
		}

	}()
	return &b
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
		select {
		case <-b.shutdown:
			return
		case b.cleanup <- m.Id:
		}
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

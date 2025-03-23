package service

import (
	"sync"

	"github.com/arrno/chanman/domain"
)

type AddressBookProvider interface {
	FetchTopicSubscriptions(topics []string) []*domain.TopicSubscription
	CacheSubscription(subscription *domain.TopicSubscription) bool
	UnCacheChannelSubscription(topics []string, id string)
	Clear()
	Size() uint
}

// Can offload to cache server. Represents one worker-channel
type AddressBook struct {
	mu          sync.RWMutex
	sizeMu      sync.RWMutex
	ChannelId   string
	size        uint
	subscribers map[string]map[string]*domain.TopicSubscription // map[topic]map[id]
}

func NewAddressBook(id string) *AddressBook {
	return &AddressBook{
		ChannelId: id,
		subscribers: map[string]map[string]*domain.TopicSubscription{
			"all": map[string]*domain.TopicSubscription{},
		},
	}
}

func (b *AddressBook) Clear() {
	b.subscribers = nil
}
func (b *AddressBook) Size() uint {
	return b.size
}
func (b *AddressBook) inc(n uint) {
	b.sizeMu.Lock()
	defer b.sizeMu.Unlock()
	b.size += n
}
func (b *AddressBook) dec(n uint) {
	b.sizeMu.Lock()
	defer b.sizeMu.Unlock()
	b.size -= n
}

func (b *AddressBook) CacheSubscription(subscription *domain.TopicSubscription) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	var n uint = 0
	defer b.inc(n)

	for _, topic := range subscription.Topics {
		if _, ok := b.subscribers[topic]; !ok {
			b.subscribers[topic] = map[string]*domain.TopicSubscription{}
		}
		if _, ok := b.subscribers[topic][subscription.Id]; ok {
			continue
		}
		b.subscribers[topic][subscription.Id] = subscription
		n++
	}

	return true
}

func (b *AddressBook) UnCacheChannelSubscription(topics []string, id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var n uint = 0
	defer b.dec(n)

	for _, topic := range topics {
		if _, ok := b.subscribers[topic]; !ok {
			continue
		}
		if _, ok := b.subscribers[topic][id]; ok {
			n++
			delete(b.subscribers[topic], id)
		}
	}
}

func (b *AddressBook) FetchTopicSubscriptions(topics []string) []*domain.TopicSubscription {
	var subs []*domain.TopicSubscription

	for _, topic := range topics {
		subMap, ok := b.subscribers[topic]
		if !ok {
			continue
		}
		for _, sub := range subMap {
			subs = append(subs, sub)
		}
	}

	return subs
}

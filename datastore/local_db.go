package datastore

import (
	"fmt"
	"sync"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

func docId(docType string, docId string) string {
	return fmt.Sprintf("%s:%s", docType, docId)
}

type LocalDB struct {
	data sync.Map
}

func NewInMemoryDB() *LocalDB {
	return &LocalDB{
		data: sync.Map{},
	}
}

func (db *LocalDB) StoreChannel(ch *domain.Channel) error {
	if ch.Id == "" {
		ch.Id = util.Nanoid()
	}
	db.data.Store(docId("channel", ch.Id), ch)
	return nil
}
func (db *LocalDB) GetChannel(id string) (*domain.Channel, error) {
	val, ok := db.data.Load(docId("channel", id))
	if ok {
		return val.(*domain.Channel), nil
	}
	return nil, ErrNotFound
}
func (db *LocalDB) UpdateChannel(ch *domain.Channel) error {
	return db.StoreChannel(ch)
}
func (db *LocalDB) ArchiveChannel(id string) error {
	db.data.Delete(docId("channel", id))
	return nil
}

func (db *LocalDB) StoreWorker(w *domain.Channel) error {
	if w.Id == "" {
		w.Id = util.Nanoid()
	}
	db.data.Store(docId("worker", w.Id), w)
	return nil
}

func (db *LocalDB) GetWorker(id string) (*domain.Channel, error) {
	val, ok := db.data.Load(docId("worker", id))
	if ok {
		return val.(*domain.Channel), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateWorker(w *domain.Channel) error {
	return db.StoreWorker(w)
}

func (db *LocalDB) ArchiveWorker(id string) error {
	db.data.Delete(docId("worker", id))
	return nil
}

func (db *LocalDB) StoreTopicSubscriber(topicSub *domain.TopicSubscription) error {
	if topicSub.Id == "" {
		topicSub.Id = util.Nanoid()
	}
	db.data.Store(docId("topic_subscriber", topicSub.Id), topicSub)
	return nil
}

func (db *LocalDB) GetTopicSubscriber(id string) (*domain.TopicSubscription, error) {
	val, ok := db.data.Load(docId("topic_subscriber", id))
	if ok {
		return val.(*domain.TopicSubscription), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateTopicSubscriber(topicSub *domain.TopicSubscription) error {
	return db.StoreTopicSubscriber(topicSub)
}

func (db *LocalDB) ArchiveTopicSubscriber(id string) error {
	db.data.Delete(docId("topic_subscriber", id))
	return nil
}

func (db *LocalDB) StoreChannelSubscriber(sub *domain.ChannelSubscription) error {
	if sub.Id == "" {
		sub.Id = util.Nanoid()
	}
	db.data.Store(docId("channel_subscriber", sub.Id), sub)
	return nil
}

func (db *LocalDB) GetChannelSubscriber(id string) (*domain.ChannelSubscription, error) {
	val, ok := db.data.Load(docId("channel_subscriber", id))
	if ok {
		return val.(*domain.ChannelSubscription), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateChannelSubscriber(sub *domain.ChannelSubscription) error {
	return db.StoreChannelSubscriber(sub)
}

func (db *LocalDB) ArchiveChannelSubscriber(id string) error {
	db.data.Delete(docId("channel_subscriber", id))
	return nil
}

func (db *LocalDB) StoreMessage(msg *domain.Message) error {
	if msg.Id == "" {
		msg.Id = util.Nanoid()
	}
	db.data.Store(docId("message", msg.Id), msg)
	return nil
}

func (db *LocalDB) GetMessage(id string) (*domain.Message, error) {
	val, ok := db.data.Load(docId("message", id))
	if ok {
		return val.(*domain.Message), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateMessage(msg *domain.Message) error {
	return db.StoreMessage(msg)
}

func (db *LocalDB) ArchiveMessage(id string) error {
	db.data.Delete(docId("message", id))
	return nil
}

// Transaction wrapper, executing function within a transaction
func (db *LocalDB) WithTx(f func(tx DatastoreProvider) error) error {
	return f(db) // In-memory DB does not support real transactions, so we just execute the function.
}

package datastore

import (
	"fmt"
	"sort"
	"strings"
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

// CHANNEL
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

// TOPIC SUBSCRIPTION
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

// CHAN SUBSCRIPTION
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

// MESSAGE
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

// CLIENT
func (db *LocalDB) StoreClient(client *domain.Client) error {
	if client.Id == "" {
		client.Id = util.Nanoid()
	}
	db.data.Store(docId("client", client.Id), client)
	return nil
}

func (db *LocalDB) GetClient(id string) (*domain.Client, error) {
	val, ok := db.data.Load(docId("client", id))
	if ok {
		return val.(*domain.Client), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) GetClients(pagination domain.Pagination[*domain.Client]) ([]*domain.Client, error) {
	clients := []*domain.Client{}

	db.data.Range(func(key any, value any) bool {
		if strings.HasPrefix(key.(string), "client:") {
			clients = append(clients, value.(*domain.Client))
		}
		return true
	})
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Id < clients[j].Id
	})

	return pagination.DoPaginate(clients), nil
}

func (db *LocalDB) UpdateClient(client *domain.Client) error {
	return db.StoreClient(client)
}

func (db *LocalDB) ArchiveClient(id string) error {
	db.data.Delete(docId("client", id))
	return nil
}

// PASSWORD
func (db *LocalDB) StoreClientSecret(pw *domain.ClientSecret) error {
	if pw.Id == "" {
		pw.Id = util.Nanoid()
	}
	db.data.Store(docId("client_secret_token", pw.Id), pw)
	db.data.Store(docId("client_secret_token_by_client_id", pw.ClientId), pw)
	db.data.Store(docId("client_secret_token_by_token", pw.Hash), pw)
	return nil
}

func (db *LocalDB) GetClientSecret(id string) (*domain.ClientSecret, error) {
	val, ok := db.data.Load(docId("client_secret_token", id))
	if ok {
		return val.(*domain.ClientSecret), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) GetClientSecretByClientId(clientId string) (*domain.ClientSecret, error) {
	val, ok := db.data.Load(docId("client_secret_token_by_client_id", clientId))
	if ok {
		return val.(*domain.ClientSecret), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) GetClientGroupByToken(token string) (*domain.ClientGroup, error) {
	val, ok := db.data.Load(docId("client_secret_token_by_token", token))
	if !ok {
		return nil, ErrNotFound
	}
	clientGroup, err := db.GetClientGroupByClientId(val.(*domain.ClientSecret).ClientId)
	if err != nil {
		return nil, ErrNotFound
	}
	return clientGroup, nil
}

func (db *LocalDB) UpdateClientSecret(pw *domain.ClientSecret) error {
	return db.StoreClientSecret(pw)
}

func (db *LocalDB) ArchiveClientSecret(id string) error {
	pw, err := db.GetClientSecret(id)
	if err != nil {
		return err
	}
	db.data.Delete(docId("client_secret_token", id))
	db.data.Delete(docId("client_secret_token_by_client_id", pw.ClientId))
	db.data.Delete(docId("client_secret_token_by_token", pw.Hash))
	return nil
}

// GROUP
func (db *LocalDB) StoreGroup(group *domain.Group) error {
	if group.Id == "" {
		group.Id = util.Nanoid()
	}
	db.data.Store(docId("group", group.Id), group)
	return nil
}

func (db *LocalDB) GetGroup(id string) (*domain.Group, error) {
	val, ok := db.data.Load(docId("group", id))
	if ok {
		return val.(*domain.Group), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateGroup(group *domain.Group) error {
	return db.StoreGroup(group)
}

func (db *LocalDB) ArchiveGroup(id string) error {
	db.data.Delete(docId("group", id))
	return nil
}

// CLIENT GROUP
func (db *LocalDB) StoreClientGroup(cg *domain.ClientGroup) error {
	if cg.Id == "" {
		cg.Id = util.Nanoid()
	}
	db.data.Store(docId("client_group", cg.Id), cg)
	db.data.Store(docId("client_group_by_client_id", cg.ClientId), cg)
	return nil
}

func (db *LocalDB) GetClientGroup(id string) (*domain.ClientGroup, error) {
	val, ok := db.data.Load(docId("client_group", id))
	if ok {
		return val.(*domain.ClientGroup), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) GetClientGroupByClientId(clientId string) (*domain.ClientGroup, error) {
	val, ok := db.data.Load(docId("client_group_by_client_id", clientId))
	if ok {
		return val.(*domain.ClientGroup), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdateClientGroup(cg *domain.ClientGroup) error {
	return db.StoreClientGroup(cg)
}

func (db *LocalDB) ArchiveClientGroup(id string) error {
	cg, err := db.GetClientGroup(id)
	if err != nil {
		return err
	}
	db.data.Delete(docId("client_group", id))
	db.data.Delete(docId("client_group_by_client_id", cg.ClientId))
	return nil
}

// PORTAL
func (db *LocalDB) StorePortal(portal *domain.Portal) error {
	if portal.Id == "" {
		portal.Id = util.Nanoid()
	}
	db.data.Store(docId("portal", portal.Id), portal)
	return nil
}

func (db *LocalDB) GetPortal(id string) (*domain.Portal, error) {
	val, ok := db.data.Load(docId("portal", id))
	if ok {
		return val.(*domain.Portal), nil
	}
	return nil, ErrNotFound
}

func (db *LocalDB) UpdatePortal(portal *domain.Portal) error {
	return db.StorePortal(portal)
}

func (db *LocalDB) ArchivePortal(id string) error {
	db.data.Delete(docId("portal", id))
	return nil
}

// Transaction wrapper, executing function within a transaction
func (db *LocalDB) WithTx(f func(tx DatastoreProvider) error) error {
	return f(db) // In-memory DB does not support real transactions, so we just execute the function.
}

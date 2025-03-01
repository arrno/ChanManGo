package domain

import "github.com/arrno/chanban/util"

// ChannelSubscription is a connection from a worker to a broker channel
type ChannelSubscription struct {
	Id       string                       `json:"id"`
	PortalId string                       `json:"portalId"`
	State    STATE                        `json:"state"`
	Archived bool                         `json:"archived"`
	Callback func(message *Message) error `json:"-"`
}

// TopicSubscription is a connection from a client to a channel worker over a subset of topics
type TopicSubscription struct {
	Id       string                       `json:"id"`
	ClientId string                       `json:"clientId"`
	Topics   []string                     `json:"topicName"`
	State    STATE                        `json:"state"`
	Archived bool                         `json:"archived"`
	Callback func(message *Message) error `json:"-"`
}

func NewTopicSubscription(clientId string, topics []string) *TopicSubscription {
	return &TopicSubscription{
		Id:       util.Nanoid(),
		ClientId: clientId,
		Topics:   topics,
		State:    STATE_ACTIVE,
	}
}

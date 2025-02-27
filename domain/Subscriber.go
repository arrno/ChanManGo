package domain

import "github.com/arrno/chanban/util"

type Subscription struct {
	Id        string                       `json:"id"`
	ClientId  string                       `json:"clientId"`
	TopicName string                       `json:"topicName"`
	State     STATE                        `json:"state"`
	Archived  bool                         `json:"archived"`
	Callback  func(message *Message) error `json:"-"`
}

func NewSubscription(clientId, topicName string) *Subscription {
	return &Subscription{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		TopicName: topicName,
		State:     ACTIVE,
	}
}

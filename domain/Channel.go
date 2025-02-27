package domain

import (
	"fmt"
)

type Channel struct {
	Id        string `json:"id"`
	PortalId  string `json:"portalId"`
	TopicName string `json:"topicName"`
	State     STATE  `json:"state"`
	Archived  bool   `json:"archived"`
}

func NewChannel(portalId, topic string) *Channel {
	return &Channel{
		Id:        fmt.Sprintf("%s__%s", portalId, topic),
		PortalId:  portalId,
		TopicName: topic,
		State:     ACTIVE,
	}
}

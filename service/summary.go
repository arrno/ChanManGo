package service

import (
	"encoding/json"
	"fmt"

	"github.com/arrno/chanman/domain"
	"github.com/rs/zerolog/log"
)

type bookSummary struct {
	Size          uint            `json:"size"`
	TopicSubCount map[string]uint `json:"topicSubscriptionCount"`
}

type channelSummary struct {
	Name             string       `json:"name"`
	State            domain.STATE `json:"state"`
	Publishers       uint         `json:"publisherCount"`
	TopicSubscribers bookSummary  `json:"subscribers"`
}

type managerSummary struct {
	State            domain.STATE     `json:"state"`
	AllChannels      uint             `json:"channelCount"`
	IdleChannels     uint             `json:"idleChannelCount"`
	ChannelSummaries []channelSummary `json:"channels"`
}

func (s managerSummary) Print() {
	b, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		log.Error().Msg("Failed to marshal for summary info. Err: " + err.Error())
		return
	}
	fmt.Println(string(b))
}

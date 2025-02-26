package domain

import "time"

type Message struct {
	Id             string         `json:"id"`
	ClientId       string         `json:"clientId"`
	AcknowledgedAt *time.Time     `json:"acknowledged"`
	State          STATE          `json:"state"`
	Archived       bool           `json:"archived"`
	Payload        map[string]any `json:"payload"`
}

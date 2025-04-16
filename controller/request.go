package controller

import (
	"time"

	"github.com/arrno/chanman/domain"
)

// Request structures

// ClientRequest is used for client-related operations
type ClientRequest struct {
	ClientId   string `json:"clientId"`
	GroupId    string `json:"groupId,omitempty"`
	Admin      bool   `json:"admin"`
	Email      string `json:"email"`
	ClientName string `json:"clientName"`
}

// ChannelRequest is used for channel operations
type ChannelRequest struct {
	Name            string `json:"name"`
	GroupId         string `json:"groupId"`
	PortalId        string `json:"portalId,omitempty"`
	IntervalSeconds uint   `json:"intervalSeconds"`
}

// GroupRequest is used for group operations
type GroupRequest struct {
	Name     string `json:"name"`
	ParentId string `json:"parentId,omitempty"`
	PortalId string `json:"portalId,omitempty"`
}

// MessageRequest is used for publishing messages
type MessageRequest struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

type UpdateClientRequest struct {
	Email      string      `json:"email,omitempty"`
	Tier       domain.Tier `json:"tier,omitempty"`
	ClientName string      `json:"clientName,omitempty"`
}

// Response structures

// StandardResponse is a consistent response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ClientTokenResponse is returned when creating/authenticating clients
type ClientTokenResponse struct {
	ClientId      string      `json:"clientId"`
	Email         string      `json:"email,omitempty"`
	Tier          domain.Tier `json:"tier,omitempty"`
	ClientChannel string      `json:"channel,omitempty"`
	Token         string      `json:"token,omitempty"`
	ClientName    string      `json:"clientName,omitempty"`
	ClientGroup   string      `json:"clientGroup,omitempty"`
	PubCount      uint        `json:"pubCount,omitempty"`
	SubCount      uint        `json:"subCount,omitempty"`
	ExpiresAt     *time.Time  `json:"tokenExpiresAt,omitempty"`
}

// ChannelResponse includes channel information
type ChannelResponse struct {
	Id              string     `json:"id"`
	Name            string     `json:"name"`
	GroupId         string     `json:"groupId"`
	State           string     `json:"state"`
	IntervalSeconds uint       `json:"intervalSeconds"`
	HeartbeatAt     *time.Time `json:"heartbeatAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

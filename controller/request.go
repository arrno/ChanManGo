package controller

import "time"

// Request structures

// ClientRequest is used for client-related operations
type ClientRequest struct {
	ClientId   string `json:"clientId"`
	GroupId    string `json:"groupId,omitempty"`
	Admin      bool   `json:"admin"`
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
	ClientChannel string     `json:"channel"`
	Token         string     `json:"token"`
	ClientId      string     `json:"clientId,omitempty"`
	ClientName    string     `json:"clientName,omitempty"`
	ClientGroup   string     `json:"clientGroup,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
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

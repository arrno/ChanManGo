package controller

// Request / Response shapes

type ClientRequest struct {
	ClientId   string `json:"clientId"`
	GroupId    string `json:"groupId omitempty"`
	Admin      bool   `json:"admin"`
	ClientName string `json:"clientName"`
}
type ClientTokenResponse struct {
	ClientId   string `json:"clientId"`
	ClientName string `json:"clientName omitempty"`
	Token      string `json:"token"`
}

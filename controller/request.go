package controller

// Request / Response shapes

type ClientRequest struct {
	ClientId   string `json:"clientId"`
	Admin      bool   `json:"admin"`
	ClientName string `json:"clientName"`
}
type ClientTokenResponse struct {
	ClientId   string `json:"clientId"`
	ClientName string `json:"clientName omitempty"`
	Token      string `json:"token"`
}

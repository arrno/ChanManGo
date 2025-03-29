package controller

// Request / Response shapes

type ClientRequest struct {
	ClientId   string `json:"clientId"`
	GroupId    string `json:"groupId omitempty"`
	Admin      bool   `json:"admin"`
	ClientName string `json:"clientName"`
}
type ClientTokenResponse struct {
	ClientURN   string `json:"clientUrn"`
	ClientId    string `json:"clientId"`
	ClientGroup string `json:"clientGroup"`
	Token       string `json:"token"`
}

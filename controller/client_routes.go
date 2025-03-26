package controller

import (
	"encoding/json"
	"net/http"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/service"
)

// Client
func (c *Controller) GetClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to parse client request"})
		return
	}
	client, err := service.NewClient(c.db).GetClient(req.ClientId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to get client. Err " + err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": client})
	return
}

//	func (c *Controller) GetClients(w http.ResponseWriter, r *http.Request) {
//		service.NewClient(c.db).CreateClient()
//	}

func (c *Controller) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to parse client request"})
		return
	}
	client := domain.NewClient(req.ClientName, req.Admin)
	token, err := service.NewClient(c.db).CreateClient(client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to create client. Err " + err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": ClientTokenResponse{
		ClientId:   client.Id,
		ClientName: client.Name,
		Token:      token.Raw,
	}})
	return
}

func (c *Controller) RotateToken(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to parse client request"})
		return
	}
	token, err := service.NewClient(c.db).RotateClientSecret(req.ClientId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to rotate client token. Err " + err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": ClientTokenResponse{
		ClientId: req.ClientId,
		Token:    token.Raw,
	}})
	return
}

func (c *Controller) ArchiveClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to parse client request"})
		return
	}
	if err := service.NewClient(c.db).ArchiveClient(req.ClientId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Unable to archive client. Err " + err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok"})
	return
}

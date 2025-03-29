package controller

import (
	"encoding/json"
	"net/http"

	"github.com/arrno/chanman/domain"
)

// Client
func (c *Controller) GetClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		http.Error(w, "Unable to parse client request", http.StatusBadRequest)
		return
	}
	client, err := c.system.GetClient(req.ClientId)
	if err != nil {
		http.Error(w, "Unable to get client. Err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": client})
}

func (c *Controller) GetClients(w http.ResponseWriter, r *http.Request) {
	var req domain.Pagination[*domain.Client]
	if err := UnmarshalTo(&req, r); err != nil {
		http.Error(w, "Unable to parse clients request", http.StatusBadRequest)
		return
	}
	client, err := c.system.GetClients(req)
	if err != nil {
		http.Error(w, "Unable to get clients. Err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Results": client})
}

func (c *Controller) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		http.Error(w, "Unable to parse client request", http.StatusBadRequest)
		return
	}

	groupId := req.GroupId
	if groupId == "" {
		groupId = domain.MAIN_GROUP
	}

	client, err := domain.NewClient(req.ClientName, groupId, domain.MAIN_PORTAL, req.Admin)
	if err != nil {
		http.Error(w, "Failed to handle client: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := c.system.CreateClient(client, domain.NewReadWrite(client.Id))
	if err != nil {
		http.Error(w, "Unable to create client. Err: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": ClientTokenResponse{
		ClientURN:   client.Id,
		ClientGroup: groupId,
		ClientId:    client.Name,
		Token:       token.Raw,
	}})
}

func (c *Controller) RotateToken(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		http.Error(w, "Unable to parse client request", http.StatusBadRequest)
		return
	}
	token, err := c.system.RotateClientSecret(req.ClientId)
	if err != nil {
		http.Error(w, "Unable to rotate client token. Err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": ClientTokenResponse{
		ClientId: req.ClientId,
		Token:    token.Raw,
	}})
}

func (c *Controller) ArchiveClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		http.Error(w, "Unable to parse client request", http.StatusBadRequest)
		return
	}
	if err := c.system.ArchiveClient(req.ClientId); err != nil {
		http.Error(w, "Unable to archive client. Err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok"})
}

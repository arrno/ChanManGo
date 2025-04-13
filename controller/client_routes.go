package controller

import (
	"errors"
	"net/http"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

func (c *Controller) ClientHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetClient(w, r)
	case http.MethodPost:
		c.CreateClient(w, r)
	case http.MethodDelete:
		c.ArchiveClient(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Client
func (c *Controller) GetClient(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("clientId")

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	client, err := c.system.GetClient(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Client not found", err)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, client)
}

func (c *Controller) GetClients(w http.ResponseWriter, r *http.Request) {
	var pg domain.Pagination[*domain.Client]
	err := pg.FromReq(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid pagination request format", err)
		return
	}
	filter := domain.NewFilter().FromReq(r)

	// Validate pagination parameters
	if err := pg.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	clients, err := c.system.GetClients(&pg, filter)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve clients", err)
		return
	}
	pg.Items = clients

	RespondWithJSON(w, http.StatusOK, pg)
}

func (c *Controller) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := UnmarshalTo(&req, r); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid client request format", err)
		return
	}

	// Validate required fields
	if req.ClientName == "" {
		RespondWithError(w, http.StatusBadRequest, "Client name is required", nil)
		return
	}

	groupId := req.GroupId
	if groupId == "" {
		groupId = domain.MAIN_GROUP
	}

	// TODO, validate group belongs to caller

	// Create client
	client, err := domain.NewClient("", req.ClientName, groupId, req.Email, req.Admin)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to create client", err)
		return
	}

	// Create client with permissions
	permission := domain.NewReadWrite(client.Id)
	token, err := c.system.CreateClient(client, permission)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to store client", err)
		return
	}

	// Create response
	response := ClientTokenResponse{
		ClientGroup: groupId,
		ClientId:    client.Id,
		Email:       client.Email,
		Tier:        permission.Tier,
		ClientName:  client.Name,
		Token:       token.Serialize(),
	}

	RespondWithJSON(w, http.StatusCreated, response)
}

func (c *Controller) RotateToken(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// Rotate token
	token, err := c.system.RotateClientSecret(clientId)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to rotate client token", err)
		return
	}

	// Create response
	response := ClientTokenResponse{
		ClientId: clientId,
		Token:    token.Serialize(),
	}

	RespondWithJSON(w, http.StatusOK, response)
}

func (c *Controller) CheckTokenBeta(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// via c.system SystemProvider
	// Fetch client/clientGroup/clientChannel/clientPermission

	// Create response
	response := ClientTokenResponse{
		ClientId:      clientId,
		Email:         "TODO",
		Tier:          "TODO",
		ClientChannel: "TODO",
		ClientName:    "TODO",
		ClientGroup:   "TODO",
	}

	RespondWithJSON(w, http.StatusOK, response)
}

func (c *Controller) UpdateTokenBeta(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")

	// TODO parse body into UpdateClientRequest
	// if requester is trying to update Tier and is not admin (request header per middleware), respond with Forbidden

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// via c.system SystemProvider...
	// Update permission tier if applicable
	// Update client email/name if applicable

	// Create response
	response := ClientTokenResponse{
		ClientId:      clientId,
		Email:         "TODO",
		Tier:          "TODO",
		ClientChannel: "TODO",
		ClientName:    "TODO",
		ClientGroup:   "TODO",
	}

	RespondWithJSON(w, http.StatusOK, response)
}

func (c *Controller) ArchiveClient(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("clientId")

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// Archive client
	if err := c.system.ArchiveClient(clientId); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to archive client", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]bool{"archived": true})
}

package controller

import (
	"errors"
	"net/http"
	"time"

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

// Contact handles form submissions from the public contact page.
// This is a public endpoint that does not require authentication.
//
// Endpoint: POST /contact
//
// Request body:
//
//	{
//	  "name": "User Name",
//	  "email": "user@example.com",
//	  "message": "Contact message content"
//	}
//
// Response: 200 OK on success
//
//	{
//	  "success": true,
//	  "message": "Contact request submitted successfully"
//	}
func (c *Controller) Contact(w http.ResponseWriter, r *http.Request) {
	// Parse the contact request from the request body
	var contactReq struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}

	if err := UnmarshalTo(&contactReq, r); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid contact request format", err)
		return
	}

	// Validate required fields
	if contactReq.Name == "" {
		RespondWithError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}

	if contactReq.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	if contactReq.Message == "" {
		RespondWithError(w, http.StatusBadRequest, "Message is required", nil)
		return
	}

	// Create a contact request domain object
	contactRequest := domain.NewContactRequest(
		contactReq.Name,
		contactReq.Email,
		contactReq.Message,
	)

	// Store the contact request
	if err := c.system.StoreContactRequest(contactRequest); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to submit contact request", err)
		return
	}

	// Return success response
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Contact request submitted successfully",
	})
}

// CheckTokenBeta verifies the client's token and returns their information.
// This endpoint requires authentication through the auth middleware.
//
// Endpoint: GET /beta/check-token
//
// Headers required:
// - clientId: The client ID (provided by auth middleware)
//
// Response: 200 OK on success
//
//	{
//	  "clientId": "client-id",
//	  "email": "user@example.com",
//	  "tier": "free|pro",
//	  "channel": "client-channel",
//	  "clientName": "Client Name",
//	  "clientGroup": "group-id"
//	}
func (c *Controller) CheckTokenBeta(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// Get client information
	client, err := c.system.GetClient(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Client not found", err)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client", err)
		return
	}

	// Get client permission
	permission, err := c.system.GetClientPermissionByClientId(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Client permission not found", err)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client permission", err)
		return
	}

	// Get client groups
	clientGroups, err := c.system.GetClientGroupsByClientId(clientId)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client groups", err)
		return
	}

	// Default group ID or get the first one
	var groupId string
	if len(clientGroups) > 0 {
		groupId = clientGroups[0].GroupId
	} else {
		groupId = domain.MAIN_GROUP
	}

	// Create response
	response := ClientTokenResponse{
		ClientId:      clientId,
		Email:         client.Email,
		Tier:          permission.Tier,
		ClientName:    client.Name,
		ClientGroup:   groupId,
		ClientChannel: "", // Client channel is optional
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// UpdateTokenBeta updates client information based on the request.
// This endpoint requires authentication through the auth middleware.
// Only admin users can update the client tier.
//
// Endpoint: POST /beta/update-token
//
// Headers required:
// - clientId: The client ID (provided by auth middleware)
// - admin: "true" if the requester is an admin (provided by auth middleware)
//
// Request body:
//
//	{
//	  "email": "user@example.com", // Optional
//	  "tier": "free|pro",          // Optional, requires admin privileges
//	  "clientName": "Client Name"  // Optional
//	}
//
// Response: 200 OK on success
//
//	{
//	  "clientId": "client-id",
//	  "email": "updated@example.com",
//	  "tier": "free|pro",
//	  "channel": "client-channel",
//	  "clientName": "Updated Name",
//	  "clientGroup": "group-id"
//	}
func (c *Controller) UpdateTokenBeta(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	isAdmin := r.Header.Get("admin") == "true"

	// Validate required fields
	if clientId == "" {
		RespondWithError(w, http.StatusBadRequest, "Client ID is required", nil)
		return
	}

	// Parse the update request
	var updateReq UpdateClientRequest
	if err := UnmarshalTo(&updateReq, r); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Check if the request includes tier update
	tierUpdateRequested := updateReq.Tier != ""

	// If trying to update the tier without admin privileges, return forbidden
	if tierUpdateRequested && !isAdmin {
		RespondWithError(w, http.StatusForbidden, "Only administrators can update tier", nil)
		return
	}

	// Get the current client information
	client, err := c.system.GetClient(clientId)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Client not found", err)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client", err)
		return
	}

	// Update client information if provided
	updated := false
	if updateReq.Email != "" && updateReq.Email != client.Email {
		client.Email = updateReq.Email
		updated = true
	}

	if updateReq.ClientName != "" && updateReq.ClientName != client.Name {
		client.Name = updateReq.ClientName
		updated = true
	}

	// Save client changes if needed
	if updated {
		client.UpdatedAt = time.Now()
		if err := c.system.UpdateClient(client); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update client information", err)
			return
		}
	}

	var permission *domain.ClientPermission
	// Update tier if requested and authorized
	if tierUpdateRequested {
		permission, err = c.system.GetClientPermissionByClientId(clientId)
		if err != nil {
			if errors.Is(err, datastore.ErrNotFound) {
				RespondWithError(w, http.StatusNotFound, "Client permission not found", err)
				return
			}
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client permission", err)
			return
		}

		// Update the tier if different
		if permission.Tier != updateReq.Tier {
			permission.Tier = updateReq.Tier
			permission.UpdatedAt = time.Now()
			if err := c.system.UpdateClientPermission(permission); err != nil {
				RespondWithError(w, http.StatusInternalServerError, "Failed to update client permissions", err)
				return
			}
		}
	}

	if permission == nil {
		// Get updated client permission
		permission, err = c.system.GetClientPermissionByClientId(clientId)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client permission", err)
			return
		}
	}

	// Get client groups
	clientGroups, err := c.system.GetClientGroupsByClientId(clientId)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve client groups", err)
		return
	}

	// Default group ID or get the first one
	var groupId string
	if len(clientGroups) > 0 {
		groupId = clientGroups[0].GroupId
	} else {
		groupId = domain.MAIN_GROUP
	}

	// Create response with updated information
	response := ClientTokenResponse{
		ClientId:      clientId,
		Email:         client.Email,
		Tier:          permission.Tier,
		ClientName:    client.Name,
		ClientGroup:   groupId,
		ClientChannel: "", // Client channel is optional
	}

	RespondWithJSON(w, http.StatusOK, response)
}

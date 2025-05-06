package controller

import (
	"errors"
	"net/http"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
	"github.com/rs/zerolog/log"
)

// SignUpBeta creates a new demo environment with client, group, and channel
func (c *Controller) SignUpBeta(w http.ResponseWriter, r *http.Request) {

	// Generate unique IDs for demo resources
	groupName := "MyGroup"
	channelName := "MyChan"
	clientName := "client-" + util.Nanoid()

	email := r.Header.Get("email")

	if len(email) == 0 {
		RespondWithError(w, http.StatusBadRequest, "Email value is required in URL", errors.New("Email value is required in URL"))
		return
	}

	// Create domain objects
	group, err := domain.NewGroup("", domain.GLOBAL_ADMIN, domain.MAIN_PORTAL, groupName)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create group", err)
		return
	}

	client, err := domain.NewClient("", clientName, group.Id, email, false)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create client", err)
		return
	}

	channel, err := domain.NewChannel("", domain.MAIN_PORTAL, group.Id, channelName, 60)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create channel", err)
		return
	}

	// permission := domain.NewReadOnly(client.Id)
	permission := domain.NewReadWrite(client.Id)

	// Use the database's transaction support
	token, err := c.system.InitSandbox(client, channel, group, permission)

	if err != nil {
		log.Error().Err(err).Msg("Failed to complete signup transaction")
		RespondWithError(w, http.StatusInternalServerError, "Failed to complete signup", err)
		return
	}

	// Create successful response
	response := ClientTokenResponse{
		ClientId:      client.Id,
		ClientChannel: channelName,
		Token:         token.Serialize(),
		ExpiresAt:     token.ExpiresAt,
	}

	// Return success response
	RespondWithJSON(w, http.StatusCreated, response)
}

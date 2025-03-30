package controller

import (
	"net/http"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
	"github.com/rs/zerolog/log"
)

// SignUpDemo creates a new demo environment with client, group, and channel
func (c *Controller) SignUpDemo(w http.ResponseWriter, r *http.Request) {

	// Generate unique IDs for demo resources
	groupName := util.Nanoid()
	channelName := "MyChan"
	clientName := "client-" + util.Nanoid()

	// Create domain objects
	group, err := domain.NewGroup(domain.ADMIN_USER, domain.MAIN_PORTAL, groupName)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create group", err)
		return
	}

	client, err := domain.NewClient(clientName, group.Id, false)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create client", err)
		return
	}

	channel, err := domain.NewChannel(domain.MAIN_PORTAL, group.Id, channelName, 3600)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create channel", err)
		return
	}

	permission := domain.NewReadOnly(client.Id)

	// Use the database's transaction support
	token, err := c.system.InitSandbox(client, channel, group, permission)

	if err != nil {
		log.Error().Err(err).Msg("Failed to complete signup transaction")
		RespondWithError(w, http.StatusInternalServerError, "Failed to complete signup", err)
		return
	}

	// Create successful response
	response := ClientTokenResponse{
		ClientChannel: channelName,
		Token:         token.Raw,
		ExpiresAt:     token.ExpiresAt,
	}

	// Return success response
	RespondWithJSON(w, http.StatusCreated, response)
}

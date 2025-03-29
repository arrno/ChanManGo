package controller

import (
	"encoding/json"
	"net/http"

	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

func (c *Controller) SignUpDemo(w http.ResponseWriter, r *http.Request) {

	groupId := util.Nanoid()
	clientId := util.Nanoid()
	chanelId := util.Nanoid()

	group, err := domain.NewGroup(domain.ADMIN_USER, domain.MAIN_PORTAL, groupId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to make group. Err " + err.Error()})
		return
	}
	client, err := domain.NewClient(clientId, groupId, domain.MAIN_PORTAL, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to make client. Err " + err.Error()})
		return
	}
	channel, err := domain.NewChannel(domain.MAIN_PORTAL, domain.MAIN_GROUP, chanelId, 3600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to make channel. Err " + err.Error()})
		return
	}
	permission := domain.NewReadOnly(clientId)

	// Ideally this runs in a transaction
	err = c.system.CreateGroup(group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to create group. Err " + err.Error()})
		return
	}
	_, err = c.system.CreateChannel(channel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to create channel. Err " + err.Error()})
		return
	}
	token, err := c.system.CreateClient(client, permission)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to create client. Err " + err.Error()})
		return
	}
	err = c.system.CreateClientGroup(clientId, groupId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to create client group. Err " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"Response": "Ok", "Result": ClientTokenResponse{
		ClientURN:   client.Id,
		ClientGroup: groupId,
		ClientId:    client.Name,
		Token:       token.Raw,
	}})
	return
}

package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/arrno/chanman/domain"
)

// Auth
func (c *Controller) Auth(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			token = r.URL.Query().Get("Authorization")
		}

		tokens := strings.Split(token, " ")
		if len(tokens) == 2 {
			token = tokens[1]
		}
		clientGroup, isAdmin, err := c.db.GetClientGroupByToken(token)
		if err != nil || clientGroup == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unauthorized"})
			return
		}
		// group := r.URL.Query().Get("group")
		channel := r.URL.Query().Get("channel")
		topics := r.URL.Query().Get("topics")
		// w.Header().Set("Content-Type", "application/json")

		r.Header.Add("channel", channel)
		r.Header.Add("topics", topics)
		r.Header.Add("portalId", domain.MAIN_PORTAL)
		r.Header.Add("groupId", clientGroup.GroupId)
		r.Header.Add("clientId", clientGroup.ClientId)
		if isAdmin {
			r.Header.Add("admin", "true")
		}
		handler(w, r)
	}
}

// Rate limit

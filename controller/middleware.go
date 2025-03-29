package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
)

// Auth
func (c *Controller) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			token = r.URL.Query().Get("Authorization")
		}

		tokens := strings.Split(token, " ")
		if len(tokens) == 2 {
			token = tokens[1]
		}
		clientGroup, permission, isAdmin, err := c.db.GetClientAuthByToken(token)
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
		permission.AddToHeaders(r)
		next.ServeHTTP(w, r)
	}
}

func (c *Controller) Admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin := r.Header.Get("admin")
		if admin != "true" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Unnecessary as is with prefixing but that may not always be the case
func (c *Controller) ChannelContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupId := r.Header.Get("groupId")
		portalId := r.Header.Get("portalId")
		channelName := r.Header.Get("channel")

		deny := func() {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unauthorized"})
		}

		if len(channelName) == 0 {
			deny()
			return
		}
		channel, err := c.db.GetChannel(domain.GroupId(channelName, groupId, portalId))
		if err != nil && errors.Is(err, datastore.ErrNotFound) {
			deny()
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Internal error"})
			return
		}
		if channel.GroupId != groupId || channel.PortalId != portalId {
			deny()
			return
		}
		next.ServeHTTP(w, r)
	}
}

func MethodGuard(allowedMethod string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

// Rate limit

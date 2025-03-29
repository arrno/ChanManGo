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
func (c *Controller) auth(next http.HandlerFunc) http.HandlerFunc {
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

// Unnecessary as is with prefixing but that may not always be the case
func (c *Controller) channelContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupId := r.Header.Get("groupId")
		portalId := r.Header.Get("portalId")
		channelName := r.Header.Get("channel")

		if len(channelName) == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		channel, err := c.db.GetChannel(domain.GroupId(channelName, groupId, portalId))
		if err != nil && errors.Is(err, datastore.ErrNotFound) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Internal error"})
			return
		}
		if channel.GroupId != groupId || channel.PortalId != portalId {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (c *Controller) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if !c.rateLimiter.Allow(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin := r.Header.Get("admin")
		if admin != "true" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func methodGuard(allowedMethod string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

package controller

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Auth
func (c *Controller) Auth(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tokens := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		var token string
		if len(tokens) == 2 {
			token = tokens[1]
		}
		if token != "auth-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unauthorized"})
			return
		}
		// w.Header().Set("Content-Type", "application/json")
		// TODO these come from auth token
		r.Header.Add("portalId", "testPortal")
		r.Header.Add("clientId", "testClient")
		handler(w, r)
	}
}

// Rate limit

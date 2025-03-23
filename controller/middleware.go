package controller

import (
	"encoding/json"
	"net/http"
	"strings"
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
		if token != "auth-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unauthorized"})
			return
		}
		channel := r.URL.Query().Get("channel")
		topics := r.URL.Query().Get("topics")
		// w.Header().Set("Content-Type", "application/json")
		// TODO these come from auth token
		r.Header.Add("channel", channel)
		r.Header.Add("topics", topics)
		r.Header.Add("portalId", "main")
		r.Header.Add("clientId", "dev")
		handler(w, r)
	}
}

// Rate limit

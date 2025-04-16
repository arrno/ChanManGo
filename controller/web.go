package controller

import "net/http"

// ENV VARS...
const PORT string = "8080"

func RunWeb() {

	c := NewController()

	// public
	r := NewRouter().
		Chain(c.rateLimit).
		Chain(c.cors)
	r.Add("/beta/signup", c.Contact, []string{http.MethodPost, http.MethodOptions})
	r.Add("/contact", c.Contact, []string{http.MethodPost, http.MethodOptions})

	rr := NewRouter().
		Chain(c.rateLimit).
		Chain(c.auth).
		Chain(c.cors)
	rr.Add("/rotate-token", c.RotateToken, []string{http.MethodPost, http.MethodOptions})
	rr.Add("/beta/check-token", c.CheckTokenBeta, []string{http.MethodGet, http.MethodOptions})
	rr.Add("/beta/update-token", c.RotateToken, []string{http.MethodPost, http.MethodOptions})

	// Channel router
	rc := NewRouter().
		Chain(c.rateLimit).
		Chain(c.auth).
		Chain(c.channelContext)
	rc.Add("/pub-ws", c.HandlePublishWS, []string{http.MethodGet, http.MethodPost}) // Support both GET and POST for WebSocket
	rc.Post("/pub", c.HandlePublish)
	rc.Add("/sub", c.HandleSubscribe, []string{http.MethodGet, http.MethodPost}) // Support both GET and POST for WebSocket

	// Admin router
	ra := NewRouter().
		Chain(c.rateLimit).
		Chain(c.auth).
		Chain(onlyAdmin)
	ra.Add(
		"/client",
		c.ClientHandler,
		[]string{http.MethodGet, http.MethodPost, http.MethodDelete},
	)
	ra.Get("/clients", c.GetClients)

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		panic(err.Error())
	}
}

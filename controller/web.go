package controller

import "net/http"

// ENV VARS...
const PORT string = "8080"

func RunWeb() {

	c := NewController()

	// public
	r := NewRouter().
		Chain(c.rateLimit)
	r.Post("/signup-demo", c.SignUpDemo)

	rr := NewRouter().
		Chain(c.rateLimit).
		Chain(c.auth)
	rr.Post("/rotate-token", c.RotateToken)

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
		Chain(admin)
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

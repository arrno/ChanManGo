package controller

import "net/http"

// ENV VARS...
const PORT string = "8080"

func RunWeb() {

	c := NewController()

	// Channel router
	rc := NewRouter().
		Chain(c.Auth).
		Chain(c.ChannelContext)
	rc.Post("/pub-ws", c.HandlePublishWS)
	rc.Post("/pub", c.HandlePublish)
	rc.Post("/sub", c.HandleSubscribe)

	// Admin router
	ra := NewRouter().
		Chain(c.Auth).
		Chain(c.Admin)
	ra.Get("/client", c.GetClient)
	ra.Get("/clients", c.GetClients)
	ra.Post("/client", c.CreateClient)
	ra.Delete("/client", c.ArchiveClient)
	ra.Post("/rotate-token", c.RotateToken)

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		panic(err.Error())
	}
}

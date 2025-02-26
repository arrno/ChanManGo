package controller

import "net/http"

// ENV VARS...
const PORT string = "8080"

func RunWeb() {

	c := NewController()

	http.HandleFunc("/publish", c.Auth(c.HandlePublish))
	http.HandleFunc("/subscribe", c.Auth(c.HandleSubscribe))

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		panic(err.Error())
	}
}

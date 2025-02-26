package controller

import (
	"net/http"

	"github.com/coder/websocket"
)

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {

	// Parse payload
	// if chan does not exist, c.CreateChannel

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer r.Body.Close()

	// var data map[string]any
	// err = json.Unmarshal(body, &data)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.CloseNow()

	// incr publisher count
	// defer decrement publisher count

	// c.PublishMessage as they come in

	conn.Close(websocket.StatusNormalClosure, "")

}

func (c *Controller) HandleSubscribe(w http.ResponseWriter, r *http.Request) {

	// Parse payload
	// if chan does not exist, err

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.CloseNow()

	// NewSubscription
	// c.AddSubscription()
	// defer unsubscribe

	conn.Close(websocket.StatusNormalClosure, "")
}

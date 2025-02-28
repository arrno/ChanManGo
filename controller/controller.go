package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/service"
	"github.com/coder/websocket"
)

type Controller struct {
	mgr service.ManagerProvider
}

func NewController() *Controller {
	return new(Controller)
}

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	topic := r.Header.Get("topic")
	channel := domain.NewChannel(portalId, topic)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing connection")

	log.Println("Client connected for reading")

	worker, err := c.mgr.ExpectChannel(channel)
	if err != nil {
		log.Printf("Failed to open channel: %v", err)
		return
	}
	decrement := worker.IncrPublishers()
	defer decrement()

	for {
		typ, msg, err := conn.Read(r.Context())
		if err != nil {
			log.Printf("Read error or client disconnected: %v", err)
			break
		}
		c.mgr.PublishMessage(channel, domain.NewMessage(clientId, msg))
		log.Printf("Received message (type %v): %s", typ, msg)
	}
}

func (c *Controller) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	topic := r.Header.Get("topic")
	channel := domain.NewChannel(portalId, topic)
	sub := domain.NewSubscription(clientId, topic)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing connection")

	log.Println("Client connected for writing")

	ctx := context.Background()

	sub.Callback = func(message *domain.Message) error {
		err := conn.Write(ctx, websocket.MessageText, message.Payload)
		if err != nil {
			log.Printf("Write error or client disconnected: %v", err)
			return err
		}
		log.Println("Sent message to client")
		return nil
	}
	c.mgr.AddSubscription(sub, channel)
	<-r.Context().Done()
}

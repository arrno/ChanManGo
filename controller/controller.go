package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/service"
	"github.com/coder/websocket"
)

type Controller struct {
	mgr service.ManagerProvider
}

type MessageWrapper struct {
	Topic   string `json:"topic"`
	Message []byte `json:"message"`
}

func NewController() *Controller {
	return &Controller{service.NewManager()}
}

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")
	channel := domain.NewChannel(portalId, channelName)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing connection")

	log.Println("Client connected for writing")

	worker, err := c.mgr.ExpectChannel(channel)
	if err != nil {
		log.Printf("Failed to open channel: %v", err)
		return
	}
	decrement := worker.IncrPublishers()
	defer decrement()

	for {
		// TODO topic should be embedded in message
		typ, msg, err := conn.Read(r.Context())
		if err != nil {
			log.Printf("Read error or client disconnected: %v", err)
			break
		}
		var wrapper MessageWrapper
		err = json.Unmarshal(msg, &wrapper)
		if err != nil {
			log.Printf("Json unmarshal err: %v", err)
			break
		}
		c.mgr.PublishMessage(channel, domain.NewMessage(clientId, wrapper.Topic, wrapper.Message))
		log.Printf("Received message (type %v): %s", typ, msg)
		// TODO send receipt
	}
}

func (c *Controller) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")
	topicSet := r.Header.Get("topics")
	var topics []string
	if len(topicSet) == 0 {
		topics = []string{"all"}
	} else {
		topics = strings.Split(topicSet, ",")
	}
	channel := domain.NewChannel(portalId, channelName)
	// todo handle topics
	sub := domain.NewTopicSubscription(clientId, topics)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing connection")

	log.Println("Client connected for reading")

	ctx := context.Background()

	sub.Callback = func(message *domain.Message) error {
		err := conn.Write(ctx, websocket.MessageText, message.Payload)
		if err != nil {
			log.Printf("Write error or client disconnected: %v", err)
			return err
		}
		// todo listen for ack
		log.Println("Sent message to client")
		return nil
	}
	if _, err := c.mgr.AddSubscription(sub, channel); err != nil {
		log.Printf("Failed to add subscription: %v", err)
		return
	}
	defer func() {
		if err := c.mgr.RemoveSubscription(sub, channel); err != nil {
			log.Printf("Failed to remove subscription: %v", err)
		}
	}()
	<-r.Context().Done()
}

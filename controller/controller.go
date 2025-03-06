package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/arrno/chanban/domain"
	"github.com/arrno/chanban/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections, modify this in production for better security
	},
}

type Controller struct {
	mgr service.ManagerProvider
}

type MessageWrapper struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func NewController() *Controller {
	return &Controller{service.NewManager()}
}

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")
	channel := domain.NewChannel(portalId, channelName)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

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
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Println("Received CloseMessage from client, closing WebSocket")
				break
			}
			log.Printf("Read error or client disconnected: %v", err)
			break
		}
		if messageType == websocket.CloseMessage {
			log.Println("Received CloseMessage from client, closing WebSocket")
			break
		}
		var wrapper MessageWrapper
		err = json.Unmarshal(msg, &wrapper)
		if err != nil {
			log.Printf("Json unmarshal err: %v", err)
			break
		}
		c.mgr.PublishMessage(channel, domain.NewMessage(clientId, wrapper.Topic, []byte(wrapper.Payload)))
		log.Printf("Received message (type %v): %s", messageType, msg)
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	done := make(chan any)
	go func() {
		defer close(done)
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					log.Println("Received CloseMessage from client, closing WebSocket")
					break
				}
				log.Printf("Read error or client disconnected: %v", err)
				break
			}
			if messageType == websocket.CloseMessage {
				log.Println("Received CloseMessage from client, closing WebSocket")
				break
			}
		}
	}()

	log.Println("Client connected for reading")

	sub.Callback = func(message *domain.Message) error {
		wrapper := MessageWrapper{
			Topic:   message.Topic,
			Payload: string(message.Payload),
		}
		b, err := json.Marshal(wrapper)
		if err != nil {
			log.Panicln("Marshal err")
			return err
		}
		err = conn.WriteMessage(websocket.TextMessage, b)
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
	<-done
}

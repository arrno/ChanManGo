package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arrno/chanman/broker"
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/service"
	ioService "github.com/arrno/chanman/service/io"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections, modify this in production for better security
	},
}

type Controller struct {
	db      datastore.DatastoreProvider
	chanMan service.ManagerProvider
	system  ioService.SystemProvider
}

type MessageWrapper struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func NewController() *Controller {
	db := datastore.NewInMemoryDB()
	datastore.SeedDb(db)
	broker := broker.NewInMemoryBroker()
	return &Controller{
		db:      db,
		chanMan: service.NewManager(db, broker),
		system:  ioService.NewSystem(),
	}
}

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {

	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")

	if channelName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "No channel provided"})
		return
	}

	channel := domain.NewChannel(portalId, channelName, 60)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to parse message"})
		return
	}
	defer r.Body.Close()

	var wrapper MessageWrapper
	err = json.Unmarshal(body, &wrapper)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to parse message"})
		return
	}
	err = c.chanMan.PublishMessage(clientId, channel, domain.NewMessage(clientId, wrapper.Topic, []byte(wrapper.Payload)))
	if err != nil && errors.Is(err, domain.ErrChanOpNotPermitted) {
		log.Error().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"Response": err.Error()})
		return
	} else if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Error encountered publishing message: " + err.Error()})
		return
	}
}

func (c *Controller) HandlePublishWS(w http.ResponseWriter, r *http.Request) {

	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")

	if channelName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "No channel provided"})
		return
	}
	channel := domain.NewChannel(portalId, channelName, 60)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to accept WebSocket connection: " + err.Error()})
		return
	}
	defer conn.Close()

	log.Info().Msg("Client connected for writing")

	worker, err := c.chanMan.ExpectChannel(channel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to open channel: " + err.Error()})
		return
	}
	decrement := worker.IncrPublishers()
	defer decrement()

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Info().Msg("Received CloseMessage from client, closing WebSocket")
				break
			}
			log.Error().Msg(fmt.Sprintf("Read error or client disconnected: %v", err))
			break
		}
		if messageType == websocket.CloseMessage {
			log.Info().Msg("Received CloseMessage from client, closing WebSocket")
			break
		}
		var wrapper MessageWrapper
		err = json.Unmarshal(msg, &wrapper)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("Json unmarshal err: %v", err))
			break
		}
		err = c.chanMan.PublishMessage(clientId, channel, domain.NewMessage(clientId, wrapper.Topic, []byte(wrapper.Payload)))
		if err != nil && errors.Is(err, domain.ErrChanOpNotPermitted) {
			log.Error().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"Response": err.Error()})
			return
		} else if err != nil {
			log.Error().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Error encountered publishing message: " + err.Error()})
			return
		}
		log.Info().Msg(fmt.Sprintf("Received message (type %v): %s", messageType, msg))
	}
}

func (c *Controller) HandleSubscribe(w http.ResponseWriter, r *http.Request) {

	clientId := r.Header.Get("clientId")
	portalId := r.Header.Get("portalId")
	channelName := r.Header.Get("channel")

	if channelName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Response": "No channel provided"})
		return
	}
	topicSet := r.Header.Get("topics")
	var topics []string
	if len(topicSet) == 0 {
		topics = []string{"all"}
	} else {
		topics = strings.Split(topicSet, ",")
	}
	channel := domain.NewChannel(portalId, channelName, 60)
	sub := domain.NewTopicSubscription(clientId, channel.Id, topics)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to accept WebSocket connection: %v", err))
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
					log.Info().Msg("Received CloseMessage from client, closing WebSocket")
					break
				}
				log.Error().Msg(fmt.Sprintf("Read error or client disconnected: %v", err))
				break
			}
			if messageType == websocket.CloseMessage {
				log.Info().Msg("Received CloseMessage from client, closing WebSocket")
				break
			}
		}
	}()

	log.Info().Msg("Client connected for reading")

	sub.Callback = func(message *domain.Message) error {
		wrapper := MessageWrapper{
			Topic:   message.Topic,
			Payload: string(message.Payload),
		}
		b, err := json.Marshal(wrapper)
		if err != nil {
			log.Error().Msg("Marshal err")
			return err
		}
		err = conn.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("Write error or client disconnected: %v", err))
			return err
		}
		log.Info().Msg("Sent message to client")
		return nil
	}

	_, err = c.chanMan.AddSubscription(sub, channel)

	if err != nil && errors.Is(err, domain.ErrChanOpNotPermitted) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"Response": err.Error()})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"Response": "Failed to add subscription: " + err.Error()})
		return
	}

	defer func() {
		if err := c.chanMan.RemoveSubscription(sub, channel); err != nil {
			log.Error().Msg(fmt.Sprintf("Failed to remove subscription: %v", err))
		}
	}()

	<-done
}

func (c *Controller) GetChannel(w http.ResponseWriter, r *http.Request)     {}
func (c *Controller) CreateChannel(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdateChannel(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchiveChannel(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) GetGroup(w http.ResponseWriter, r *http.Request)     {}
func (c *Controller) CreateGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdateGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchiveGroup(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) AddClientToGroup(w http.ResponseWriter, r *http.Request)       {}
func (c *Controller) RemoveClientFromGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) AddChannelToGroup(w http.ResponseWriter, r *http.Request)      {}
func (c *Controller) RemoveChannelFromGroup(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) CreatePortal(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdatePortal(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchivePortal(w http.ResponseWriter, r *http.Request) {}

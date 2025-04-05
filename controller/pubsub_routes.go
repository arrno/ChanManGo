package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arrno/chanman/domain"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func (c *Controller) HandlePublish(w http.ResponseWriter, r *http.Request) {

	clientId, groupId, portalId, channelName, channelId, _, write := parseHeaders(r.Header)

	if write != "true" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if channelName == "" {
		http.Error(w, "No channel provided", http.StatusBadRequest)
		return
	}

	channel, err := domain.NewChannel(channelId, portalId, groupId, channelName, 60)
	if err != nil {
		http.Error(w, "Failed to handle channel: "+err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to parse message", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var wrapper MessageWrapper
	err = json.Unmarshal(body, &wrapper)
	if err != nil {
		http.Error(w, "Failed to parse message", http.StatusBadRequest)
		return
	}
	message, err := domain.NewMessage(clientId, channel.Id, wrapper.Topic, []byte(wrapper.Payload))
	if err != nil {
		http.Error(w, "Failed to process message", http.StatusBadRequest)
		return
	}
	err = c.chanMan.PublishMessage(clientId, channel, message)

	if err != nil && errors.Is(err, domain.ErrChanOpNotPermitted) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, "Failed to publish message: %s"+err.Error(), http.StatusInternalServerError)
		return
	}
	RespondWithJSON(w, http.StatusOK, StandardResponse{
		Success: true,
	})
}

func (c *Controller) HandlePublishWS(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request details for debugging
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.RawQuery).
		Str("remote", r.RemoteAddr).
		Str("proto", r.Proto).
		Msg("WebSocket publish request received")

	clientId, groupId, portalId, channelName, channelId, _, write := parseHeaders(r.Header)

	// Log authentication details (sanitized)
	log.Info().
		Str("clientId", clientId).
		Str("groupId", groupId).
		Str("portalId", portalId).
		Str("channelName", channelName).
		Str("channelId", channelId).
		Str("write", write).
		Msg("Authentication details")

	if write != "true" {
		log.Warn().Str("clientId", clientId).Msg("Write permission denied")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if channelName == "" {
		log.Warn().Str("clientId", clientId).Msg("No channel provided")
		http.Error(w, "No channel provided", http.StatusBadRequest)
		return
	}

	channel, err := domain.NewChannel(channelId, portalId, groupId, channelName, 60)
	if err != nil {
		log.Error().Err(err).Str("clientId", clientId).Msg("Failed to handle channel")
		http.Error(w, "Failed to handle channel: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Perform WebSocket upgrade
	log.Info().Str("clientId", clientId).Str("channel", channelName).Msg("Upgrading to WebSocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Str("clientId", clientId).Msg("WebSocket upgrade failed")
		http.Error(w, "Failed to accept WebSocket connection: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Info().Str("clientId", clientId).Str("channel", channelName).Msg("WebSocket connection established")
	defer conn.Close()

	log.Info().Msg("Client connected for writing")

	worker, err := c.chanMan.ExpectChannel(channel)
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1002, "Error: Failed to open channel"))
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
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1002, "Error: Failed to unmarshal"))
			return
		}

		message, err := domain.NewMessage(clientId, channel.Id, wrapper.Topic, []byte(wrapper.Payload))
		if err != nil {
			http.Error(w, "Failed to process message", http.StatusBadRequest)
			return
		}

		err = c.chanMan.PublishMessage(clientId, channel, message)
		if err != nil {
			if errors.Is(err, domain.ErrChanOpNotPermitted) {
				log.Error().Msg(fmt.Sprintf("Auth failure. Failed to publish message: %s", err.Error()))
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "Error: Unauthorized"))
			} else {
				log.Error().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1002, "Error: Failed to publish message"))
			}
			break
		}

		log.Info().Msg(fmt.Sprintf("Received message (type %v): %s", messageType, msg))
	}
}

func (c *Controller) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request details for debugging
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.RawQuery).
		Str("remote", r.RemoteAddr).
		Str("proto", r.Proto).
		Msg("WebSocket subscribe request received")

	clientId, groupId, portalId, channelName, channelId, read, _ := parseHeaders(r.Header)

	// Log authentication details (sanitized)
	log.Info().
		Str("clientId", clientId).
		Str("groupId", groupId).
		Str("portalId", portalId).
		Str("channelName", channelName).
		Str("channelId", channelId).
		Str("read", read).
		Msg("Authentication details for subscription")

	if read != "true" {
		log.Warn().Str("clientId", clientId).Msg("Read permission denied for subscription")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if channelName == "" {
		log.Warn().Str("clientId", clientId).Msg("No channel provided for subscription")
		http.Error(w, "No channel provided", http.StatusBadRequest)
		return
	}

	topicSet := r.Header.Get("topics")
	var topics []string
	if len(topicSet) == 0 {
		topics = []string{"all"}
	} else {
		topics = strings.Split(topicSet, ",")
	}

	channel, err := domain.NewChannel(channelId, portalId, groupId, channelName, 60)
	if err != nil {
		log.Error().Err(err).Str("clientId", clientId).Msg("Failed to handle channel for subscription")
		http.Error(w, "Failed to handle channel: "+err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := domain.NewTopicSubscription("", clientId, channel.Id, topics)
	if err != nil {
		log.Error().Err(err).Str("clientId", clientId).Msg("Failed to create topic subscription")
		http.Error(w, "Failed to handle subscription: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Log subscription details
	log.Info().
		Str("clientId", clientId).
		Str("channelId", channel.Id).
		Strs("topics", topics).
		Msg("Creating subscription")

	// Perform WebSocket upgrade
	log.Info().Str("clientId", clientId).Str("channel", channelName).Msg("Upgrading to WebSocket for subscription")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Str("clientId", clientId).Msg("WebSocket upgrade failed for subscription")
		http.Error(w, "Failed to accept WebSocket connection: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Info().Str("clientId", clientId).Str("channel", channelName).Msg("WebSocket subscription connection established")
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
	if err != nil {
		if errors.Is(err, domain.ErrChanOpNotPermitted) {
			log.Error().Msg(fmt.Sprintf("Auth failure. Failed to add subscription: %s", err.Error()))
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "Error: Unauthorized"))
		} else {
			log.Error().Msg(fmt.Sprintf("Failed to add subscription: %s", err.Error()))
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1002, "Error: Failed to add subscription"))
		}
		return
	}

	defer func() {
		if err := c.chanMan.RemoveSubscription(sub, channel); err != nil {
			log.Error().Msg(fmt.Sprintf("Failed to remove subscription: %v", err))
		}
	}()

	<-done
}

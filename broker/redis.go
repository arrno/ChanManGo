package broker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arrno/chanban/domain"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	// Stream prefix for Redis keys
	streamPrefix = "channel:%s:messages"
	// Notification channel prefix for Redis Pub/Sub
	notificationPrefix = "channel:%s:notifications"
	// Metadata prefix for Redis keys
	metadataPrefix = "channel:%s:metadata"
	// Subscription prefix for Redis keys
	// subscriptionPrefix = "worker:%s:subscriptions"
	// Maximum number of messages to keep in a stream
	maxStreamLength = 1000
	// Default read batch size
	defaultBatchSize = 100
	// Stream cleanup interval
	streamCleanupInterval = 1 * time.Hour
)

// RedisBroker implements the BrokerProvider interface using Redis
type RedisBroker struct {
	// workerId         string
	// subscriptions          sync.Map
	client                 *redis.Client
	subscriptionsMu        sync.RWMutex
	subscriptionsByChannel map[string]map[string]*domain.ChannelSubscription
	subscribedTopics       sync.Map
	shutdownCh             chan struct{}
	wg                     sync.WaitGroup
}

// NewRedisBroker creates a new Redis-based message broker
func NewRedisBroker(redisAddr, redisPassword string, redisDB int) (*RedisBroker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Generate unique worker ID
	// workerId := uuid.New().String()

	broker := &RedisBroker{
		// workerId:      workerId,
		client:                 client,
		subscriptionsByChannel: make(map[string]map[string]*domain.ChannelSubscription),
		shutdownCh:             make(chan struct{}),
	}

	// Start stream cleanup goroutine
	broker.startStreamCleanup()

	return broker, nil
}

// Shutdown gracefully shuts down the broker
func (r *RedisBroker) Shutdown(ctx context.Context) error {
	close(r.shutdownCh)

	// Wait for all goroutines to finish with a timeout
	waitCh := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		return r.client.Close()
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}

// CreateChannel creates a new channel in Redis
func (r *RedisBroker) CreateChannel(channel *domain.Channel) error {
	ctx := context.Background()
	streamKey := fmt.Sprintf(streamPrefix, channel.Id)
	metadataKey := fmt.Sprintf(metadataPrefix, channel.Id)

	pipe := r.client.Pipeline()

	// Create stream with a dummy message that will be trimmed later
	// ... this doesn't actually get trimmed as is
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: map[string]interface{}{
			"init":      true,
			"createdAt": time.Now().UnixNano(),
		},
	})

	// Store channel metadata
	pipe.HSet(ctx, metadataKey,
		"id", channel.Id,
		"portalId", channel.PortalId,
		"name", channel.Name,
		"state", string(channel.State),
		"archived", channel.Archived,
		"createdAt", time.Now().UnixNano(),
	)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Trim the stream to remove initialization message if it's the only one
	r.client.XTrimMaxLen(ctx, streamKey, maxStreamLength)

	return nil
}

func (r *RedisBroker) UpdateChannel(channel *domain.Channel) error {
	ctx := context.Background()
	metadataKey := fmt.Sprintf(metadataPrefix, channel.Id)
	notificationChannel := fmt.Sprintf(notificationPrefix, channel.Id)

	pipe := r.client.Pipeline()
	// Store channel metadata
	pipe.HSet(ctx, metadataKey,
		"id", channel.Id,
		"portalId", channel.PortalId,
		"name", channel.Name,
		"state", string(channel.State),
		"archived", channel.Archived,
		"createdAt", time.Now().UnixNano(),
	)
	pipe.Publish(ctx, notificationChannel, "channel.updated")

	// Set expiry on metadata if channel is archived
	if channel.Archived {
		pipe.Expire(ctx, metadataKey, 7*24*time.Hour) // 7 days retention for archived channels
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// DestroyChannel removes a channel and its related data from Redis
func (r *RedisBroker) DestroyChannel(channel *domain.Channel) error {
	ctx := context.Background()
	streamKey := fmt.Sprintf(streamPrefix, channel.Id)
	metadataKey := fmt.Sprintf(metadataPrefix, channel.Id)
	notificationChannel := fmt.Sprintf(notificationPrefix, channel.Id)

	// Notify subscribers that the channel is being destroyed
	destroyMsg := &domain.Message{
		// ProcessedByWorkerId: r.workerId,
		Id:       uuid.New().String(),
		Topic:    "system.channel.destroyed",
		ClientId: "system",
		State:    domain.STATE_PROCESSED,
		Payload:  []byte(fmt.Sprintf(`{"channelId":"%s"}`, channel.Id)),
	}

	now := time.Now()
	destroyMsg.ProcessedAt = &now

	// Publish destroy message before removing the channel
	r.PublishMessage(ctx, channel, destroyMsg)

	// Give a small window for subscribers to process the destroy message
	time.Sleep(200 * time.Millisecond)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, streamKey)
	pipe.Del(ctx, metadataKey)
	pipe.Publish(ctx, notificationChannel, "channel.destroyed")

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to destroy channel: %w", err)
	}

	return nil
}

// PublishMessage publishes a message to a channel
func (r *RedisBroker) PublishMessage(ctx context.Context, channel *domain.Channel, m *domain.Message) error {
	streamKey := fmt.Sprintf(streamPrefix, channel.Id)
	notificationChannel := fmt.Sprintf(notificationPrefix, channel.Id)

	// Convert message to map for Redis
	values := map[string]interface{}{
		"id":                  m.Id,
		"topic":               m.Topic,
		"clientId":            m.ClientId,
		"processedByWorkerId": m.ProcessedByWorkerId,
		"processedAt":         m.ProcessedAt.UnixNano(),
		"state":               string(m.State),
		"attempts":            m.Attempts,
	}

	// Add payload as either string or byte array depending on content
	if m.Payload != nil {
		values["payload"] = m.Payload
	}

	pipe := r.client.Pipeline()

	// Add to stream
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: values,
	})

	// Trim stream to prevent unbounded growth
	pipe.XTrim(ctx, streamKey, maxStreamLength)

	// Notify other workers
	pipe.Publish(ctx, notificationChannel, "new-message")

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// SubscribeForMessages subscribes to messages on a channel
func (r *RedisBroker) SubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.ChannelSubscription) error {

	r.subscriptionsMu.Lock()
	// Initialize map for this channel if it doesn't exist
	if _, ok := r.subscriptionsByChannel[channel.Id]; !ok {
		r.subscriptionsByChannel[channel.Id] = make(map[string]*domain.ChannelSubscription)
	}
	// Store subscription directly in the channel-specific map
	r.subscriptionsByChannel[channel.Id][subscriber.Id] = subscriber
	_, alreadySubscribed := r.subscribedTopics.LoadOrStore(channel.Id, true)
	r.subscriptionsMu.Unlock()

	// Check if we're already subscribed to this channel's notifications
	if !alreadySubscribed {
		// Start a new goroutine for this channel's subscription
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.subscribeToChannelMessages(ctx, false, channel)
		}()
	}

	return nil
}

// UnsubscribeForMessages removes a subscription from a channel
func (r *RedisBroker) UnsubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.ChannelSubscription) error {
	hasSubscribers := false

	r.subscriptionsMu.Lock()
	defer r.subscriptionsMu.Unlock()

	if subscriptions, ok := r.subscriptionsByChannel[channel.Id]; ok {
		hasSubscribers = len(subscriptions) > 1
		delete(subscriptions, subscriber.Id)
	}

	// If no more subscribers for this channel, remove from subscribed topics
	if !hasSubscribers {
		r.subscribedTopics.Delete(channel.Id)
	}

	return nil
}

// subscribeToChannelMessages handles the subscription process for a specific channel
func (r *RedisBroker) subscribeToChannelMessages(ctx context.Context, withHistory bool, channel *domain.Channel) {
	streamKey := fmt.Sprintf(streamPrefix, channel.Id)
	notificationChannel := fmt.Sprintf(notificationPrefix, channel.Id)

	// Keep track of the last processed message ID
	lastId := "0" // Start from beginning

	// For subscribers that want only new messages
	if !withHistory {
		// Get the latest ID from the stream
		latestInfo, err := r.client.XInfoStream(ctx, streamKey).Result()
		if err == nil && latestInfo.LastGeneratedID != "" {
			lastId = latestInfo.LastGeneratedID
		} else {
			lastId = "$" // Fallback to the special "$" ID meaning "latest"
		}
	}

	// Subscribe to notifications
	pubsub := r.client.Subscribe(context.Background(), notificationChannel)
	defer pubsub.Close()

	// Create a context that's tied to both the provided context and the broker's shutdown
	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for context cancellation
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-r.shutdownCh:
			cancel()
		}
	}()

	// Listen for notifications and read messages
	ch := pubsub.Channel()

	// Read initial messages to catch up
	r.readAndProcessMessages(subCtx, streamKey, channel, &lastId)

	for {
		select {
		case <-subCtx.Done():
			return
		case msg := <-ch:
			log.Printf("Received message on notification channel %s: %s", notificationChannel, msg.String())
			if msg.Payload == "new-message" {
				// Read new messages when notified
				r.readAndProcessMessages(subCtx, streamKey, channel, &lastId)
			}
		case <-time.After(5 * time.Second):
			// Periodically poll for messages in case we missed a notification
			r.readAndProcessMessages(subCtx, streamKey, channel, &lastId)
		}

		// Check if we're still subscribed to this channel
		if _, ok := r.subscribedTopics.Load(channel.Id); !ok {
			return
		}
	}
}

// readAndProcessMessages reads messages from a stream and processes them
func (r *RedisBroker) readAndProcessMessages(ctx context.Context, streamKey string, channel *domain.Channel, lastId *string) {
	// Read messages
	streams, err := r.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, *lastId},
		Count:   defaultBatchSize,
		Block:   0, // No timeout
	}).Result()

	if err != nil {
		if err != context.Canceled && err != redis.Nil {
			log.Printf("Error reading stream %s: %v", streamKey, err)
		}
		return
	}

	// No messages found
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return
	}

	// Process each message
	for _, xMessage := range streams[0].Messages {
		// Update the last seen ID
		*lastId = xMessage.ID

		// Convert Redis message to domain message
		message := r.mapRedisMessageToDomainMessage(xMessage)

		// Deliver to all subscribers for this channel
		r.deliverMessageToSubscribers(channel, message)
	}
}

// deliverMessageToSubscribers sends a message to all subscribers of a channel
func (r *RedisBroker) deliverMessageToSubscribers(channel *domain.Channel, message *domain.Message) {

	// r.subscriptionsMu.RLock()
	// subscriptions := r.subscriptionsByChannel[channel.Id]
	// r.subscriptionsMu.RUnlock()

	// Collect all subscribers for this channel
	if subscriptions, ok := r.subscriptionsByChannel[channel.Id]; ok {
		for _, subscription := range subscriptions {
			if subscription.Callback != nil && subscription.State != domain.STATE_ARCHIVED {
				// Execute callback in a new goroutine to prevent blocking
				r.wg.Add(1)
				go func(sub *domain.ChannelSubscription, msg *domain.Message) {
					defer r.wg.Done()
					if err := sub.Callback(msg); err != nil {
						log.Printf("Error in subscription callback: %v", err)
					}
				}(subscription, message)
			}
		}
	}
}

// mapRedisMessageToDomainMessage converts a Redis XMessage to a domain Message
func (r *RedisBroker) mapRedisMessageToDomainMessage(xMessage redis.XMessage) *domain.Message {
	msg := &domain.Message{
		Id: getString(xMessage.Values, "id"),
	}

	// Set basic fields
	msg.Topic = getString(xMessage.Values, "topic")
	msg.ClientId = getString(xMessage.Values, "clientId")
	msg.ProcessedByWorkerId = getString(xMessage.Values, "processedByWorkerId")
	msg.State = domain.STATE(getString(xMessage.Values, "state"))
	msg.Attempts = getUint(xMessage.Values, "attempts")

	// Set timestamps
	if processedAt := getInt64(xMessage.Values, "processedAt"); processedAt > 0 {
		t := time.Unix(0, processedAt)
		msg.ProcessedAt = &t
	}

	if acknowledgedAt := getInt64(xMessage.Values, "acknowledgedAt"); acknowledgedAt > 0 {
		t := time.Unix(0, acknowledgedAt)
		msg.AcknowledgedAt = &t
	}

	// Set payload
	if payload, ok := xMessage.Values["payload"]; ok {
		switch p := payload.(type) {
		case string:
			msg.Payload = []byte(p)
		case []byte:
			msg.Payload = p
		}
	}

	return msg
}

// startStreamCleanup starts a goroutine to periodically clean up old streams
func (r *RedisBroker) startStreamCleanup() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(streamCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.cleanupStreams()
			case <-r.shutdownCh:
				return
			}
		}
	}()
}

// cleanupStreams trims all message streams and removes data for archived channels
func (r *RedisBroker) cleanupStreams() {
	ctx := context.Background()

	// Find all channel streams
	iter := r.client.Scan(ctx, 0, "channel:*:messages", 100).Iterator()

	for iter.Next(ctx) {
		streamKey := iter.Val()

		// Extract channel ID from key
		channelId := extractChannelIdFromKey(streamKey)
		if channelId == "" {
			continue
		}

		metadataKey := fmt.Sprintf(metadataPrefix, channelId)

		// Check if channel is archived
		archived, err := r.client.HGet(ctx, metadataKey, "archived").Bool()
		if err != nil && err != redis.Nil {
			log.Printf("Error checking if channel %s is archived: %v", channelId, err)
			continue
		}

		if archived {
			// For archived channels, keep only a small number of messages
			r.client.XTrimMaxLen(ctx, streamKey, 100)
		} else {
			// For active channels, keep the configured maximum
			r.client.XTrimMaxLen(ctx, streamKey, maxStreamLength)
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("Error scanning for streams: %v", err)
	}
}

// Helper functions for converting Redis values
func getString(values map[string]interface{}, key string) string {
	if val, ok := values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt64(values map[string]interface{}, key string) int64 {
	if val, ok := values[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case string:
			if i, err := time.ParseDuration(v); err == nil {
				return int64(i)
			}
		}
	}
	return 0
}

func getUint(values map[string]interface{}, key string) uint {
	if val, ok := values[key]; ok {
		switch v := val.(type) {
		case uint:
			return v
		case int:
			if v >= 0 {
				return uint(v)
			}
		case float64:
			if v >= 0 {
				return uint(v)
			}
		case string:
			var u uint
			fmt.Sscanf(v, "%d", &u)
			return u
		}
	}
	return 0
}

// extractChannelIdFromKey extracts the channel ID from a Redis key
func extractChannelIdFromKey(key string) string {
	var channelId string
	fmt.Sscanf(key, "channel:%s:messages", &channelId)
	return channelId
}

package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/arrno/chanman/domain"
	"github.com/jackc/pgx/v5"
)

// MessageRepo implements the MessageRepoProvider interface using PostgreSQL
type MessageRepo struct {
	db SQLDatabase
}

// NewMessageRepo creates a new PostgreSQL message repository
func NewMessageRepo(db SQLDatabase) *MessageRepo {
	return &MessageRepo{
		db: db,
	}
}

// StoreChannel stores a channel in the database
func (r *MessageRepo) StoreChannel(channel *domain.Channel) error {
	query := `
		INSERT INTO channels (
			id, portal_id, group_id, name, state, archived, interval_seconds, 
			heartbeat, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			portal_id = $2,
			group_id = $3,
			name = $4,
			state = $5,
			archived = $6,
			interval_seconds = $7,
			heartbeat = $8,
			updated_at = $10,
			archived_at = $11
	`

	_, err := r.db.Exec(ctx, query,
		channel.Id, channel.PortalId, channel.GroupId, channel.Name,
		channel.State, channel.Archived, channel.IntervalSeconds,
		channel.Heartbeat, channel.CreatedAt, channel.UpdatedAt, channel.ArchivedAt)

	return err
}

func (r *MessageRepo) GetChannelByGroupAndName(groupId, channelName string) (*domain.Channel, error) {
	query := `
		SELECT id, portal_id, group_id, name, state, archived, interval_seconds, 
			heartbeat, created_at, updated_at, archived_at
		FROM channels
		WHERE group_id = $1 AND name = $2
	`

	var channel domain.Channel
	err := r.db.Get(ctx, &channel, query, groupId, channelName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &channel, nil
}

// GetChannel retrieves a channel by ID
func (r *MessageRepo) GetChannel(id string) (*domain.Channel, error) {
	query := `
		SELECT id, portal_id, group_id, name, state, archived, interval_seconds, 
			heartbeat, created_at, updated_at, archived_at
		FROM channels
		WHERE id = $1
	`

	var channel domain.Channel
	err := r.db.Get(ctx, &channel, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &channel, nil
}

// UpdateChannel updates a channel in the database
func (r *MessageRepo) UpdateChannel(channel *domain.Channel) error {
	query := `
		UPDATE channels
		SET portal_id = $2, group_id = $3, name = $4, state = $5, 
			archived = $6, interval_seconds = $7, heartbeat = $8, 
			updated_at = $9, archived_at = $10
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		channel.Id, channel.PortalId, channel.GroupId, channel.Name,
		channel.State, channel.Archived, channel.IntervalSeconds,
		channel.Heartbeat, channel.UpdatedAt, channel.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveChannel archives a channel by ID
func (r *MessageRepo) ArchiveChannel(id string) error {
	// First get the channel to ensure it exists
	channel, err := r.GetChannel(id)
	if err != nil {
		return err
	}

	// Archive the channel
	now := channel.UpdatedAt
	channel.Archived = true
	channel.State = domain.STATE_ARCHIVED
	channel.ArchivedAt = &now
	channel.UpdatedAt = now

	return r.UpdateChannel(channel)
}

// StoreTopicSubscriber stores a topic subscription in the database
func (r *MessageRepo) StoreTopicSubscriber(sub *domain.TopicSubscription) error {
	query := `
		INSERT INTO topic_subscriptions (
			id, client_id, channel_id, topics, state, archived, 
			created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			client_id = $2,
			channel_id = $3,
			topics = $4,
			state = $5,
			archived = $6,
			updated_at = $8,
			archived_at = $9
	`

	_, err := r.db.Exec(ctx, query,
		sub.Id, sub.ClientId, sub.ChannelId, sub.Topics, sub.State,
		sub.Archived, sub.CreatedAt, sub.UpdatedAt, sub.ArchivedAt)

	return err
}

// GetTopicSubscriber retrieves a topic subscription by ID
func (r *MessageRepo) GetTopicSubscriber(id string) (*domain.TopicSubscription, error) {
	query := `
		SELECT id, client_id, channel_id, topics, state, archived, 
			created_at, updated_at, archived_at
		FROM topic_subscriptions
		WHERE id = $1
	`

	var sub domain.TopicSubscription
	err := r.db.Get(ctx, &sub, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &sub, nil
}

// UpdateTopicSubscriber updates a topic subscription in the database
func (r *MessageRepo) UpdateTopicSubscriber(sub *domain.TopicSubscription) error {
	query := `
		UPDATE topic_subscriptions
		SET client_id = $2, channel_id = $3, topics = $4, state = $5,
			archived = $6, updated_at = $7, archived_at = $8
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		sub.Id, sub.ClientId, sub.ChannelId, sub.Topics,
		sub.State, sub.Archived, sub.UpdatedAt, sub.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveTopicSubscriber archives a topic subscription by ID
func (r *MessageRepo) ArchiveTopicSubscriber(id string) error {
	// First get the subscription to ensure it exists
	sub, err := r.GetTopicSubscriber(id)
	if err != nil {
		return err
	}

	// Archive the subscription
	now := sub.UpdatedAt
	sub.Archived = true
	sub.State = domain.STATE_ARCHIVED
	sub.ArchivedAt = &now
	sub.UpdatedAt = now

	return r.UpdateTopicSubscriber(sub)
}

// StoreChannelSubscriber stores a channel subscription in the database
func (r *MessageRepo) StoreChannelSubscriber(sub *domain.ChannelSubscription) error {
	// Get the portal ID from the channel
	// TODO this shouldn't be necessary
	var portalID string
	channelQuery := `SELECT portal_id FROM channels WHERE id = $1`
	err := r.db.Get(ctx, &portalID, channelQuery, sub.ChannelId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("channel not found: %w", ErrNotFound)
		}
		return fmt.Errorf("error getting channel: %w", err)
	}

	query := `
		INSERT INTO channel_subscriptions (
			id, portal_id, channel_id, state, archived, 
			created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			portal_id = $2,
			channel_id = $3,
			state = $4,
			archived = $5,
			updated_at = $7,
			archived_at = $8
	`

	_, err = r.db.Exec(ctx, query,
		sub.Id, portalID, sub.ChannelId, sub.State,
		sub.Archived, sub.CreatedAt, sub.UpdatedAt, sub.ArchivedAt)

	return err
}

// GetChannelSubscriber retrieves a channel subscription by ID
func (r *MessageRepo) GetChannelSubscriber(id string) (*domain.ChannelSubscription, error) {
	query := `
		SELECT id, channel_id, state, archived, 
			created_at, updated_at, archived_at
		FROM channel_subscriptions
		WHERE id = $1
	`

	var sub domain.ChannelSubscription
	err := r.db.Get(ctx, &sub, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &sub, nil
}

// UpdateChannelSubscriber updates a channel subscription in the database
func (r *MessageRepo) UpdateChannelSubscriber(sub *domain.ChannelSubscription) error {
	query := `
		UPDATE channel_subscriptions
		SET channel_id = $2, state = $3, archived = $4, 
			updated_at = $5, archived_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		sub.Id, sub.ChannelId, sub.State, sub.Archived,
		sub.UpdatedAt, sub.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveChannelSubscriber archives a channel subscription by ID
func (r *MessageRepo) ArchiveChannelSubscriber(id string) error {
	// First get the subscription to ensure it exists
	sub, err := r.GetChannelSubscriber(id)
	if err != nil {
		return err
	}

	// Archive the subscription
	now := sub.UpdatedAt
	sub.Archived = true
	sub.State = domain.STATE_ARCHIVED
	sub.ArchivedAt = &now
	sub.UpdatedAt = now

	return r.UpdateChannelSubscriber(sub)
}

// StoreMessage stores a message in the database
func (r *MessageRepo) StoreMessage(msg *domain.Message) error {
	query := `
		INSERT INTO messages (
			id, topic, client_id, channel_id, processed_at, acknowledged_at, destinations,
			state, archived, payload, attempts, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			processed_at = $5,
			acknowledged_at = $6,
			destinations = $7,
			state = $8,
			archived = $9,
			attempts = $11,
			updated_at = $13,
			archived_at = $14
	`

	_, err := r.db.Exec(ctx, query,
		msg.Id, msg.Topic, msg.ClientId, msg.ChannelId, msg.ProcessedAt, msg.AcknowledgedAt, msg.Destinations,
		msg.State, msg.Archived, msg.Payload, msg.Attempts, msg.CreatedAt, msg.UpdatedAt, msg.ArchivedAt)

	return err
}

// GetMessage retrieves a message by ID
func (r *MessageRepo) GetMessage(id string) (*domain.Message, error) {
	query := `
		SELECT id, topic, client_id, channel_id, processed_at, acknowledged_at, destinations,
			state, archived, payload, attempts, created_at, updated_at, archived_at
		FROM messages
		WHERE id = $1
	`

	var msg domain.Message
	err := r.db.Get(ctx, &msg, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &msg, nil
}

func (r *MessageRepo) PruneMessages(retentionHours uint) error {

	return r.db.WithTx(context.Background(), func(tx SQLDatabase) error {
		query := `DELETE FROM messages WHERE created_at < now() - ($1 * INTERVAL '1 hour')`

		_, err := tx.Exec(ctx, query, retentionHours)
		if err != nil {
			return err
		}

		query = `DELETE FROM message_subscribers WHERE created_at < now() - ($1 * INTERVAL '1 hour')`

		_, err = tx.Exec(ctx, query, retentionHours)
		if err != nil {
			return err
		}

		query = `DELETE FROM topic_subscriptions WHERE created_at < now() - ($1 * INTERVAL '1 hour') AND state = 'inactive'`

		_, err = tx.Exec(ctx, query, retentionHours)
		if err != nil {
			return err
		}
		return nil
	})

}

// UpdateMessage updates a message in the database
func (r *MessageRepo) UpdateMessage(msg *domain.Message) error {
	query := `
		UPDATE messages
		SET topic = $2, client_id = $3, channel_id = $4, processed_at = $5,
			acknowledged_at = $6, destinations = $7, state = $8, archived = $9, payload = $10,
			attempts = $11, updated_at = $12, archived_at = $13
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		msg.Id, msg.Topic, msg.ClientId, msg.ChannelId, msg.ProcessedAt,
		msg.AcknowledgedAt, msg.Destinations, msg.State, msg.Archived, msg.Payload,
		msg.Attempts, msg.UpdatedAt, msg.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveMessage archives a message by ID
func (r *MessageRepo) ArchiveMessage(id string) error {
	// First get the message to ensure it exists
	msg, err := r.GetMessage(id)
	if err != nil {
		return err
	}

	// Archive the message
	now := msg.UpdatedAt
	msg.Archived = true
	msg.State = domain.STATE_ARCHIVED
	msg.ArchivedAt = &now
	msg.UpdatedAt = now

	return r.UpdateMessage(msg)
}

// StoreMessageSubscribers stores a batch of message subscribers in the database using bulk insert
func (r *MessageRepo) StoreMessageSubscribers(subscribers []*domain.MessageSubscriber) error {
	if len(subscribers) == 0 {
		return nil // Nothing to store
	}

	// Build the bulk insert query with value placeholders
	// Format: INSERT INTO message_subscribers VALUES ($1,$2,$3,$4), ($5,$6,$7,$8), ...
	query := "INSERT INTO message_subscribers (id, message_id, client_id, created_at) VALUES "
	args := []interface{}{}

	for i, sub := range subscribers {
		// Add placeholders for this row
		if i > 0 {
			query += ", "
		}

		// Calculate the position of parameters for this record
		pos := i * 4
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)",
			pos+1, pos+2, pos+3, pos+4)

		// Add values to the args slice
		args = append(args, sub.Id, sub.MessageId, sub.ClientId, sub.CreatedAt)
	}

	// Add conflict handling
	query += " ON CONFLICT (id) DO NOTHING"

	// Execute the bulk insert
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert message subscribers: %w", err)
	}

	return nil
}

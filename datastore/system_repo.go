package datastore

// import (
// 	"fmt"
// 	"time"

// 	"github.com/arrno/chanman/domain"
// 	"github.com/jackc/pgx"
// )

// // TODO
// // PostgresRepo implements RepoProvider with PostgreSQL
// type PostgresRepo struct {
// 	db SQLDatabase
// }

// // NewPostgresRepo creates a new PostgreSQL repository
// func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
// 	db, err := NewPostgresDB(dsn)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}
// 	return &PostgresRepo{db: db}, nil
// }

// // WithTx implements transactions for the repository
// func (r *PostgresRepo) WithTx(f func(tx RepoProvider) error) error {
// 	return r.db.WithTx(ctx, func(sqlTx SQLDatabase) error {
// 		txRepo := &PostgresRepo{db: sqlTx}
// 		return f(txRepo)
// 	})
// }

// // Channel methods

// func (r *PostgresRepo) StoreChannel(ch *domain.Channel) error {
// 	query := `
// 		INSERT INTO channels (
// 			id, portal_id, group_id, name, state, archived, interval_seconds,
// 			heartbeat, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
// 		ON CONFLICT (id) DO UPDATE SET
// 			state = $5,
// 			archived = $6,
// 			interval_seconds = $7,
// 			heartbeat = $8,
// 			updated_at = $10,
// 			archived_at = $11
// 	`
// 	_, err := r.db.Exec(ctx, query,
// 		ch.Id, ch.PortalId, ch.GroupId, ch.Name, ch.State, ch.Archived, ch.IntervalSeconds,
// 		ch.Heartbeat, ch.CreatedAt, ch.UpdatedAt, ch.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetChannel(id string) (*domain.Channel, error) {
// 	query := `
// 		SELECT id, portal_id, group_id, name, state, archived, interval_seconds,
// 		       heartbeat, created_at, updated_at, archived_at
// 		FROM channels
// 		WHERE id = $1 AND archived = false
// 	`

// 	var ch domain.Channel
// 	err := r.db.QueryRowStruct(ctx, &ch, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &ch, nil
// }

// func (r *PostgresRepo) UpdateChannel(ch *domain.Channel) error {
// 	ch.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE channels SET
// 			state = $1,
// 			archived = $2,
// 			interval_seconds = $3,
// 			heartbeat = $4,
// 			updated_at = $5,
// 			archived_at = $6
// 		WHERE id = $7
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		ch.State, ch.Archived, ch.IntervalSeconds, ch.Heartbeat,
// 		ch.UpdatedAt, ch.ArchivedAt, ch.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveChannel(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE channels SET
// 			archived = true,
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Topic subscription methods

// func (r *PostgresRepo) StoreTopicSubscriber(sub *domain.TopicSubscription) error {
// 	if sub.CreatedAt.IsZero() {
// 		sub.CreatedAt = time.Now()
// 	}
// 	if sub.UpdatedAt.IsZero() {
// 		sub.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO topic_subscriptions (
// 			id, client_id, channel_id, topics, state, archived, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 		ON CONFLICT (id) DO UPDATE SET
// 			topics = $4,
// 			state = $5,
// 			archived = $6,
// 			updated_at = $8,
// 			archived_at = $9
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		sub.Id, sub.ClientId, sub.ChannelId, sub.Topics, sub.State,
// 		sub.Archived, sub.CreatedAt, sub.UpdatedAt, sub.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetTopicSubscriber(id string) (*domain.TopicSubscription, error) {
// 	query := `
// 		SELECT id, client_id, channel_id, topics, state, archived, created_at, updated_at, archived_at
// 		FROM topic_subscriptions
// 		WHERE id = $1 AND archived = false
// 	`

// 	var sub domain.TopicSubscription
// 	err := r.db.QueryRowStruct(ctx, &sub, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &sub, nil
// }

// func (r *PostgresRepo) UpdateTopicSubscriber(sub *domain.TopicSubscription) error {
// 	sub.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE topic_subscriptions SET
// 			topics = $1,
// 			state = $2,
// 			archived = $3,
// 			updated_at = $4,
// 			archived_at = $5
// 		WHERE id = $6
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		sub.Topics, sub.State, sub.Archived, sub.UpdatedAt, sub.ArchivedAt, sub.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveTopicSubscriber(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE topic_subscriptions SET
// 			archived = true,
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Channel subscription methods

// func (r *PostgresRepo) StoreChannelSubscriber(sub *domain.ChannelSubscription) error {
// 	if sub.CreatedAt.IsZero() {
// 		sub.CreatedAt = time.Now()
// 	}
// 	if sub.UpdatedAt.IsZero() {
// 		sub.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO channel_subscriptions (
// 			id, channel_id, state, archived, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (id) DO UPDATE SET
// 			state = $3,
// 			archived = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		sub.Id, sub.ChannelId, sub.State, sub.Archived,
// 		sub.CreatedAt, sub.UpdatedAt, sub.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetChannelSubscriber(id string) (*domain.ChannelSubscription, error) {
// 	query := `
// 		SELECT id, channel_id, state, archived, created_at, updated_at, archived_at
// 		FROM channel_subscriptions
// 		WHERE id = $1 AND archived = false
// 	`

// 	var sub domain.ChannelSubscription
// 	err := r.db.QueryRowStruct(ctx, &sub, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &sub, nil
// }

// func (r *PostgresRepo) UpdateChannelSubscriber(sub *domain.ChannelSubscription) error {
// 	sub.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE channel_subscriptions SET
// 			state = $1,
// 			archived = $2,
// 			updated_at = $3,
// 			archived_at = $4
// 		WHERE id = $5
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		sub.State, sub.Archived, sub.UpdatedAt, sub.ArchivedAt, sub.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveChannelSubscriber(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE channel_subscriptions SET
// 			archived = true,
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Message methods

// func (r *PostgresRepo) StoreMessage(msg *domain.Message) error {
// 	if msg.CreatedAt.IsZero() {
// 		msg.CreatedAt = time.Now()
// 	}
// 	if msg.UpdatedAt.IsZero() {
// 		msg.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO messages (
// 			id, topic, client_id, channel_id, processed_at, acknowledged_at,
// 			state, archived, payload, attempts, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
// 		ON CONFLICT (id) DO UPDATE SET
// 			processed_at = $5,
// 			acknowledged_at = $6,
// 			state = $7,
// 			archived = $8,
// 			attempts = $10,
// 			updated_at = $12,
// 			archived_at = $13
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		msg.Id, msg.Topic, msg.ClientId, msg.ChannelId, msg.ProcessedAt, msg.AcknowledgedAt,
// 		msg.State, msg.Archived, msg.Payload, msg.Attempts, msg.CreatedAt, msg.UpdatedAt, msg.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetMessage(id string) (*domain.Message, error) {
// 	query := `
// 		SELECT id, topic, client_id, channel_id, processed_at, acknowledged_at,
// 		       state, archived, payload, attempts, created_at, updated_at, archived_at
// 		FROM messages
// 		WHERE id = $1
// 	`

// 	var msg domain.Message
// 	err := r.db.QueryRowStruct(ctx, &msg, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &msg, nil
// }

// func (r *PostgresRepo) UpdateMessage(msg *domain.Message) error {
// 	msg.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE messages SET
// 			processed_at = $1,
// 			acknowledged_at = $2,
// 			state = $3,
// 			archived = $4,
// 			attempts = $5,
// 			updated_at = $6,
// 			archived_at = $7
// 		WHERE id = $8
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		msg.ProcessedAt, msg.AcknowledgedAt, msg.State, msg.Archived,
// 		msg.Attempts, msg.UpdatedAt, msg.ArchivedAt, msg.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveMessage(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE messages SET
// 			archived = true,
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Client methods

// func (r *PostgresRepo) StoreClient(client *domain.Client) error {
// 	if client.CreatedAt.IsZero() {
// 		client.CreatedAt = time.Now()
// 	}
// 	if client.UpdatedAt.IsZero() {
// 		client.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO clients (
// 			id, name, active, admin, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (id) DO UPDATE SET
// 			name = $2,
// 			active = $3,
// 			admin = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		client.Id, client.Name, client.Active, client.Admin,
// 		client.CreatedAt, client.UpdatedAt, client.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetClient(id string) (*domain.Client, error) {
// 	query := `
// 		SELECT id, name, active, admin, created_at, updated_at, archived_at
// 		FROM clients
// 		WHERE id = $1
// 	`

// 	var client domain.Client
// 	err := r.db.QueryRowStruct(ctx, &client, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &client, nil
// }

// func (r *PostgresRepo) GetClients(pagination domain.Pagination[*domain.Client]) ([]*domain.Client, error) {
// 	query := `
// 		SELECT id, name, active, admin, created_at, updated_at, archived_at
// 		FROM clients
// 		WHERE active = true
// 		ORDER BY name
// 		LIMIT $1 OFFSET $2
// 	`

// 	countQuery := `SELECT COUNT(*) FROM clients WHERE active = true`

// 	var count int
// 	err := r.db.QueryRow(ctx, countQuery).Scan(&count)
// 	if err != nil {
// 		return nil, err
// 	}

// 	pagination.Found = count
// 	offset := (pagination.Page - 1) * pagination.Count

// 	var clients []*domain.Client
// 	err = r.db.QueryStruct(ctx, &clients, query, pagination.Count, offset)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return clients, nil
// }

// func (r *PostgresRepo) UpdateClient(client *domain.Client) error {
// 	client.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE clients SET
// 			name = $1,
// 			active = $2,
// 			admin = $3,
// 			updated_at = $4,
// 			archived_at = $5
// 		WHERE id = $6
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		client.Name, client.Active, client.Admin,
// 		client.UpdatedAt, client.ArchivedAt, client.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveClient(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE clients SET
// 			active = false,
// 			archived_at = $1,
// 			updated_at = $1
// 		WHERE id = $2
// 	`

// 	result, err := r.db.Exec(ctx, query, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Client auth methods

// func (r *PostgresRepo) GetClientAuthByToken(token string) (*domain.ClientGroup, *domain.ClientPermission, bool, error) {
// 	// First, get the client secret by token
// 	secretQuery := `
// 		SELECT client_id FROM client_secret_tokens
// 		WHERE encrypted_secret = $1 AND state = 'active'
// 	`

// 	var clientId string
// 	err := r.db.QueryRow(ctx, secretQuery, token).Scan(&clientId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, nil, false, ErrNotFound
// 		}
// 		return nil, nil, false, err
// 	}

// 	// Get client groups
// 	groupsQuery := `
// 		SELECT cg.id, cg.client_id, cg.group_id, cg.state, cg.created_at, cg.updated_at, cg.archived_at
// 		FROM client_groups cg
// 		WHERE cg.client_id = $1 AND cg.state = 'active'
// 		LIMIT 1
// 	`

// 	var clientGroup domain.ClientGroup
// 	err = r.db.QueryRowStruct(ctx, &clientGroup, groupsQuery, clientId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, nil, false, ErrNotFound
// 		}
// 		return nil, nil, false, err
// 	}

// 	// Get client permissions
// 	permQuery := `
// 		SELECT id, client_id, read, write, created_at, updated_at, archived_at
// 		FROM client_permissions
// 		WHERE client_id = $1
// 	`

// 	var permission domain.ClientPermission
// 	err = r.db.QueryRowStruct(ctx, &permission, permQuery, clientId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, nil, false, ErrNotFound
// 		}
// 		return nil, nil, false, err
// 	}

// 	// Check if client is admin
// 	adminQuery := `SELECT admin FROM clients WHERE id = $1`

// 	var isAdmin bool
// 	err = r.db.QueryRow(ctx, adminQuery, clientId).Scan(&isAdmin)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, nil, false, ErrNotFound
// 		}
// 		return nil, nil, false, err
// 	}

// 	return &clientGroup, &permission, isAdmin, nil
// }

// // Client secret methods

// func (r *PostgresRepo) StoreClientSecret(secret *domain.ClientSecret) error {
// 	if secret.CreatedAt.IsZero() {
// 		secret.CreatedAt = time.Now()
// 	}
// 	if secret.UpdatedAt.IsZero() {
// 		secret.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO client_secret_tokens (
// 			id, client_id, encrypted_secret, state, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (id) DO UPDATE SET
// 			encrypted_secret = $3,
// 			state = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		secret.Id, secret.ClientId, secret.EncryptedSecret, "active",
// 		secret.CreatedAt, secret.UpdatedAt, secret.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetClientSecret(id string) (*domain.ClientSecret, error) {
// 	query := `
// 		SELECT id, client_id, encrypted_secret, created_at, updated_at, archived_at
// 		FROM client_secret_tokens
// 		WHERE id = $1 AND state = 'active'
// 	`

// 	var secret domain.ClientSecret
// 	err := r.db.QueryRowStruct(ctx, &secret, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &secret, nil
// }

// func (r *PostgresRepo) GetClientSecretByClientId(clientId string) (*domain.ClientSecret, error) {
// 	query := `
// 		SELECT id, client_id, encrypted_secret, created_at, updated_at, archived_at
// 		FROM client_secret_tokens
// 		WHERE client_id = $1 AND state = 'active'
// 		ORDER BY created_at DESC
// 		LIMIT 1
// 	`

// 	var secret domain.ClientSecret
// 	err := r.db.QueryRowStruct(ctx, &secret, query, clientId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &secret, nil
// }

// func (r *PostgresRepo) UpdateClientSecret(secret *domain.ClientSecret) error {
// 	secret.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE client_secret_tokens SET
// 			encrypted_secret = $1,
// 			updated_at = $2,
// 			archived_at = $3
// 		WHERE id = $4
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		secret.EncryptedSecret, secret.UpdatedAt, secret.ArchivedAt, secret.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveClientSecret(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE client_secret_tokens SET
// 			state = 'archived',
// 			archived_at = $1,
// 			updated_at = $1
// 		WHERE id = $2
// 	`

// 	result, err := r.db.Exec(ctx, query, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Client permission methods

// func (r *PostgresRepo) StoreClientPermission(perm *domain.ClientPermission) error {
// 	if perm.CreatedAt.IsZero() {
// 		perm.CreatedAt = time.Now()
// 	}
// 	if perm.UpdatedAt.IsZero() {
// 		perm.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO client_permissions (
// 			id, client_id, read, write, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (id) DO UPDATE SET
// 			read = $3,
// 			write = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		perm.Id, perm.ClientId, perm.Read, perm.Write,
// 		perm.CreatedAt, perm.UpdatedAt, perm.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetClientPermission(id string) (*domain.ClientPermission, error) {
// 	query := `
// 		SELECT id, client_id, read, write, created_at, updated_at, archived_at
// 		FROM client_permissions
// 		WHERE id = $1
// 	`

// 	var perm domain.ClientPermission
// 	err := r.db.QueryRowStruct(ctx, &perm, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &perm, nil
// }

// func (r *PostgresRepo) GetClientPermissionByClientId(clientId string) (*domain.ClientPermission, error) {
// 	query := `
// 		SELECT id, client_id, read, write, created_at, updated_at, archived_at
// 		FROM client_permissions
// 		WHERE client_id = $1
// 	`

// 	var perm domain.ClientPermission
// 	err := r.db.QueryRowStruct(ctx, &perm, query, clientId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &perm, nil
// }

// func (r *PostgresRepo) UpdateClientPermission(perm *domain.ClientPermission) error {
// 	perm.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE client_permissions SET
// 			read = $1,
// 			write = $2,
// 			updated_at = $3,
// 			archived_at = $4
// 		WHERE id = $5
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		perm.Read, perm.Write, perm.UpdatedAt, perm.ArchivedAt, perm.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveClientPermission(id string) error {
// 	now := time.Now()

// 	// First get the permission to get client ID
// 	perm, err := r.GetClientPermission(id)
// 	if err != nil {
// 		return err
// 	}

// 	// Update the permission
// 	query := `
// 		UPDATE client_permissions SET
// 			archived_at = $1,
// 			updated_at = $1
// 		WHERE id = $2
// 	`

// 	result, err := r.db.Exec(ctx, query, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Group methods

// func (r *PostgresRepo) StoreGroup(group *domain.Group) error {
// 	if group.CreatedAt.IsZero() {
// 		group.CreatedAt = time.Now()
// 	}
// 	if group.UpdatedAt.IsZero() {
// 		group.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO groups (
// 			id, parent_group_id, owner_id, portal_id, name, state, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 		ON CONFLICT (id) DO UPDATE SET
// 			parent_group_id = $2,
// 			name = $5,
// 			state = $6,
// 			updated_at = $8,
// 			archived_at = $9
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		group.Id, group.ParentGroupId, group.OwnerId, group.PortalId, group.Name,
// 		group.State, group.CreatedAt, group.UpdatedAt, group.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetGroup(id string) (*domain.Group, error) {
// 	query := `
// 		SELECT id, parent_group_id, owner_id, portal_id, name, state, created_at, updated_at, archived_at
// 		FROM groups
// 		WHERE id = $1
// 	`

// 	var group domain.Group
// 	err := r.db.QueryRowStruct(ctx, &group, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &group, nil
// }

// func (r *PostgresRepo) UpdateGroup(group *domain.Group) error {
// 	group.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE groups SET
// 			parent_group_id = $1,
// 			name = $2,
// 			state = $3,
// 			updated_at = $4,
// 			archived_at = $5
// 		WHERE id = $6
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		group.ParentGroupId, group.Name, group.State,
// 		group.UpdatedAt, group.ArchivedAt, group.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveGroup(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE groups SET
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Client group methods

// func (r *PostgresRepo) StoreClientGroup(cg *domain.ClientGroup) error {
// 	if cg.CreatedAt.IsZero() {
// 		cg.CreatedAt = time.Now()
// 	}
// 	if cg.UpdatedAt.IsZero() {
// 		cg.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO client_groups (
// 			id, client_id, group_id, state, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (client_id, group_id) DO UPDATE SET
// 			id = $1,
// 			state = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		cg.Id, cg.ClientId, cg.GroupId, cg.State,
// 		cg.CreatedAt, cg.UpdatedAt, cg.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetClientGroup(id string) (*domain.ClientGroup, error) {
// 	query := `
// 		SELECT id, client_id, group_id, state, created_at, updated_at, archived_at
// 		FROM client_groups
// 		WHERE id = $1
// 	`

// 	var cg domain.ClientGroup
// 	err := r.db.QueryRowStruct(ctx, &cg, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &cg, nil
// }

// func (r *PostgresRepo) GetClientGroupsByClientId(clientId string) ([]*domain.ClientGroup, error) {
// 	query := `
// 		SELECT id, client_id, group_id, state, created_at, updated_at, archived_at
// 		FROM client_groups
// 		WHERE client_id = $1 AND state != 'archived'
// 	`

// 	var groups []*domain.ClientGroup
// 	err := r.db.QueryStruct(ctx, &groups, query, clientId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(groups) == 0 {
// 		return nil, ErrNotFound
// 	}

// 	return groups, nil
// }

// func (r *PostgresRepo) UpdateClientGroup(cg *domain.ClientGroup) error {
// 	cg.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE client_groups SET
// 			state = $1,
// 			updated_at = $2,
// 			archived_at = $3
// 		WHERE id = $4
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		cg.State, cg.UpdatedAt, cg.ArchivedAt, cg.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchiveClientGroup(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE client_groups SET
// 			state = $1,
// 			archived_at = $2,
// 			updated_at = $2
// 		WHERE id = $3
// 	`

// 	result, err := r.db.Exec(ctx, query, domain.STATE_ARCHIVED, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Portal methods

// func (r *PostgresRepo) StorePortal(portal *domain.Portal) error {
// 	if portal.CreatedAt.IsZero() {
// 		portal.CreatedAt = time.Now()
// 	}
// 	if portal.UpdatedAt.IsZero() {
// 		portal.UpdatedAt = time.Now()
// 	}

// 	query := `
// 		INSERT INTO portals (
// 			id, name, owner_id, active, created_at, updated_at, archived_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		ON CONFLICT (id) DO UPDATE SET
// 			name = $2,
// 			active = $4,
// 			updated_at = $6,
// 			archived_at = $7
// 	`

// 	_, err := r.db.Exec(ctx, query,
// 		portal.Id, portal.Name, portal.OwnerId, portal.Active,
// 		portal.CreatedAt, portal.UpdatedAt, portal.ArchivedAt)

// 	return err
// }

// func (r *PostgresRepo) GetPortal(id string) (*domain.Portal, error) {
// 	query := `
// 		SELECT id, name, owner_id, active, created_at, updated_at, archived_at
// 		FROM portals
// 		WHERE id = $1
// 	`

// 	var portal domain.Portal
// 	err := r.db.QueryRowStruct(ctx, &portal, query, id)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}

// 	return &portal, nil
// }

// func (r *PostgresRepo) UpdatePortal(portal *domain.Portal) error {
// 	portal.UpdatedAt = time.Now()
// 	query := `
// 		UPDATE portals SET
// 			name = $1,
// 			active = $2,
// 			updated_at = $3,
// 			archived_at = $4
// 		WHERE id = $5
// 	`

// 	result, err := r.db.Exec(ctx, query,
// 		portal.Name, portal.Active, portal.UpdatedAt, portal.ArchivedAt, portal.Id)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// func (r *PostgresRepo) ArchivePortal(id string) error {
// 	now := time.Now()
// 	query := `
// 		UPDATE portals SET
// 			active = false,
// 			archived_at = $1,
// 			updated_at = $1
// 		WHERE id = $2
// 	`

// 	result, err := r.db.Exec(ctx, query, now, id)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return ErrNotFound
// 	}

// 	return nil
// }

// // Ensure PostgresRepo implements RepoProvider interface
// var _ RepoProvider = (*PostgresRepo)(nil)

package datastore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arrno/chanman/domain"
	"github.com/jackc/pgx/v5"
)

// SystemRepo implements the SystemRepoProvider interface using PostgreSQL
type SystemRepo struct {
	db SQLDatabase
}

// NewSystemRepo creates a new PostgreSQL system repository
func NewSystemRepo(db SQLDatabase) *SystemRepo {
	return &SystemRepo{
		db: db,
	}
}

// StoreClient stores a client in the database
func (r *SystemRepo) StoreClient(client *domain.Client) error {
	query := `
		INSERT INTO clients (
			id, name, active, admin, created_at, updated_at, archived_at, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = $2,
			active = $3,
			admin = $4,
			updated_at = $6,
			archived_at = $7,
			email = $8
	`

	_, err := r.db.Exec(ctx, query,
		client.Id, client.Name, client.Active, client.Admin,
		client.CreatedAt, client.UpdatedAt, client.ArchivedAt, client.Email)

	return err
}

// GetClient retrieves a client by ID
func (r *SystemRepo) GetClient(id string) (*domain.Client, error) {
	query := `
		SELECT id, name, active, admin, email, created_at, updated_at, archived_at
		FROM clients
		WHERE id = $1
	`

	var client domain.Client
	err := r.db.Get(ctx, &client, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &client, nil
}

// GetClients retrieves clients with pagination
func (r *SystemRepo) GetClients(pagination *domain.Pagination[*domain.Client], filter *domain.Filter) ([]*domain.Client, error) {
	// Validate pagination parameters
	if err := pagination.Validate(); err != nil {
		return nil, err
	}

	var clients []*domain.Client
	offset := (pagination.Page - 1) * pagination.Count

	var notNil bool
	where := filter.Mask(
		domain.Filter{
			Archived: &notNil,
			State:    nil,
		},
	).ToWhere()

	if len(where) > 0 {
		where = "WHERE " + where
	}

	query := fmt.Sprintf(`
		SELECT id, name, active, admin, email, created_at, updated_at, archived_at
		FROM clients
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, where)

	err := r.db.Select(ctx, &clients, query, pagination.Count, offset)
	if err != nil {
		return nil, err
	}

	// Get total count for pagination
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM clients`
	err = r.db.Get(ctx, &totalCount, countQuery)
	if err != nil {
		return nil, err
	}

	pagination.Total = totalCount
	pagination.Found = len(clients)

	return clients, nil
}

// UpdateClient updates a client in the database
func (r *SystemRepo) UpdateClient(client *domain.Client) error {
	query := `
		UPDATE clients
		SET name = $2, active = $3, admin = $4, updated_at = $5, archived_at = $6, email = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		client.Id, client.Name, client.Active, client.Admin,
		client.UpdatedAt, client.ArchivedAt, client.Email)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveClient archives a client by ID
func (r *SystemRepo) ArchiveClient(id string) error {
	// First get the client to ensure it exists
	client, err := r.GetClient(id)
	if err != nil {
		return err
	}

	// Archive the client
	client.Archive()
	return r.UpdateClient(client)
}

// StoreClientSecret stores a client secret in the database
func (r *SystemRepo) StoreClientSecret(secret *domain.ClientSecret) error {
	return r.db.WithTx(context.Background(), func(tx SQLDatabase) error {
		// Archive existing
		query := `
			UPDATE client_secret_tokens 
			SET updated_at = NOW(), archived_at = NOW()
			WHERE client_id = $1 AND archived_at IS NULL;
		`
		_, err := tx.Exec(ctx, query, secret.ClientId)
		if err != nil {
			return err
		}

		// Create new
		query = `
			INSERT INTO client_secret_tokens (
				id, client_id, encrypted_secret, salt, state, created_at, updated_at, archived_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				encrypted_secret = $3,
				salt = $4,
				state = $5,
				updated_at = $7,
				archived_at = $8
		`
		_, err = tx.Exec(ctx, query,
			secret.Id, secret.ClientId, secret.EncryptedSecret, secret.Salt, secret.State,
			secret.CreatedAt, secret.UpdatedAt, secret.ArchivedAt)
		if err != nil {
			return err
		}
		return nil
	})
}

// GetClientSecret retrieves a client secret by ID
func (r *SystemRepo) GetClientSecret(id string) (*domain.ClientSecret, error) {
	query := `
		SELECT id, client_id, encrypted_secret, state, created_at, updated_at, archived_at
		FROM client_secret_tokens
		WHERE id = $1
	`

	var secret domain.ClientSecret
	err := r.db.Get(ctx, &secret, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &secret, nil
}

// GetClientSecretByClientId retrieves a client secret by client ID
func (r *SystemRepo) GetClientSecretByClientId(clientId string) (*domain.ClientSecret, error) {
	query := `
		SELECT id, client_id, encrypted_secret, state, created_at, updated_at, archived_at
		FROM client_secret_tokens
		WHERE client_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var secret domain.ClientSecret
	err := r.db.Get(ctx, &secret, query, clientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &secret, nil
}

// UpdateClientSecret updates a client secret in the database
func (r *SystemRepo) UpdateClientSecret(secret *domain.ClientSecret) error {
	query := `
		UPDATE client_secret_tokens
		SET encrypted_secret = $2, salt = $3, state = $4, updated_at = $5, archived_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		secret.Id, secret.EncryptedSecret, secret.Salt, secret.State,
		secret.UpdatedAt, secret.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveClientSecret archives a client secret by ID
func (r *SystemRepo) ArchiveClientSecret(id string) error {
	// First get the client secret to ensure it exists
	secret, err := r.GetClientSecret(id)
	if err != nil {
		return err
	}

	// Archive the client secret
	secret.Archive()
	return r.UpdateClientSecret(secret)
}

// StoreClientPermission stores a client permission in the database
func (r *SystemRepo) StoreClientPermission(perm *domain.ClientPermission) error {
	query := `
		INSERT INTO client_permissions (
			id, client_id, allow_read, allow_write, allow_create, created_at, updated_at, archived_at, tier
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			allow_read = $3,
			allow_write = $4,
			allow_create = $5,
			updated_at = $7,
			archived_at = $8,
			tier = $9
	`

	_, err := r.db.Exec(ctx, query,
		perm.Id, perm.ClientId, perm.Read, perm.Write, perm.Create,
		perm.CreatedAt, perm.UpdatedAt, perm.ArchivedAt, perm.Tier)

	return err
}

// GetClientPermission retrieves a client permission by ID
func (r *SystemRepo) GetClientPermission(id string) (*domain.ClientPermission, error) {
	query := `
		SELECT id, client_id, allow_read, allow_write, allow_create, tier, created_at, updated_at, archived_at
		FROM client_permissions
		WHERE id = $1
	`

	var perm domain.ClientPermission
	err := r.db.Get(ctx, &perm, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &perm, nil
}

// GetClientPermissionByClientId retrieves a client permission by client ID
func (r *SystemRepo) GetClientPermissionByClientId(clientId string) (*domain.ClientPermission, error) {
	query := `
		SELECT id, client_id, allow_read, allow_write, allow_create, tier, created_at, updated_at, archived_at
		FROM client_permissions
		WHERE client_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var perm domain.ClientPermission
	err := r.db.Get(ctx, &perm, query, clientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &perm, nil
}

// UpdateClientPermission updates a client permission in the database
func (r *SystemRepo) UpdateClientPermission(perm *domain.ClientPermission) error {
	query := `
		UPDATE client_permissions
		SET allow_read = $2, allow_write = $3, allow_create = $4, updated_at = $5, archived_at = $6, tier = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		perm.Id, perm.Read, perm.Write, perm.Create, perm.UpdatedAt, perm.ArchivedAt, perm.Tier)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveClientPermission archives a client permission by ID
func (r *SystemRepo) ArchiveClientPermission(id string) error {
	// First get the client permission to ensure it exists
	perm, err := r.GetClientPermission(id)
	if err != nil {
		return err
	}

	// Archive the client permission
	perm.Archive()
	return r.UpdateClientPermission(perm)
}

// StoreGroup stores a group in the database
func (r *SystemRepo) StoreGroup(group *domain.Group) error {
	query := `
		INSERT INTO groups (
			id, parent_group_id, owner_id, portal_id, name, state, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			parent_group_id = $2,
			owner_id = $3,
			portal_id = $4,
			name = $5,
			state = $6,
			updated_at = $8,
			archived_at = $9
	`

	_, err := r.db.Exec(ctx, query,
		group.Id, group.ParentGroupId, group.OwnerId, group.PortalId, group.Name, group.State,
		group.CreatedAt, group.UpdatedAt, group.ArchivedAt)

	return err
}

// GetGroup retrieves a group by ID
func (r *SystemRepo) GetGroup(id string) (*domain.Group, error) {
	query := `
		SELECT id, parent_group_id, owner_id, portal_id, name, state, created_at, updated_at, archived_at
		FROM groups
		WHERE id = $1
	`

	var group domain.Group
	err := r.db.Get(ctx, &group, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &group, nil
}

// UpdateGroup updates a group in the database
func (r *SystemRepo) UpdateGroup(group *domain.Group) error {
	query := `
		UPDATE groups
		SET parent_group_id = $2, owner_id = $3, portal_id = $4, name = $5, state = $6, updated_at = $7, archived_at = $8
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		group.Id, group.ParentGroupId, group.OwnerId, group.PortalId, group.Name, group.State,
		group.UpdatedAt, group.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveGroup archives a group by ID
func (r *SystemRepo) ArchiveGroup(id string) error {
	// First get the group to ensure it exists
	group, err := r.GetGroup(id)
	if err != nil {
		return err
	}

	// Archive the group
	group.Archive()
	return r.UpdateGroup(group)
}

// StoreClientGroup stores a client-group relationship in the database
func (r *SystemRepo) StoreClientGroup(clientGroup *domain.ClientGroup) error {
	query := `
		INSERT INTO client_groups (
			id, client_id, group_id, state, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			client_id = $2,
			group_id = $3,
			state = $4,
			updated_at = $6,
			archived_at = $7
	`

	_, err := r.db.Exec(ctx, query,
		clientGroup.Id, clientGroup.ClientId, clientGroup.GroupId, clientGroup.State,
		clientGroup.CreatedAt, clientGroup.UpdatedAt, clientGroup.ArchivedAt)

	return err
}

// GetClientGroupsByClientId retrieves client-group relationships by client ID
func (r *SystemRepo) GetClientGroupsByClientId(clientId string) ([]*domain.ClientGroup, error) {
	query := `
		SELECT id, client_id, group_id, state, created_at, updated_at, archived_at
		FROM client_groups
		WHERE client_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
	`

	var clientGroups []*domain.ClientGroup
	err := r.db.Select(ctx, &clientGroups, query, clientId)
	if err != nil {
		return nil, err
	}

	return clientGroups, nil
}

// GetClientGroup retrieves a client-group relationship by ID
func (r *SystemRepo) GetClientGroup(id string) (*domain.ClientGroup, error) {
	query := `
		SELECT id, client_id, group_id, state, created_at, updated_at, archived_at
		FROM client_groups
		WHERE id = $1
	`

	var clientGroup domain.ClientGroup
	err := r.db.Get(ctx, &clientGroup, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &clientGroup, nil
}

// UpdateClientGroup updates a client-group relationship in the database
func (r *SystemRepo) UpdateClientGroup(clientGroup *domain.ClientGroup) error {
	query := `
		UPDATE client_groups
		SET client_id = $2, group_id = $3, state = $4, updated_at = $5, archived_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		clientGroup.Id, clientGroup.ClientId, clientGroup.GroupId, clientGroup.State,
		clientGroup.UpdatedAt, clientGroup.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchiveClientGroup archives a client-group relationship by ID
func (r *SystemRepo) ArchiveClientGroup(id string) error {
	// First get the client group to ensure it exists
	clientGroup, err := r.GetClientGroup(id)
	if err != nil {
		return err
	}

	// Archive the client group
	clientGroup.Archive()
	return r.UpdateClientGroup(clientGroup)
}

// StorePortal stores a portal in the database
func (r *SystemRepo) StorePortal(portal *domain.Portal) error {
	query := `
		INSERT INTO portals (
			id, name, owner_id, active, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = $2,
			owner_id = $3,
			active = $4,
			updated_at = $6,
			archived_at = $7
	`

	_, err := r.db.Exec(ctx, query,
		portal.Id, portal.Name, portal.OwnerId, portal.Active,
		portal.CreatedAt, portal.UpdatedAt, portal.ArchivedAt)

	return err
}

// GetPortal retrieves a portal by ID
func (r *SystemRepo) GetPortal(id string) (*domain.Portal, error) {
	query := `
		SELECT id, name, owner_id, active, created_at, updated_at, archived_at
		FROM portals
		WHERE id = $1
	`

	var portal domain.Portal
	err := r.db.Get(ctx, &portal, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &portal, nil
}

// UpdatePortal updates a portal in the database
func (r *SystemRepo) UpdatePortal(portal *domain.Portal) error {
	query := `
		UPDATE portals
		SET name = $2, owner_id = $3, active = $4, updated_at = $5, archived_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		portal.Id, portal.Name, portal.OwnerId, portal.Active,
		portal.UpdatedAt, portal.ArchivedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ArchivePortal archives a portal by ID
func (r *SystemRepo) ArchivePortal(id string) error {
	// First get the portal to ensure it exists
	portal, err := r.GetPortal(id)
	if err != nil {
		return err
	}

	// Archive the portal
	portal.Archive()
	return r.UpdatePortal(portal)
}

// GetClientAuthByToken retrieves client authentication information by token
func (r *SystemRepo) GetClientAuthByToken(providedToken domain.Token) (*domain.ClientGroup, *domain.ClientPermission, bool, bool, error) {
	// Find the client secret by identity
	query := `
		SELECT cs.id, cs.client_id, cs.encrypted_secret, cs.salt, cs.state, cs.created_at, cs.updated_at, cs.archived_at
		FROM client_secret_tokens cs
		WHERE cs.client_id = $1 AND cs.state = 'active' AND cs.archived_at IS NULL
		LIMIT 1
	`

	var clientSecret domain.ClientSecret
	err := r.db.Get(ctx, &clientSecret, query, providedToken.Identity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, false, fmt.Errorf("invalid token")
		}
		return nil, nil, false, false, fmt.Errorf("error finding token: %w", err)
	}

	var expired bool
	haveToken := domain.TokenFromSecret(clientSecret)
	err = haveToken.Validate(providedToken.Raw)
	if err != nil {
		if errors.Is(err, domain.ErrTokenInvalid) {
			return nil, nil, false, false, fmt.Errorf("invalid token")
		} else if errors.Is(err, domain.ErrTokenExpired) {
			expired = true
		} else {
			return nil, nil, false, false, err
		}
	}

	// Get client information to check if the client is active
	var client domain.Client
	clientQuery := `
		SELECT id, name, active, admin, email, created_at, updated_at, archived_at
		FROM clients
		WHERE id = $1
	`
	err = r.db.Get(ctx, &client, clientQuery, clientSecret.ClientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, false, fmt.Errorf("client not found")
		}
		return nil, nil, false, false, fmt.Errorf("error finding client: %w", err)
	}

	if !client.Active || client.ArchivedAt != nil {
		return nil, nil, false, false, fmt.Errorf("client is not active")
	}

	// Get client permission
	var clientPermission domain.ClientPermission
	permQuery := `
		SELECT id, client_id, allow_read, allow_write, allow_create, tier, created_at, updated_at, archived_at
		FROM client_permissions
		WHERE client_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	err = r.db.Get(ctx, &clientPermission, permQuery, clientSecret.ClientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, false, fmt.Errorf("client permission not found")
		}
		return nil, nil, false, false, fmt.Errorf("error finding client permission: %w", err)
	}

	// Get the primary client group
	var clientGroup domain.ClientGroup
	groupQuery := `
		SELECT cg.id, cg.client_id, cg.group_id, cg.state, cg.created_at, cg.updated_at, cg.archived_at
		FROM client_groups cg
		WHERE cg.client_id = $1 AND cg.state = 'active' AND cg.archived_at IS NULL
		ORDER BY cg.created_at DESC
		LIMIT 1
	`
	err = r.db.Get(ctx, &clientGroup, groupQuery, clientSecret.ClientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, false, fmt.Errorf("client group not found")
		}
		return nil, nil, false, false, fmt.Errorf("error finding client group: %w", err)
	}

	return &clientGroup, &clientPermission, client.Admin, expired, nil
}

// AggregateClientMessageUsage retrieves message usage statistics for clients during a time period
func (r *SystemRepo) AggregateClientMessageUsageByClient(start, stop time.Time, clientId string) (domain.ClientUsage, error) {
	query := `
		SELECT 
		client_id,
		COUNT(*) FILTER (WHERE role = 'sender') AS pub_count,
		COUNT(*) FILTER (WHERE role = 'receiver') AS sub_count
		FROM (
			SELECT messages.client_id AS client_id, 'sender' AS role
			FROM messages
			JOIN message_subscribers ON message_subscribers.message_id = messages.id
			WHERE messages.created_at >= $1 AND messages.created_at < $2 AND messages.client_id = $3

			UNION ALL

			SELECT message_subscribers.client_id AS client_id, 'receiver' AS role
			FROM messages
			JOIN message_subscribers ON message_subscribers.message_id = messages.id
			WHERE messages.created_at >= $1 AND messages.created_at < $2 AND message_subscribers.client_id = $3
		) AS combined
		GROUP BY client_id
		ORDER BY client_id;
	`

	var clientUsageCounts []domain.ClientUsage
	err := r.db.Select(ctx, &clientUsageCounts, query, start, stop, clientId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ClientUsage{}, ErrNotFound
		}
		return domain.ClientUsage{}, err
	} else if len(clientUsageCounts) == 0 {
		return domain.ClientUsage{}, nil
	}
	clientUsageCount := clientUsageCounts[0]

	startStr, stopStr := start.Format("20060102150405"), stop.Format("20060102150405")
	// Generate IDs and set start/stop times for the usage records
	clientUsageCount.Id = fmt.Sprintf("%s-%s-%s",
		clientUsageCount.ClientId,
		startStr,
		stopStr)
	clientUsageCount.Start = start
	clientUsageCount.Stop = stop

	return clientUsageCount, nil
}

// AggregateClientMessageUsage retrieves message usage statistics for clients during a time period
func (r *SystemRepo) AggregateClientMessageUsage(start, stop time.Time) ([]domain.ClientUsage, error) {
	query := `
		SELECT 
		client_id,
		COUNT(*) FILTER (WHERE role = 'sender') AS pub_count,
		COUNT(*) FILTER (WHERE role = 'receiver') AS sub_count
		FROM (
			SELECT messages.client_id AS client_id, 'sender' AS role
			FROM messages
			JOIN message_subscribers ON message_subscribers.message_id = messages.id
			WHERE messages.created_at >= $1 AND messages.created_at < $2

			UNION ALL

			SELECT message_subscribers.client_id AS client_id, 'receiver' AS role
			FROM messages
			JOIN message_subscribers ON message_subscribers.message_id = messages.id
			WHERE messages.created_at >= $1 AND messages.created_at < $2
		) AS combined
		GROUP BY client_id
		ORDER BY client_id;
	`

	var clientUsageCounts []domain.ClientUsage
	err := r.db.Select(ctx, &clientUsageCounts, query, start, stop)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	startStr, stopStr := start.Format("20060102150405"), stop.Format("20060102150405")
	// Generate IDs and set start/stop times for the usage records
	for i := range clientUsageCounts {
		clientUsageCounts[i].Id = fmt.Sprintf("%s-%s-%s",
			clientUsageCounts[i].ClientId,
			startStr,
			stopStr)
		clientUsageCounts[i].Start = start
		clientUsageCounts[i].Stop = stop
	}

	return clientUsageCounts, nil
}

// GetClientUsageByDay retrieves client usage statistics for a specific day
func (r *SystemRepo) GetClientUsageByDay(date time.Time, clientId string) ([]domain.ClientUsage, error) {
	// Create start and end times for the specified day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	args := []any{startOfDay, endOfDay}

	where := ""
	if clientId != "" {
		where = "AND client_id = $3"
		args = append(args, clientId)
	}
	query := fmt.Sprintf(`
		SELECT 
			id,
			client_id,
			sub_count,
			pub_count,
			start,
			stop
		FROM client_usage
		WHERE start >= $1 AND stop < $2 %s
		ORDER BY client_id, start;
	`, where)

	var clientUsage []domain.ClientUsage
	err := r.db.Select(ctx, &clientUsage, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty array instead of error for no usage data
			return []domain.ClientUsage{}, nil
		}
		return nil, err
	}

	return clientUsage, nil
}

// StoreClientUsage stores client usage records in the database
func (r *SystemRepo) StoreClientUsage(usageRecords []domain.ClientUsage) error {
	// Use a transaction to ensure all records are stored atomically
	return r.db.WithTx(ctx, func(tx SQLDatabase) error {
		for _, record := range usageRecords {
			query := `
				INSERT INTO client_usage (
					id, client_id, pub_count, sub_count, start, stop
				) VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO UPDATE SET
					pub_count = client_usage.pub_count + EXCLUDED.pub_count,
					sub_count = client_usage.sub_count + EXCLUDED.sub_count
			`

			_, err := tx.Exec(ctx, query,
				record.Id, record.ClientId, record.PubCount, record.SubCount, record.Start, record.Stop)
			if err != nil {
				return fmt.Errorf("failed to store client usage record: %w", err)
			}
		}
		return nil
	})
}

func (r *SystemRepo) StoreContactRequest(contactRequest *domain.ContactRequest) error {
	query := `
		INSERT INTO contact_requests (
			id, name, email, message, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		contactRequest.Id, contactRequest.Name, contactRequest.Email,
		contactRequest.Message, contactRequest.CreatedAt)

	return err
}

// // WithTx executes the given function within a transaction
// func (r *SystemRepo) WithTx(f func(tx RepoProvider) error) error {
// 	return r.db.WithTx(ctx, func(txDb SQLDatabase) error {
// 		txRepo := &SystemRepo{db: txDb}
// 		return f(txRepo)
// 	})
// }

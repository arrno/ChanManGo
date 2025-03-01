-- Create enum type for STATE
CREATE TYPE state_enum AS ENUM (
  'pending',
  'processed',
  'running',
  'active',
  'complete',
  'inactive',
  'archived',
  'error'
);

-- Create Portals table
CREATE TABLE portals (
  id VARCHAR(255) PRIMARY KEY,
  owner_id VARCHAR(255) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP
);

-- Create Groups table
CREATE TABLE groups (
  id VARCHAR(255) PRIMARY KEY,
  owner_id VARCHAR(255) NOT NULL,
  portal_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (portal_id) REFERENCES portals(id) ON DELETE CASCADE
);

-- Create index on portal_id for groups
CREATE INDEX idx_groups_portal_id ON groups(portal_id);

-- Create Clients table
CREATE TABLE clients (
  id VARCHAR(255) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP
);

-- Create Channels table
CREATE TABLE channels (
  id VARCHAR(255) PRIMARY KEY,
  portal_id VARCHAR(255) NOT NULL,
  group_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  archived BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (portal_id) REFERENCES portals(id) ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

-- Create indexes for channels
CREATE INDEX idx_channels_portal_id ON channels(portal_id);
CREATE INDEX idx_channels_group_id ON channels(group_id);

-- Create ClientGroup junction table for many-to-many relationship
CREATE TABLE client_groups (
  id VARCHAR(255) PRIMARY KEY,
  client_id VARCHAR(255) NOT NULL,
  group_id VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  UNIQUE (client_id, group_id)
);

-- Create indexes for client_groups
CREATE INDEX idx_client_groups_client_id ON client_groups(client_id);
CREATE INDEX idx_client_groups_group_id ON client_groups(group_id);

-- Create ChannelGroup junction table for many-to-many relationship
CREATE TABLE channel_groups (
  id VARCHAR(255) PRIMARY KEY,
  channel_id VARCHAR(255) NOT NULL,
  group_id VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  UNIQUE (channel_id, group_id)
);

-- Create indexes for channel_groups
CREATE INDEX idx_channel_groups_channel_id ON channel_groups(channel_id);
CREATE INDEX idx_channel_groups_group_id ON channel_groups(group_id);

-- Create ChannelSubscription table
CREATE TABLE channel_subscriptions (
  id VARCHAR(255) PRIMARY KEY,
  portal_id VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  archived BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (portal_id) REFERENCES portals(id) ON DELETE CASCADE
);

-- Create index for channel_subscriptions
CREATE INDEX idx_channel_subscriptions_portal_id ON channel_subscriptions(portal_id);

-- Create Workers table
CREATE TABLE workers (
  id VARCHAR(255) PRIMARY KEY,
  channel_id VARCHAR(255) NOT NULL,
  state state_enum NOT NULL,
  archived BOOLEAN NOT NULL DEFAULT false,
  interval_seconds INTEGER NOT NULL,
  heartbeat TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
);

-- Create index for workers
CREATE INDEX idx_workers_channel_id ON workers(channel_id);

-- Create TopicSubscription table
CREATE TABLE topic_subscriptions (
  id VARCHAR(255) PRIMARY KEY,
  client_id VARCHAR(255) NOT NULL,
  topics TEXT[] NOT NULL,
  state state_enum NOT NULL,
  archived BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
);

-- Create index for topic_subscriptions
CREATE INDEX idx_topic_subscriptions_client_id ON topic_subscriptions(client_id);

-- Create Messages table
CREATE TABLE messages (
  id VARCHAR(255) PRIMARY KEY,
  topic VARCHAR(255) NOT NULL,
  client_id VARCHAR(255) NOT NULL,
  processed_by_worker_id VARCHAR(255),
  processed_at TIMESTAMP,
  acknowledged_by_worker_id VARCHAR(255),
  acknowledged_at TIMESTAMP,
  state state_enum NOT NULL,
  archived BOOLEAN NOT NULL DEFAULT false,
  payload BYTEA NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
  FOREIGN KEY (processed_by_worker_id) REFERENCES workers(id) ON DELETE SET NULL,
  FOREIGN KEY (acknowledged_by_worker_id) REFERENCES workers(id) ON DELETE SET NULL
);

-- Create indexes for messages
CREATE INDEX idx_messages_client_id ON messages(client_id);
CREATE INDEX idx_messages_topic ON messages(topic);
CREATE INDEX idx_messages_state ON messages(state);
CREATE INDEX idx_messages_processed_by_worker_id ON messages(processed_by_worker_id);
CREATE INDEX idx_messages_acknowledged_by_worker_id ON messages(acknowledged_by_worker_id);
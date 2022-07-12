package postgres

// Migration 0
const EventsDDL = `
-- EventsDDL is the postgresql data definition language that defines the
-- relational schema of events.
--
-- Events are expected to be referred to by the triple of
-- ( sensu_namespace, sensu_check, sensu_entity ), and not by their primary
-- key. The primary key is intended only for future use by joins that will
-- reference the events table.
--
CREATE TABLE IF NOT EXISTS events (
	id                  bigserial     PRIMARY KEY,
	sensu_namespace     text          NOT NULL,
	sensu_entity        text          NOT NULL,
	sensu_check         text          NOT NULL,
	status              integer       NOT NULL,
	last_ok             integer       NOT NULL,
	occurrences         integer       NOT NULL,
	occurrences_wm      integer       NOT NULL,
	history_index       integer       DEFAULT 2,
	history_status      integer[],
	history_ts          integer[],
	serialized          bytea,
	previous_serialized bytea,
	UNIQUE ( sensu_namespace, sensu_check, sensu_entity )
);`

// Migration 1
const MigrateEventsID = `
ALTER TABLE IF EXISTS events ALTER COLUMN id SET DATA TYPE bigint;
ALTER SEQUENCE IF EXISTS events_id_seq AS bigint;
`

// Migration 2
const MigrateAddSortHistoryFunc = `
CREATE OR REPLACE FUNCTION sort_history(a INTEGER[], idx INTEGER)
RETURNS INTEGER[] AS $$
DECLARE
	lhs INTEGER[];
	rhs INTEGER[];
BEGIN
	rhs := a[1:idx];
	lhs := a[idx + 1:array_length(a, 1)];
	IF lhs IS NULL THEN
		return rhs;
	END IF;
	IF rhs IS NULL THEN
		return lhs;
	END IF;
	return (lhs || rhs)::INTEGER[];
END; $$
LANGUAGE plpgsql IMMUTABLE;
`

// Migration 3
const MigrateAddSelectorColumn = `
ALTER TABLE IF EXISTS events ADD COLUMN selectors jsonb;
CREATE INDEX IF NOT EXISTS idxginevents ON events USING GIN (selectors jsonb_path_ops);
`

// Migration 4
// NOTE: see migrateUpdateSelectors function

// Migration 5
const ringSchema = `
-- rings are named sets of items. The items can be iterated over in order,
-- wrapping around at the end. The entities in a ring can also exist in other
-- rings. Deleting a ring causes all of its entities to be removed from its
-- membership, via an ON DELETE CASCADE trigger.
CREATE TABLE IF NOT EXISTS rings (
	id bigserial PRIMARY KEY,
	name text UNIQUE NOT NULL
);

-- entities are sensu entities that exist within a namespace.
-- table renamed to entity_states by migration 9
-- namespace column replaced by namespace_id in migration 14
--
CREATE TABLE IF NOT EXISTS entities (
	id bigserial PRIMARY KEY,
	namespace text NOT NULL,
	name text NOT NULL,
	expires_at timestamp NOT NULL,
	UNIQUE ( namespace, name )
);

-- ring_entities are associations between rings and entities.
CREATE TABLE IF NOT EXISTS ring_entities (
	ring_id bigint NOT NULL REFERENCES rings (id) ON DELETE CASCADE,
	entity_id bigint NOT NULL REFERENCES entities (id) ON DELETE CASCADE,
	PRIMARY KEY ( ring_id, entity_id )
);

-- ring_subscribers are named subscribers to a particular ring. Independent of
-- any other subscribers, they can iterate the ring with the parameters of
-- their choosing. Unlike a fifo, a ring can be iterated concurrently by any
-- number of subscribers, without one iteration having a side effect on another.
--
-- Each subscriber contains a pointer to a ring member. If the ring member is
-- deleted, then the pointer becomes NULL. The queries that interact with
-- ring_subscribers must therefore be tolerant of NULL pointer values.
--
-- Subscribers iterate through the ring by modifying the value of their pointer.
--
-- Subscribers are automatically deleted when their rings are deleted.
CREATE TABLE IF NOT EXISTS ring_subscribers (
	ring_id bigint NOT NULL REFERENCES rings (id) ON DELETE CASCADE,
	name text NOT NULL,
	pointer bigint REFERENCES entities ( id ) ON DELETE SET NULL,
	last_updated timestamp DEFAULT '01-01-01 01:00:00' NOT NULL,
	PRIMARY KEY ( ring_id, name )
);

CREATE INDEX IF NOT EXISTS rings_name_idx
ON rings ( name );

CREATE INDEX IF NOT EXISTS entities_name_idx
ON entities ( namespace, name );

CREATE INDEX IF NOT EXISTS entities_expires_at_idx
ON entities ( expires_at );

CREATE INDEX IF NOT EXISTS ring_subscribers_last_updated_idx
ON ring_subscribers ( last_updated );
`

// Migration 6
// NOTE: see fixMissingStateSelector function

// Migration 7
const migrateEntityState = `
-- This is a migration that expands the entities table, and creates
-- a new table that joins to it on entity_id called entities_systems.
--
ALTER TABLE entities ADD COLUMN last_seen bigint DEFAULT 0;
ALTER TABLE entities ADD COLUMN selectors jsonb;
ALTER TABLE entities ADD COLUMN annotations jsonb;

-- entities_systems contains agent system information.
-- It's mostly composed of informational text values.
--
CREATE TABLE entities_systems (
	entity_id           bigint UNIQUE REFERENCES entities (id) ON DELETE CASCADE,
	hostname            text,
	os                  text,
	platform            text,
	platform_family     text,
	platform_version    text,
	arch                text,
	arm_version         integer,
	libc_type           text,
	vm_system           text,
	vm_role             text,
	cloud_provider      text,
	float_type          text,
	sensu_agent_version text
);

-- This function is used to turn a 2D array into a 1D array. UNNEST will
-- produce scalars, no matter how many input dimensions. Taken from
-- https://wiki.postgresql.org/wiki/Array_Unnest
--
CREATE OR REPLACE FUNCTION unnest_2d_1d(anyarray)
	RETURNS SETOF anyarray AS
$func$
SELECT array_agg($1[d1][d2])
FROM   generate_series(array_lower($1,1), array_upper($1,1)) d1
	 , generate_series(array_lower($1,2), array_upper($1,2)) d2
GROUP  BY d1
ORDER  BY d1
$func$
LANGUAGE sql IMMUTABLE;

-- entities_networks are network adapters found in the agent system.
-- entities can have many network adapters.
--
CREATE TABLE IF NOT EXISTS entities_networks (
	entity_id bigint REFERENCES entities (id) ON DELETE CASCADE,
	name      text NOT NULL,
	mac       text NOT NULL,
	addresses jsonb,
	UNIQUE(entity_id, name, mac, addresses)
);

-- Index the entity_id colum on entities_networks
--
CREATE INDEX ON entities_networks ( entity_id );
`

// Migration 8
const migrateRefreshUpdatedAtProcedure = `
-- This is a procedure that can be used to set the "updated_at" column to the
-- current time on a given table. It is intended to be used with update triggers
-- to ensure that the "updated_at" column represents the last time a row was
-- updated.
CREATE OR REPLACE FUNCTION refresh_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = now();
	RETURN NEW;
END;
$$ language 'plpgsql';
`

// Migration 9
const migrateRenameEntitiesTable = `
-- This is a migration that renames the entities table to entity_states.
--
ALTER TABLE entities RENAME TO entity_states;
`

// Migration 10
const migrateRenameEntityStateUniqueConstraint = `
-- This is a migration that renames the uniqueness constraint to
-- entity_state_unique for the entity_states table.
ALTER TABLE entity_states RENAME CONSTRAINT entities_namespace_name_key TO entity_state_unique;
`

// Migration 11
const entityConfigSchema = `
-- namespace column replaced by namespace_id in migration 14
--
CREATE TABLE IF NOT EXISTS entity_configs (
	id                 bigserial PRIMARY KEY,
	namespace          text NOT NULL,
	name               text NOT NULL,
	selectors          jsonb,
	annotations        jsonb,
	created_by         text NOT NULL,
	entity_class       text NOT NULL,
	sensu_user         text,
	subscriptions      text[],
	deregister         boolean,
	deregistration     text,
	keepalive_handlers text[],
	redact             text[],
	created_at         timestamptz NOT NULL DEFAULT NOW(),
	updated_at         timestamptz NOT NULL DEFAULT NOW(),
	deleted_at         timestamptz,
	CONSTRAINT entity_config_unique UNIQUE (namespace, name)
);

CREATE TRIGGER refresh_entity_configs_updated_at BEFORE UPDATE
	ON entity_configs FOR EACH ROW EXECUTE PROCEDURE
	refresh_updated_at_column();
`

// Migration 12
const migrateAddEntityConfigIdToEntityState = `
-- This is a migration that adds an entity config foreign key to the
-- entity_states table.
--
ALTER TABLE entity_states
ADD COLUMN entity_config_id bigint NOT NULL REFERENCES entity_configs (id) ON DELETE CASCADE;
`

// Migration 13
const namespaceSchema = `
CREATE TABLE IF NOT EXISTS namespaces (
	id                 bigserial PRIMARY KEY,
	name               text NOT NULL,
	selectors          jsonb,
	annotations        jsonb,
	created_at         timestamptz NOT NULL DEFAULT NOW(),
	updated_at         timestamptz NOT NULL DEFAULT NOW(),
	deleted_at         timestamptz,
	CONSTRAINT namespace_unique UNIQUE (name)
);

CREATE TRIGGER refresh_namespaces_updated_at BEFORE UPDATE
	ON namespaces FOR EACH ROW EXECUTE PROCEDURE
	refresh_updated_at_column();
`

// Migration 14
const addNamespaceForeignKeys = `
-- Add unique namespaces across existing tables to the new namespaces table
INSERT INTO namespaces (name)
SELECT namespace FROM entity_configs WHERE namespace IS NOT NULL
UNION
SELECT namespace FROM entity_states WHERE namespace IS NOT NULL;

-- Add namespace_id column to existing tables.
ALTER TABLE entity_configs ADD COLUMN namespace_id bigint;
ALTER TABLE entity_states ADD COLUMN namespace_id bigint;

-- Set namespace_id values in the entity_configs table
UPDATE entity_configs
SET namespace_id = namespaces.id
FROM namespaces
WHERE namespaces.name = entity_configs.namespace;

-- Set namespace_id values in the entity_states table
UPDATE entity_states
SET namespace_id = namespaces.id
FROM namespaces
WHERE namespaces.name = entity_states.namespace;

-- Add foreign key constraint to entity_configs table
ALTER TABLE entity_configs
ADD FOREIGN KEY (namespace_id) REFERENCES namespaces (id)
ON DELETE CASCADE;

-- Add foreign key constraint to entity_states table
ALTER TABLE entity_states
ADD FOREIGN KEY (namespace_id) REFERENCES namespaces (id)
ON DELETE CASCADE;

-- Update unique constraint for entity_configs table
ALTER TABLE entity_configs
DROP CONSTRAINT entity_config_unique;
ALTER TABLE entity_configs
ADD CONSTRAINT entity_config_unique UNIQUE (namespace_id, name);

-- Update unique constraint for entity_states table
ALTER TABLE entity_states
DROP CONSTRAINT entity_state_unique;
ALTER TABLE entity_states
ADD CONSTRAINT entity_state_unique UNIQUE (namespace_id, name);

-- Update index for entity_states table
CREATE INDEX entity_states_name_idx ON entity_states ( namespace_id, name );
DROP INDEX entities_name_idx;

-- Drop old namespace columns
ALTER TABLE entity_configs DROP COLUMN namespace;
ALTER TABLE entity_states DROP COLUMN namespace;
`

// Migration 15
const addTimestampColumns = `
ALTER TABLE entity_states ADD COLUMN created_at timestamptz NOT NULL DEFAULT NOW();
ALTER TABLE entity_states ADD COLUMN updated_at timestamptz NOT NULL DEFAULT NOW();
ALTER TABLE entity_states ADD COLUMN deleted_at timestamptz;
`

// Migration 16
const addInitializedTable = `
CREATE TABLE IF NOT EXISTS initialized (
	initialized bool PRIMARY KEY,
	created_at         timestamptz NOT NULL DEFAULT NOW(),
	updated_at         timestamptz NOT NULL DEFAULT NOW(),
	CONSTRAINT initialized_unique CHECK (initialized)
);
`

package postgres

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

const migrateRenameEntitiesTable = `
-- This is a migration that renames the entities table to entity_states.
--
ALTER TABLE entities RENAME TO entity_states;
`

const migrateAddEntityConfigIdToEntityState = `
-- This is a migration that adds an entity config foreign key to the
-- entity_states table.
--
ALTER TABLE entity_states
ADD COLUMN entity_config_id bigint NOT NULL REFERENCES entity_configs (id) ON DELETE CASCADE;
`

const migrateRenameEntityStateUniqueConstraint = `
-- This is a migration that renames the uniqueness constraint to
-- entity_state_unique for the entity_states table.
ALTER TABLE entity_states RENAME CONSTRAINT entities_namespace_name_key TO entity_state_unique;
`

// see ring_schema.go
const createOrUpdateEntityStateQuery = `
-- This query creates a new entity state, or updates it if it already exists.
--
-- The query has several WITH clauses, that allow it to affect multiple tables
-- in one shot. This allows us to use prepared statements, as they otherwise
-- won't work with multiple commands.
--
-- Parameters:
-- $1: The entity namespace.
-- $2: The entity name.
-- $3: The time the entity was last seen by the backend.
-- $4: The label selectors of the entity state.
-- $5: The annotations of the entity state.
-- $6-$17: System facts.
-- $18: The sensu agent version.
-- $19: Network names, an array of strings.
-- $20: Network MAC addresses, an array of strings.
-- $21: Network names, an array of json-encoded arrays.
--
WITH config AS (
	SELECT id FROM entity_configs
	WHERE entity_configs.namespace = $1 AND entity_configs.name = $2
), state AS (
	INSERT INTO entity_states ( entity_config_id, namespace, name, expires_at, last_seen, selectors, annotations )
	VALUES ( (SELECT id FROM config), $1, $2, now(), $3, $4, $5 )
	ON CONFLICT ( namespace, name )
	DO UPDATE
	SET last_seen = $3, selectors = $4, annotations = $5
	RETURNING id AS id
), system AS (
	SELECT
	state.id AS entity_id,
	$6::text AS hostname,
	$7::text AS os,
	$8::text AS platform,
	$9::text AS platform_family,
	$10::text AS platform_version,
	$11::text AS arch,
	$12::integer AS arm_version, -- TODO(eric): is this a bug? Doesn't work without the cast.
	$13::text AS libc_type,
	$14::text AS vm_system,
	$15::text AS vm_role,
	$16::text AS cloud_provider,
	$17::text AS float_type,
	$18::text AS sensu_agent_version
	FROM state
), networks AS (
	SELECT
	nics.name AS name,
	nics.mac AS mac,
	nics.addresses AS addresses
	FROM UNNEST(
		cast($19 AS text[]),
		cast($20 AS text[]),
		cast($21 AS jsonb[])
	) AS nics ( name, mac, addresses )
),
sys_update AS (
	-- Insert or update the entity's system object.
	INSERT INTO entities_systems SELECT * FROM system
	ON CONFLICT (entity_id) DO UPDATE SET (
		hostname,
		os,
		platform,
		platform_family,
		platform_version,
		arch,
		arm_version,
		libc_type,
		vm_system,
		vm_role,
		cloud_provider,
		float_type,
		sensu_agent_version
	) = (
	-- Update the system if it exists already.
		SELECT
			hostname,
			os,
			platform,
			platform_family,
			platform_version,
			arch,
			arm_version,
			libc_type,
			vm_system,
			vm_role,
			cloud_provider,
			float_type,
			sensu_agent_version
		FROM system
	)
), del AS (
	-- Delete the networks that are not in the current set.
	DELETE FROM entities_networks
	USING state, networks
	WHERE entity_id = state.id
	AND ( entities_networks.name, entities_networks.mac, entities_networks.addresses ) NOT IN (SELECT * FROM networks)
)
-- Insert the networks that are new in the current set.
INSERT INTO entities_networks
SELECT state.id, networks.name, networks.mac, networks.addresses
FROM state, networks
ON CONFLICT ( entity_id, name, mac, addresses ) DO NOTHING
`

const createIfNotExistsEntityStateQuery = `
-- This query inserts rows into the entity_states table. By design, it
-- errors when an entity with the same namespace and name already
-- exists.
--
WITH config AS (
	SELECT id
	FROM entity_configs
	WHERE namespace = $1 AND name = $2
), state AS (
	INSERT INTO entity_states ( entity_config_id, namespace, name, expires_at, last_seen, selectors, annotations )
	VALUES ( SELECT config.id, $1, $2, now(), $3, $4, $5 )
	RETURNING id
), system AS (
	INSERT INTO entities_systems
	SELECT
	state.id,
	$6::text,
	$7::text,
	$8::text,
	$9::text,
	$10::text,
	$11::text,
	$12::integer, -- TODO(eric): is this a bug? Doesn't work without the cast.
	$13::text,
	$14::text,
	$15::text,
	$16::text,
	$17::text,
	$18::text
	FROM state
)
INSERT INTO entities_networks
SELECT
state.id,
nics.name,
nics.mac,
nics.addresses
FROM state, UNNEST(
	cast($19 AS text[]),
	cast($20 AS text[]),
	cast($21 AS jsonb[])
)
AS nics ( name, mac, addresses )
`

const updateIfExistsEntityStateQuery = `
-- This query updates the entity state, but only if it exists.
--
WITH state AS (
	SELECT id FROM entity_states
	WHERE entity_states.namespace = $1 AND entity_states.name = $2
), upd AS (
	UPDATE entity_states
	SET last_seen = $3, selectors = $4, annotations = $5
	FROM state
	WHERE state.id = entity_states.id
), system AS (
	SELECT
	state.id AS entity_id,
	$6::text AS hostname,
	$7::text AS os,
	$8::text AS platform,
	$9::text AS platform_family,
	$10::text AS platform_version,
	$11::text AS arch,
	$12::integer AS arm_version, -- TODO(eric): is this a bug? Doesn't work without the cast.
	$13::text AS libc_type,
	$14::text AS vm_system,
	$15::text AS vm_role,
	$16::text AS cloud_provider,
	$17::text AS float_type,
	$18::text AS sensu_agent_version
	FROM state
), networks AS (
	SELECT
	nics.name AS name,
	nics.mac AS mac,
	nics.addresses AS addresses
	FROM UNNEST(
		cast($19 AS text[]),
		cast($20 AS text[]),
		cast($21 AS jsonb[])
	)
	AS nics ( name, mac, addresses )
), sys_update AS (
	-- Insert or update the entity's system object.
	INSERT INTO entities_systems SELECT * FROM system
	ON CONFLICT (entity_id) DO UPDATE SET (
		hostname,
		os,
		platform,
		platform_family,
		platform_version,
		arch,
		arm_version,
		libc_type,
		vm_system,
		vm_role,
		cloud_provider,
		float_type,
		sensu_agent_version
	) = (
	-- Update the system if it exists already.
		SELECT
			hostname,
			os,
			platform,
			platform_family,
			platform_version,
			arch,
			arm_version,
			libc_type,
			vm_system,
			vm_role,
			cloud_provider,
			float_type,
			sensu_agent_version
		FROM system
	)
), del AS (
	-- Delete the networks that are not in the current set.
	DELETE FROM entities_networks
	USING state, networks
	WHERE entity_id = state.id
	AND ( entities_networks.name, entities_networks.mac, entities_networks.addresses ) NOT IN (SELECT * FROM networks)
), ins AS (
	-- Insert the networks that are new in the current set.
	INSERT INTO entities_networks
	SELECT state.id, networks.name, networks.mac, networks.addresses
	FROM state, networks
	ON CONFLICT ( entity_id, name, mac, addresses ) DO NOTHING
)
SELECT * FROM state;
`

const getEntityStateQuery = `
-- This query fetches a single entity state, or nothing.
--
WITH state AS (
	SELECT entity_states.id AS id
	FROM entity_states
	WHERE entity_states.namespace = $1 AND entity_states.name = $2
), network AS (
	SELECT
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, state
	WHERE entities_networks.entity_id = state.id
)
SELECT 
    entity_states.namespace,
    entity_states.name,
	entity_states.last_seen,
	entity_states.selectors,
	entity_states.annotations,
	entities_systems.hostname,
	entities_systems.os,
	entities_systems.platform,
	entities_systems.platform_family,
	entities_systems.platform_version,
	entities_systems.arch,
	entities_systems.arm_version,
	entities_systems.libc_type,
	entities_systems.vm_system,
	entities_systems.vm_role,
	entities_systems.cloud_provider,
	entities_systems.float_type,
	entities_systems.sensu_agent_version,
	network.names,
	network.macs,
	network.addresses::text[]
FROM entity_states LEFT OUTER JOIN entities_systems ON entity_states.id = entities_systems.entity_id, network, state
WHERE entity_states.id = state.id
`

const getEntityStatesQuery = `
-- This query fetches multiple entity states, including network information.
--
WITH state AS (
	SELECT entity_states.id AS id
	FROM entity_states
	WHERE entity_states.namespace = $1 AND entity_states.name IN (SELECT unnest($2::text[]))
), network AS (
	SELECT
		entities_networks.entity_id AS entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, state
	WHERE entities_networks.entity_id = state.id
	GROUP BY entities_networks.entity_id
)
SELECT
    entity_states.namespace,
    entity_states.name,
	entity_states.last_seen,
	entity_states.selectors,
	entity_states.annotations,
	entities_systems.hostname,
	entities_systems.os,
	entities_systems.platform,
	entities_systems.platform_family,
	entities_systems.platform_version,
	entities_systems.arch,
	entities_systems.arm_version,
	entities_systems.libc_type,
	entities_systems.vm_system,
	entities_systems.vm_role,
	entities_systems.cloud_provider,
	entities_systems.float_type,
	entities_systems.sensu_agent_version,
	network.names,
	network.macs,
	network.addresses::text[]
FROM entity_states LEFT OUTER JOIN entities_systems ON entity_states.id = entities_systems.entity_id, network, state
WHERE entity_states.id = state.id AND entity_states.id = network.entity_id
`

const deleteEntityStateQuery = `
-- This query deletes an entity state. Any related system & network state will
-- also be deleted via ON DELETE CASCADE triggers.
--
-- Parameters:
-- $1 Namespace
-- $2 Entity name
DELETE FROM entity_states WHERE entity_states.namespace = $1 AND entity_states.name = $2;
`

const listEntityStateQuery = `
-- This query lists entity states from a given namespace.
--
with network AS (
	SELECT
		entity_states.id as entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entity_states
	WHERE entity_states.namespace = $1 OR $1 IS NULL
	GROUP BY entities.id
)
SELECT 
    entity_states.namespace,
    entity_states.name,
	entity_states.last_seen,
	entity_states.selectors,
	entity_states.annotations,
	entities_systems.hostname,
	entities_systems.os,
	entities_systems.platform,
	entities_systems.platform_family,
	entities_systems.platform_version,
	entities_systems.arch,
	entities_systems.arm_version,
	entities_systems.libc_type,
	entities_systems.vm_system,
	entities_systems.vm_role,
	entities_systems.cloud_provider,
	entities_systems.float_type,
	entities_systems.sensu_agent_version,
	network.names,
	network.macs,
	network.addresses::text[]
FROM entity_states LEFT OUTER JOIN entities_systems ON entity_states.id = entities_systems.entity_id, network
WHERE network.entity_id = entity_states.id
ORDER BY ( entity_states.namespace, entity_states.name ) ASC
LIMIT $2
OFFSET $3
`

const listEntityStateDescQuery = `
-- This query lists entities from a given namespace.
--
with network AS (
	SELECT
		entity_states.id as entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entity_states
	WHERE entity_states.namespace = $1 OR $1 IS NULL
	GROUP BY entity_states.id
)
SELECT 
    entity_states.namespace,
    entity_states.name,
	entity_states.last_seen,
	entity_states.selectors,
	entity_states.annotations,
	entities_systems.hostname,
	entities_systems.os,
	entities_systems.platform,
	entities_systems.platform_family,
	entities_systems.platform_version,
	entities_systems.arch,
	entities_systems.arm_version,
	entities_systems.libc_type,
	entities_systems.vm_system,
	entities_systems.vm_role,
	entities_systems.cloud_provider,
	entities_systems.float_type,
	entities_systems.sensu_agent_version,
	network.names,
	network.macs,
	network.addresses::text[]
FROM entity_states LEFT OUTER JOIN entities_systems ON entity_states.id = entities_systems.entity_id, network
WHERE network.entity_id = entity_states.id
ORDER BY ( entity_states.namespace, entity_states.name ) DESC
LIMIT $2
OFFSET $3
`

const existsEntityStateQuery = `
-- This query discovers if an entity exists, without retrieving it.
--
SELECT true FROM entity_states
WHERE entity_states.namespace = $1 AND entity_states.name = $2;
`

const patchEntityStateQuery = `
-- This query updates only the last_seen and expires_at values of an
-- entity.
--
UPDATE entity_states
SET expires_at = $3 + now(), last_seen = $4
WHERE entity_states.namespace = $1 AND entity_states.name = $2;
`

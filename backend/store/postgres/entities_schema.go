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

// see ring_schema.go
const createOrUpdateEntityStateQuery = `
-- This query creates a new entity, or updates it if it already exists.
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
WITH entity AS (
	INSERT INTO entities ( namespace, name, expires_at, last_seen, selectors, annotations )
	VALUES ( $1, $2, now(), $3, $4, $5 )
	ON CONFLICT ( namespace, name )
	DO UPDATE
	SET last_seen = $3, selectors = $4, annotations = $5
	RETURNING id AS id
), system AS (
	SELECT
	entity.id AS entity_id,
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
	FROM entity
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
	USING entity, networks
	WHERE entity_id = entity.id
	AND ( entities_networks.name, entities_networks.mac, entities_networks.addresses ) NOT IN (SELECT * FROM networks)
)
-- Insert the networks that are new in the current set.
INSERT INTO entities_networks
SELECT entity.id, networks.name, networks.mac, networks.addresses
FROM entity, networks
ON CONFLICT ( entity_id, name, mac, addresses ) DO NOTHING
`

const createIfNotExistsEntityStateQuery = `
-- This query inserts rows into the entities table. By design, it
-- errors when an entity with the same namespace and name already
-- exists.
--
WITH entity AS (
	INSERT INTO entities ( namespace, name, expires_at, last_seen, selectors, annotations )
	VALUES ( $1, $2, now(), $3, $4, $5 )
	RETURNING id
), system AS (
	INSERT INTO entities_systems
	SELECT
	entity.id,
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
	FROM entity
)
INSERT INTO entities_networks
SELECT
entity.id,
nics.name,
nics.mac,
nics.addresses
FROM entity, UNNEST(
	cast($19 AS text[]),
	cast($20 AS text[]),
	cast($21 AS jsonb[])
)
AS nics ( name, mac, addresses )
`

const updateIfExistsEntityStateQuery = `
-- This query updates the entity, but only if it exists.
--
WITH entity AS (
	SELECT id FROM entities
	WHERE entities.namespace = $1 AND entities.name = $2
), upd AS (
	UPDATE entities
	SET last_seen = $3, selectors = $4, annotations = $5
	FROM entity
	WHERE entity.id = entities.id
), system AS (
	SELECT
	entity.id AS entity_id,
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
	FROM entity
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
	USING entity, networks
	WHERE entity_id = entity.id
	AND ( entities_networks.name, entities_networks.mac, entities_networks.addresses ) NOT IN (SELECT * FROM networks)
), ins AS (
	-- Insert the networks that are new in the current set.
	INSERT INTO entities_networks
	SELECT entity.id, networks.name, networks.mac, networks.addresses
	FROM entity, networks
	ON CONFLICT ( entity_id, name, mac, addresses ) DO NOTHING
)
SELECT * FROM entity;
`

const getEntityStateQuery = `
-- This query fetches a single entity, or nothing.
--
WITH entity AS (
	SELECT entities.id AS id
	FROM entities
	WHERE entities.namespace = $1 AND entities.name = $2
), network AS (
	SELECT
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entity
	WHERE entities_networks.entity_id = entity.id
)
SELECT 
    entities.namespace,
    entities.name,
	entities.last_seen,
	entities.selectors,
	entities.annotations,
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
FROM entities LEFT OUTER JOIN entities_systems ON entities.id = entities_systems.entity_id, network, entity
WHERE entities.id = entity.id
`

const getEntityStatesQuery = `
-- This query fetches multiple entity states, including network information.
--
WITH entity AS (
	SELECT entities.id AS id
	FROM entities
	WHERE entities.namespace = $1 AND entities.name IN (SELECT unnest($2::text[]))
), network AS (
	SELECT
		entities_networks.entity_id AS entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entity
	WHERE entities_networks.entity_id = entity.id
	GROUP BY entities_networks.entity_id
)
SELECT
    entities.namespace,
    entities.name,
	entities.last_seen,
	entities.selectors,
	entities.annotations,
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
FROM entities LEFT OUTER JOIN entities_systems ON entities.id = entities_systems.entity_id, network, entity
WHERE entities.id = entity.id AND entities.id = network.entity_id
`

const deleteEntityStateQuery = `
-- This query deletes a single entity.
--
DELETE FROM entities WHERE entities.namespace = $1 AND entities.name = $2;
`

const listEntityStateQuery = `
-- This query lists entities from a given namespace.
--
with network AS (
	SELECT
		entities.id as entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entities
	WHERE entities.namespace = $1 OR $1 IS NULL
	GROUP BY entities.id
)
SELECT 
    entities.namespace,
    entities.name,
	entities.last_seen,
	entities.selectors,
	entities.annotations,
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
FROM entities LEFT OUTER JOIN entities_systems ON entities.id = entities_systems.entity_id, network
WHERE network.entity_id = entities.id
ORDER BY ( entities.namespace, entities.name ) ASC
LIMIT $2
OFFSET $3
`

const listEntityStateDescQuery = `
-- This query lists entities from a given namespace.
--
with network AS (
	SELECT
		entities.id as entity_id,
		array_agg(entities_networks.name) AS names,
		array_agg(entities_networks.mac) AS macs,
		array_agg(entities_networks.addresses) AS addresses
	FROM entities_networks, entities
	WHERE entities.namespace = $1 OR $1 IS NULL
	GROUP BY entities.id
)
SELECT 
    entities.namespace,
    entities.name,
	entities.last_seen,
	entities.selectors,
	entities.annotations,
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
FROM entities LEFT OUTER JOIN entities_systems ON entities.id = entities_systems.entity_id, network
WHERE network.entity_id = entities.id
ORDER BY ( entities.namespace, entities.name ) DESC
LIMIT $2
OFFSET $3
`

const existsEntityStateQuery = `
-- This query discovers if an entity exists, without retrieving it.
--
SELECT true FROM entities
WHERE entities.namespace = $1 AND entities.name = $2;
`

const patchEntityStateQuery = `
-- This query updates only the last_seen and expires_at values of an
-- entity.
--
UPDATE entities
SET expires_at = $3 + now(), last_seen = $4
WHERE entities.namespace = $1 AND entities.name = $2;
`

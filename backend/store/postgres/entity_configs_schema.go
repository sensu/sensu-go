package postgres

const createOrUpdateEntityConfigQuery = `
-- This query creates a new entity config, or updates it if it already exists.
--
-- Parameters:
-- $1: The entity namespace.
-- $2: The entity name.
-- $3: The label selectors of the entity config.
-- $4: The annotations of the entity config.
-- $5: The user who created the entity.
-- $6: The entity class.
-- $7: The username the entity is connecting as, if the entity is an agent.
-- $8: The entity's subscriptions.
-- $9: Whether deregistration is enabled/disabled.
-- $10: The deregistration handler to use.
-- $11: A list of keepalive handlers.
-- $12: A list of keywords to redact from logs.
-- $13: The entity config ID.
-- $14: The namespace ID.
-- $15: The time that the entity config was created.
-- $16: The time that the entity config was last updated.
-- $17: The time that the entity config was soft deleted.
--
WITH ignored AS (
	SELECT
		$13::bigint,
		$15::timestamptz,
		$16::timestamptz,
		$17::timestamptz
), namespace AS (
	SELECT COALESCE (
		NULLIF($14, 0),
		(SELECT id FROM namespaces WHERE name = $1 AND deleted_at IS NULL)
	) AS id
)
INSERT INTO entity_configs (
	namespace_id,
	name,
	selectors,
	annotations,
	created_by,
	entity_class,
	sensu_user,
	subscriptions,
	deregister,
	deregistration,
	keepalive_handlers,
	redact,
	deleted_at
) VALUES (
	(SELECT id FROM namespace),
	$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL
)
ON CONFLICT ( namespace_id, name )
DO UPDATE
SET
	selectors = $3,
	annotations = $4,
	created_by = $5,
	entity_class = $6,
	sensu_user = $7,
	subscriptions = $8,
	deregister = $9,
	deregistration = $10,
	keepalive_handlers = $11,
	redact = $12,
	deleted_at = NULL
`

const createIfNotExistsEntityConfigQuery = `
-- This query creates a new entity config, or updates it if it exists and has
-- been soft deleted. By design, it errors when an entity config with the same
-- name already exists and has not been soft deleted.
--
WITH ignored AS (
	SELECT
		$13::bigint,
		$15::timestamptz,
		$16::timestamptz,
		$17::timestamptz
), namespace AS (
	SELECT COALESCE (
		NULLIF($14, 0),
		(SELECT id FROM namespaces WHERE name = $1 AND deleted_at IS NULL)
	) AS id
), upsert AS (
	UPDATE entity_configs
	SET
		selectors = $3,
		annotations = $4,
		created_by = $5,
		entity_class = $6,
		sensu_user = $7,
		subscriptions = $8,
		deregister = $9,
		deregistration = $10,
		keepalive_handlers = $11,
		redact = $12,
		deleted_at = NULL
	WHERE
		name = $2 AND
		namespace_id = (SELECT id FROM namespace) AND
		entity_configs.deleted_at IS NOT NULL
	RETURNING *
)
INSERT INTO entity_configs (
	namespace_id,
	name,
	selectors,
	annotations,
	created_by,
	entity_class,
	sensu_user,
	subscriptions,
	deregister,
	deregistration,
	keepalive_handlers,
	redact
)
SELECT (SELECT id FROM namespace), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
WHERE NOT EXISTS (SELECT * FROM upsert)
`

const updateIfExistsEntityConfigQuery = `
-- This query updates the entity config, but only if it exists.
--
WITH ignored AS (
	SELECT
		$13::bigint,
		$15::timestamptz,
		$16::timestamptz,
		$17::timestamptz
), namespace AS (
	SELECT COALESCE (
		NULLIF($14, 0),
		(SELECT id FROM namespaces WHERE name = $1 AND deleted_at IS NULL)
	) AS id
), config AS (
	SELECT id FROM entity_configs
	WHERE
		namespace_id = (SELECT id FROM namespace) AND
		name = $2
), upd AS (
	UPDATE entity_configs
	SET
		selectors = $3,
		annotations = $4,
		created_by = $5,
		entity_class = $6,
		sensu_user = $7,
		subscriptions = $8,
		deregister = $9,
		deregistration = $10,
		keepalive_handlers = $11,
		redact = $12,
		deleted_at = NULL
	FROM config
	WHERE config.id = entity_configs.id
)
SELECT * FROM config;
`

const getEntityConfigQuery = `
-- This query fetches a single entity config, or nothing.
--
SELECT
	namespaces.name,
	entity_configs.name,
	entity_configs.selectors,
	entity_configs.annotations,
	entity_configs.created_by,
	entity_configs.entity_class,
	entity_configs.sensu_user,
	entity_configs.subscriptions,
	entity_configs.deregister,
	entity_configs.deregistration,
	entity_configs.keepalive_handlers,
	entity_configs.redact,
	entity_configs.id,
	namespaces.id,
	entity_configs.created_at,
	entity_configs.updated_at,
	entity_configs.deleted_at
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE
	namespaces.name = $1 AND
	entity_configs.name = $2 AND
	entity_configs.deleted_at IS NULL
`

const getEntityConfigsQuery = `
-- This query fetches multiple entity configs.
--
SELECT
	namespaces.name,
	entity_configs.name,
	entity_configs.selectors,
	entity_configs.annotations,
	entity_configs.created_by,
	entity_configs.entity_class,
	entity_configs.sensu_user,
	entity_configs.subscriptions,
	entity_configs.deregister,
	entity_configs.deregistration,
	entity_configs.keepalive_handlers,
	entity_configs.redact,
	entity_configs.id,
	namespaces.id,
	entity_configs.created_at,
	entity_configs.updated_at,
	entity_configs.deleted_at
FROM entity_configs
LEFT OUTER JOIN namespaces ON namespaces.id = entity_configs.namespace_id
WHERE
	namespaces.name = $1 AND
	namespaces.deleted_at IS NULL AND
	entity_configs.name IN (SELECT unnest($2::text[])) AND
	entity_configs.deleted_at IS NULL
`

const hardDeleteEntityConfigQuery = `
-- This query deletes an entity config. Any related entity, system & network
-- state will also be deleted via ON DELETE CASCADE triggers.
--
-- Parameters:
-- $1 Namespace
-- $2 Entity name
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
)
DELETE FROM entity_configs
WHERE
	namespace_id = (SELECT id FROM namespace) AND
	name = $2;
`

const deleteEntityConfigQuery = `
-- This query soft deletes an entity config.
--
-- Parameters:
-- $1 Namespace
-- $2 Entity name
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
)
UPDATE entity_configs
SET deleted_at = now()
WHERE
	namespace_id = (SELECT id FROM namespace) AND
	name = $2;
`

const hardDeletedEntityConfigQuery = `
-- This query discovers if an entity config has been hard deleted.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
)
SELECT NOT EXISTS (
	SELECT true FROM entity_configs
	WHERE
		namespace_id = (SELECT id FROM namespace) AND
		name = $2
);
`

const listEntityConfigQuery = `
-- This query lists entity configs from a given namespace.
--
SELECT
	namespaces.name,
	entity_configs.name,
	entity_configs.selectors,
	entity_configs.annotations,
	entity_configs.created_by,
	entity_configs.entity_class,
	entity_configs.sensu_user,
	entity_configs.subscriptions,
	entity_configs.deregister,
	entity_configs.deregistration,
	entity_configs.keepalive_handlers,
	entity_configs.redact,
	entity_configs.id,
	entity_configs.namespace_id,
	entity_configs.created_at,
	entity_configs.updated_at,
	entity_configs.deleted_at
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE
	namespaces.name = $1 OR $1 IS NULL AND
	namespaces.deleted_at IS NULL AND
	entity_configs.deleted_at IS NULL
ORDER BY ( namespaces.name, entity_configs.name ) ASC
LIMIT $2
OFFSET $3
`

const listEntityConfigDescQuery = `
-- This query lists entities from a given namespace.
--
SELECT
	namespaces.name,
	entity_configs.name,
	entity_configs.selectors,
	entity_configs.annotations,
	entity_configs.created_by,
	entity_configs.entity_class,
	entity_configs.sensu_user,
	entity_configs.subscriptions,
	entity_configs.deregister,
	entity_configs.deregistration,
	entity_configs.keepalive_handlers,
	entity_configs.redact,
	entity_configs.id,
	entity_configs.namespace_id,
	entity_configs.created_at,
	entity_configs.updated_at,
	entity_configs.deleted_at
FROM entity_configs
LEFT OUTER JOIN namespaces ON namespaces.id = entity_configs.namespace_id
WHERE
	namespaces.name = $1 OR $1 IS NULL AND
	entity_configs.deleted_at IS NULL
ORDER BY ( namespaces.name, entity_configs.name ) DESC
LIMIT $2
OFFSET $3
`

const countEntityConfigQuery = `
-- This query counts entity configs from a given namespace and entity class.
--
SELECT
	COUNT(*)
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE
	(namespaces.name = $1 OR $1 IS NULL) AND
	(entity_configs.entity_class = $2 OR $2 IS NULL) AND
	namespaces.deleted_at IS NULL AND
	entity_configs.deleted_at IS NULL;
`

const existsEntityConfigQuery = `
-- This query discovers if an entity config exists, without retrieving it.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
)
SELECT true FROM entity_configs
WHERE
	namespace_id = (SELECT id FROM namespace) AND
	name = $2 AND
	deleted_at IS NULL;
`

const pollEntityConfigQuery = `
-- Looks for updates to the entiy_config table
-- since a specified timestamp
--
WITH ignored AS (
	SELECT
		$3::text,
		$4::text,
		$5::text,
		$6::text,
		$7::text,
		$8::text[],
		$9::boolean,
		$10::text,
		$11::text[],
		$12::text[],
		$13::bigint,
		$14::bigint,
		$15::timestamptz,
		$17::timestamptz
)
SELECT
	namespaces.name,
	entity_configs.name,
	entity_configs.selectors,
	entity_configs.annotations,
	entity_configs.created_by,
	entity_configs.entity_class,
	entity_configs.sensu_user,
	entity_configs.subscriptions,
	entity_configs.deregister,
	entity_configs.deregistration,
	entity_configs.keepalive_handlers,
	entity_configs.redact,
	entity_configs.id,
	entity_configs.namespace_id,
	entity_configs.created_at,
	entity_configs.updated_at,
	entity_configs.deleted_at
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE entity_configs.updated_at >= $16
AND CASE
		WHEN $1 <> '' THEN namespaces.name = $1
		ELSE true
	END
AND CASE
		WHEN $2 <> '' THEN entity_configs.name = $2
		ELSE TRUE
	END
`

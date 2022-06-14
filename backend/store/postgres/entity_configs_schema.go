package postgres

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
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
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
) VALUES ( (SELECT id FROM namespace), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12 )
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
	redact = $12
`

const createIfNotExistsEntityConfigQuery = `
-- This query inserts rows into the entity_configs table. By design, it
-- errors when an entity with the same namespace and name already
-- exists.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
), config AS (
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
	) VALUES ( (SELECT id FROM namespace), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12 )
	RETURNING id
)
SELECT config.id FROM config
`

const updateIfExistsEntityConfigQuery = `
-- This query updates the entity config, but only if it exists.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
), config AS (
	SELECT id FROM entity_configs
	WHERE namespace_id = (SELECT id FROM namespace) AND name = $2
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
		redact = $12
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
	entity_configs.redact
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE namespaces.name = $1 AND entity_configs.name = $2
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
	entity_configs.redact
FROM entity_configs
LEFT OUTER JOIN namespaces ON namespaces.id = entity_configs.namespace_id
WHERE namespaces.name = $1 AND entity_configs.name IN (SELECT unnest($2::text[]))
`

const deleteEntityConfigQuery = `
-- This query deletes an entity config. Any related entity, system & network
-- state will also be deleted via ON DELETE CASCADE triggers.
--
-- Parameters:
-- $1 Namespace
-- $2 Entity name
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
DELETE FROM entity_configs WHERE entity_configs.namespace_id = (SELECT id FROM namespace) AND entity_configs.name = $2;
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
	entity_configs.redact
FROM entity_configs
LEFT OUTER JOIN namespaces ON entity_configs.namespace_id = namespaces.id
WHERE namespaces.name = $1 OR $1 IS NULL
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
	entity_configs.redact
FROM entity_configs
LEFT OUTER JOIN namespaces ON namespaces.id = entity_configs.namespace_id
WHERE namespaces.name = $1 OR $1 IS NULL
ORDER BY ( namespaces.name, entity_configs.name ) DESC
LIMIT $2
OFFSET $3
`

const existsEntityConfigQuery = `
-- This query discovers if an entity config exists, without retrieving it.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
SELECT true FROM entity_configs
WHERE namespace_id = (SELECT id FROM namespace) AND name = $2;
`

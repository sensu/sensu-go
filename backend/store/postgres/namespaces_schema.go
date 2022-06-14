package postgres

const namespaceSchema = `
--
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

const createOrUpdateNamespaceQuery = `
-- This query creates a new namespace, or updates it if it already exists.
--
-- Parameters:
-- $1: The namespace name.
-- $2: The label selectors of the entity config.
-- $3: The annotations of the entity config.
--
INSERT INTO namespaces (
	name,
	selectors,
	annotations
) VALUES ( $1, $2, $3 )
ON CONFLICT ( name )
DO UPDATE
SET
	selectors = $2,
	annotations = $3
`

const createIfNotExistsNamespaceQuery = `
-- This query inserts rows into the namespaces table. By design, it
-- errors when a namespace with the same name already exists.
--
WITH namespace AS (
	INSERT INTO namespaces (
		name,
		selectors,
		annotations
	) VALUES ( $1, $2, $3 )
	RETURNING id
)
SELECT namespace.id FROM namespace
`

const updateIfExistsNamespaceQuery = `
-- This query updates the namespace, but only if it exists.
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
), upd AS (
	UPDATE namespaces
	SET
		selectors = $2,
		annotations = $3
	FROM namespace
	WHERE namespace.id = namespaces.id
)
SELECT * FROM namespace;
`

const getNamespaceQuery = `
-- This query fetches a single namespace, or nothing.
--
SELECT
	name,
	selectors,
	annotations
FROM namespaces
WHERE name = $1
`

const getNamespacesQuery = `
-- This query fetches multiple namespaces.
--
SELECT
	name,
	selectors,
	annotations
FROM namespaces WHERE name IN (SELECT unnest($1::text[]))
`

const deleteNamespaceQuery = `
-- This query deletes a namespace. Any related data (e.g. entity configs,
-- entity state, etc.) will also be deleted via ON DELETE CASCADE triggers.
--
-- Parameters:
-- $1 Name
DELETE FROM namespaces WHERE name = $1;
`

const listNamespaceQuery = `
-- This query lists all namespaces.
--
SELECT
	name,
	selectors,
	annotations
FROM namespaces
ORDER BY ( name ) ASC
LIMIT $2
OFFSET $3
`

const listNamespaceDescQuery = `
-- This query lists all namespaces.
--
SELECT
	name,
	selectors,
	annotations
FROM namespaces
ORDER BY ( name ) DESC
LIMIT $2
OFFSET $3
`

const existsNamespaceQuery = `
-- This query discovers if a namespace exists, without retrieving it.
--
SELECT true FROM namespaces
WHERE name = $1;
`

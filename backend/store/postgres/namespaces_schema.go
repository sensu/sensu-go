package postgres

const createOrUpdateNamespaceQuery = `
-- This query creates a new namespace, or updates it if it already exists.
--
-- Parameters:
-- $1: The namespace name.
-- $2: The label selectors of the entity config.
-- $3: The annotations of the entity config.
-- $4: The namespace ID.
-- $5: The time that the namespace was created.
-- $6: The time that the namespace was last updated.
-- $7: The time that the namespace was soft deleted.
--
WITH ignored AS (
	SELECT
		$4::bigint,
		$5::timestamptz,
		$6::timestamptz,
		$7::timestamptz
)
INSERT INTO namespaces (
	name,
	selectors,
	annotations,
	deleted_at
) VALUES ( $1, $2, $3, NULL )
ON CONFLICT ( name )
DO UPDATE
SET
	selectors = $2,
	annotations = $3,
	deleted_at = NULL
`

const createIfNotExistsNamespaceQuery = `
-- This query creates a new namespace, or updates it if it exists and has been
-- soft deleted. By design, it errors when a namespace with the same name
-- already exists and has not been soft deleted.
--
WITH ignored AS (
	SELECT
		$4::bigint,
		$5::timestamptz,
		$6::timestamptz,
		$7::timestamptz
), upsert AS (
	UPDATE namespaces
	SET
		selectors = $2,
		annotations = $3,
		deleted_at = NULL
	WHERE
		name = $1 AND
		namespaces.deleted_at IS NOT NULL
	RETURNING *
)
INSERT INTO namespaces (name, selectors, annotations)
SELECT $1, $2, $3
WHERE NOT EXISTS (SELECT * FROM upsert)
`

const updateIfExistsNamespaceQuery = `
-- This query updates the namespace, but only if it exists.
--
WITH ignored AS (
	SELECT
		$4::bigint,
		$5::timestamptz,
		$6::timestamptz,
		$7::timestamptz
), namespace AS (
	SELECT id FROM namespaces WHERE name = $1
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
	annotations,
	id,
	created_at,
	updated_at,
	deleted_at
FROM namespaces
WHERE
	name = $1 AND
	deleted_at IS NULL
`

const getNamespacesQuery = `
-- This query fetches multiple namespaces.
--
SELECT
	name,
	selectors,
	annotations,
	id,
	created_at,
	updated_at,
	deleted_at
FROM namespaces
WHERE
	name IN (SELECT unnest($1::text[])) AND
	deleted_at IS NULL
`

const hardDeleteNamespaceQuery = `
-- This query deletes a namespace. Any related data (e.g. entity configs,
-- entity state, etc.) will also be deleted via ON DELETE CASCADE triggers.
--
-- Parameters:
-- $1 Name
DELETE FROM namespaces WHERE name = $1;
`

const deleteNamespaceQuery = `
-- This query soft deletes a namespace.
--
-- Parameters:
-- $1 Name
UPDATE namespaces
SET deleted_at = now()
WHERE
	name = $1;
`

const hardDeletedNamespaceQuery = `
-- This query discovers if a namespace has been hard deleted.
--
SELECT NOT EXISTS (
	SELECT true FROM namespaces
	WHERE
		name = $1
);
`

const countNamespaceQuery = `
-- This query lists all namespaces.
--
SELECT
	count(*)
FROM namespaces
WHERE deleted_at IS NULL;
`

const listNamespaceQuery = `
-- This query lists all namespaces.
--
SELECT
	name,
	selectors,
	annotations,
	id,
	created_at,
	updated_at,
	deleted_at
FROM namespaces
WHERE deleted_at IS NULL
ORDER BY ( name ) ASC
LIMIT $1
OFFSET $2
`

const listNamespaceDescQuery = `
-- This query lists all namespaces.
--
SELECT
	name,
	selectors,
	annotations,
	id,
	created_at,
	updated_at,
	deleted_at
FROM namespaces
WHERE deleted_at IS NULL
ORDER BY ( name ) DESC
LIMIT $1
OFFSET $2
`

const existsNamespaceQuery = `
-- This query discovers if a namespace exists, without retrieving it.
--
SELECT true FROM namespaces
WHERE
	name = $1 AND
	deleted_at IS NULL;
`

const isEmptyNamespaceQuery = `
-- This query determines if a namespace is empty or not.
--
WITH namespace AS (
	SELECT id FROM namespaces WHERE name = $1
), entity_configs AS (
	SELECT COUNT(*) FROM entity_configs, namespace WHERE entity_configs.namespace_id = namespace.id
), entity_states AS (
	SELECT COUNT(*) FROM entity_states, namespace WHERE entity_states.namespace_id = namespace.id
)
SELECT
	entity_configs.count + entity_states.count AS count
FROM entity_configs, entity_states, namespace
WHERE namespace.id IS NOT NULL;
`

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
	SELECT $4::bigint, $5, $6, $7
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
-- This query inserts rows into the namespaces table. By design, it
-- errors when a namespace with the same name already exists.
--
WITH ignored AS (
	SELECT $4::bigint, $5, $6, $7
), namespace AS (
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
WITH ignored AS (
	SELECT $4::bigint, $5, $6, $7
), namespace AS (
	SELECT id FROM namespaces
	WHERE name = $1
), upd AS (
	UPDATE namespaces
	SET
		selectors = $2,
		annotations = $3,
		deleted_at = NULL
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

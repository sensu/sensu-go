package postgres

const updateRingSubscribersQuery = `
-- This query updates the ring_subscribers table, "advancing" the ring.
-- To do so, it must compute the next pointer to the head of the ring.
-- The next pointer will be greater than the current pointer, unless the next
-- pointer would wrap around to be a lesser value.
--
-- Parameters:
-- $1 The name of the ring
-- $2 The name of the subscriber
-- $3 The number of entities to advance by, minus 1. (This value is applied as
--    the offset, so math cannot be done on it in the function.)
-- $4 The time offset
WITH current_pointer AS (
	SELECT entity_states.name AS name
	FROM rings, entity_states, ring_entities, ring_subscribers
	WHERE
		rings.name = $1 AND
		ring_entities.ring_id = rings.id AND
		ring_entities.entity_id = entity_states.id AND
		ring_entities.entity_id = ring_subscribers.pointer AND
		ring_subscribers.name = $2 AND
		ring_subscribers.ring_id = rings.id
),
next_all AS (
	(
		-- In most iterations, the pointer value will be higher than it was
		-- previously.
		SELECT entity_states.id AS pointer, entity_states.name AS entity_name
		FROM rings, entity_states, ring_entities, ring_subscribers, current_pointer
		WHERE
			rings.name = $1 AND
			ring_entities.ring_id = rings.id AND
			ring_entities.entity_id = entity_states.id AND
			ring_subscribers.name = $2 AND
			ring_subscribers.ring_id = rings.id AND
			entity_states.name > COALESCE(current_pointer.name, '') AND
			entity_states.expires_at > now()
		ORDER BY entity_states.name ASC
	)
	UNION ALL
	(
		-- If the pointer was at the end or NULL, it will get set to the first
		-- entity in the selection (considering the offset).
		SELECT ring_entities.entity_id AS pointer, entity_states.name AS entity_name
		FROM rings, entity_states, ring_entities, ring_subscribers
		WHERE
			rings.name = $1 AND
			ring_entities.ring_id = rings.id AND
			ring_entities.entity_id = entity_states.id AND
			ring_subscribers.name = $2 AND
			ring_subscribers.ring_id = rings.id AND
			entity_states.expires_at > now()
		ORDER BY entity_states.name ASC
	)
),
next_numbered AS (
	SELECT next_all.pointer AS pointer, next_all.entity_name AS entity_name, row_number () OVER () AS rnum
	FROM next_all
),
next AS (
	-- generate_series() is used to make sure that the offset does not go
	-- beyond the bounds of the total number of entities, creating repetitions.
	SELECT next_numbered.pointer AS pointer, next_numbered.entity_name AS entity_name, next_numbered.rnum as rnum
	FROM next_numbered, generate_series(1, $3 + 1)
	ORDER BY generate_series, rnum
	LIMIT 1
	OFFSET $3
)
-- The actual update is quite simple. We just set last_updated to the current
-- time, and we set pointer to be the next pointer, which is computed above.
UPDATE ring_subscribers
SET (last_updated, pointer) = (
	now(),
	next.pointer
)
FROM rings, entity_states, ring_entities, next
WHERE
	rings.name = $1 AND
	ring_entities.ring_id = rings.id AND
	ring_entities.entity_id = entity_states.id AND
	entity_states.id = next.pointer AND
	ring_subscribers.ring_id = rings.id AND
	entity_states.expires_at > now() AND
	ring_subscribers.last_updated + $4 < now() AND
	ring_subscribers.name = $2
RETURNING entity_states.name
`

const getRingEntitiesQuery = `
-- This query finds the next N entities in the ring. To do so, it finds up to N
-- rows that are lexically larger than the current pointer for the subscriber.
-- If there aren't enough rows that fit the criteria, the smaller rows or even
-- the current row itself is selected with UNION ALL.
-- The query isn't guaranteed to fill the limit; it's up to the caller to deal
-- with that case.
--
-- Parameters:
-- $1: The name of the ring
-- $2: The name of the subscriber
-- $3: The number of entities to get per call. (LIMIT)
-- $4: Timeout. Entities updated less than now() - timeout are excluded.
-- 
-- Delete expired backend entities. This is only necessary because backend
-- entities are ephemeral, named via random identifier, and not accessible
-- to users. This should change in the future.
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = 'bsmd' OR namespaces.name = 'global'
), deletes AS (
	DELETE FROM entity_states
	WHERE entity_states.namespace_id = (SELECT id FROM namespace) AND
	entity_states.expires_at < now()
),
current_pointer AS (
	SELECT entity_states.name AS name
	FROM rings, entity_states, ring_entities, ring_subscribers
	WHERE rings.name = $1 AND
	ring_entities.ring_id = rings.id AND
	ring_entities.entity_id = ring_subscribers.pointer AND
	entity_states.id = ring_entities.entity_id AND
	ring_subscribers.name = $2 AND
	ring_subscribers.ring_id = rings.id
)
(
	-- This part of the query selects rows that have a name larger than or
	-- equal to the current pointer's name.
	SELECT entity_states.name AS name
	FROM rings, ring_entities, entity_states, ring_subscribers, current_pointer
	WHERE
		rings.name = $1 AND
		ring_entities.ring_id = rings.id AND
		ring_entities.entity_id = entity_states.id AND
		ring_subscribers.ring_id = rings.id AND
		ring_subscribers.name = $2 AND
		entity_states.name >= COALESCE(current_pointer.name, '') AND
		entity_states.expires_at > now()
	ORDER BY entity_states.name ASC
	LIMIT $3
)
UNION ALL
(
	-- This part of the query selects rows that have a name smaller than
	-- pointer.name. In many cases it will never be evaluated, as the limit
	-- will have already been satisfied by the first half of the union.
	SELECT entity_states.name AS name
	FROM rings, ring_entities, entity_states, ring_subscribers
	WHERE
		rings.name = $1 AND
		ring_entities.ring_id = rings.id AND
		ring_entities.entity_id = entity_states.id AND
		ring_subscribers.ring_id = rings.id AND
		ring_subscribers.name = $2 AND
		entity_states.expires_at > now()
	ORDER BY entity_states.name ASC
	LIMIT $3
)
LIMIT $3;
`

const insertRingQuery = `
-- This query creates a new ring, or does nothing if it already exists.
INSERT INTO rings (name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING;
`

const updateEntityStateExpiresAtQuery = `
-- This query updates an entity state's expires_at value.
--
-- Parameters:
-- $1: The entity namespace
-- $2: The entity name
-- $3: The keepalive timeout of the entity
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
UPDATE entity_states
SET expires_at = now() + $3
WHERE namespace_id = (SELECT id FROM namespace) AND name = $2
`

const insertRingEntityQuery = `
-- This query creates an association between a ring and an entity.
--
-- Parameters:
-- $1: The entity namespace
-- $2: The entity name
-- $3: The ring name
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
INSERT INTO ring_entities ( ring_id, entity_id )
SELECT rings.id, entity_states.id
FROM rings, entity_states
WHERE
	entity_states.namespace_id = (SELECT id FROM namespace) AND
	entity_states.name = $2 AND
	rings.name = $3
LIMIT 1
ON CONFLICT DO NOTHING
RETURNING TRUE;
`

const insertRingSubscriberQuery = `
-- This query creates a ring subscriber, or if it exists, does nothing.
--
-- Parameters:
-- $1: The name of the ring
-- $2: The name of the ring subscriber
--
INSERT INTO ring_subscribers ( ring_id, name, pointer )
(
	(
		-- This case is for when the ring already has entities
		SELECT rings.id AS ring_id, $2 AS subscriber_name, entity_states.id AS pointer
		FROM rings, entity_states, ring_entities
		WHERE
			rings.name = $1 AND
			rings.id = ring_entities.ring_id AND
			entity_states.id = ring_entities.entity_id
		ORDER BY entity_states.name ASC
		LIMIT 1
	)
	UNION ALL
	(
		-- This case is for when there are no entities in the ring
		SELECT rings.id AS ring_id, $2 AS subscriber_name, NULL AS pointer
		FROM rings
		WHERE rings.name = $1
	)
	ORDER BY pointer ASC
	LIMIT 1
)
ON CONFLICT DO NOTHING
RETURNING TRUE;
`

const getRingLengthQuery = `
-- This query gets the length of the named ring.
--
-- Parameters:
-- $1: Ring name
SELECT count(entity_states.id)
FROM rings, entity_states, ring_entities
WHERE
	rings.name = $1 AND
	rings.id = ring_entities.ring_id AND
	entity_states.id = ring_entities.entity_id
`

const deleteRingEntityQuery = `
-- This query deletes a named ring entity. It will only delete the association
-- between a ring and an entity, not the ring or the entity itself.
--
-- Parameters:
-- $1: Namespace
-- $2: Ring name
-- $3: Entity name
--
WITH namespace AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
DELETE FROM ring_entities
USING entity_states, rings
WHERE
	entity_states.id = ring_entities.entity_id AND
	rings.id = ring_entities.ring_id AND
	entity_states.namespace_id = (SELECT id FROM namespace) AND
	entity_states.name = $3 AND
	rings.name = $2;
`

// The notification mechanism for ring subscribers
const notifyRingChannelQuery = `SELECT pg_notify($1::text, '');`

package postgres

const DeleteEventQuery = `
DELETE FROM events
WHERE
	sensu_namespace = $1 AND
	sensu_entity = $2 AND
	sensu_check = $3;
`

const GetAllEventsQuery = `
SELECT
	last_ok,
	occurrences,
	occurrences_wm,
	history_ts,
	history_status,
	history_index,
	serialized
FROM events
ORDER BY ( sensu_namespace, sensu_entity, sensu_check )
LIMIT $1
OFFSET $2;
`

const GetEventsQuery = `
SELECT
	last_ok,
	occurrences,
	occurrences_wm,
	history_ts,
	history_status,
	history_index,
	serialized
FROM events
WHERE sensu_namespace = $1
ORDER BY ( sensu_entity, sensu_check )
LIMIT $2
OFFSET $3;
`

const GetEventsByEntityQuery = `
SELECT
	last_ok,
	occurrences,
	occurrences_wm,
	history_ts,
	history_status,
	history_index,
	serialized
FROM events
WHERE sensu_namespace = $1 AND sensu_entity = $2
ORDER BY ( sensu_check )
LIMIT $3
OFFSET $4;
`

const GetEventByEntityCheckQuery = `
SELECT
	last_ok,
	occurrences,
	occurrences_wm,
	history_ts,
	history_status,
	history_index,
	serialized
FROM events
WHERE
	sensu_namespace = $1 AND
	sensu_entity = $2 AND
	sensu_check = $3;
`

// UpdateEventQuery does an "upsert" - insert or update, depending if the
// event already exists. Requires Postgres >=9.5.
// $1 -> namespace
// $2 -> entity
// $3 -> check
// $4 -> status
// $5 -> last_ok
// $6 -> serialized
// $7 -> selectors
const UpdateEventQuery = `
INSERT INTO events (
	sensu_namespace,
	sensu_entity,
	sensu_check,
	status,
	last_ok,
	occurrences,
	occurrences_wm,
	history_status,
	history_ts,
	serialized,
	selectors
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	CASE
		WHEN $4 = 0 THEN $5
		ELSE 0
	END,
	1,
	1,
	ARRAY[$4]::integer[],
	ARRAY[$5]::integer[],
	$6,
	$7
)
ON CONFLICT (
	sensu_namespace,
	sensu_entity,
	sensu_check
) DO UPDATE SET (
	status,
	last_ok,
	occurrences,
	occurrences_wm,
	history_status[events.history_index],
	history_ts[events.history_index],
	history_index,
	serialized,
	previous_serialized,
	selectors
) = (
	$4,
	CASE
		WHEN $4 = 0 THEN $5
		ELSE events.last_ok
	END,
	CASE
		WHEN events.status = $4 THEN events.occurrences + 1
		ELSE 1
	END,
	CASE
		WHEN events.status = 0 AND $4 != 0 THEN 1
		WHEN events.status != $4 THEN events.occurrences_wm
		WHEN events.occurrences < events.occurrences_wm THEN events.occurrences_wm
		ELSE events.occurrences_wm + 1
	END,
	$4,
	$5,
	(events.history_index % 21) + 1,
	$6,
	events.serialized,
	$7
)
RETURNING
	history_ts,
	history_status,
	history_index,
	last_ok,
	occurrences,
	occurrences_wm,
	previous_serialized;
`

// UpdateEventOnlyQuery updates the serialized event.
const UpdateEventOnlyQuery = `
UPDATE events
SET (serialized, selectors) = ($1, $2)
WHERE
	sensu_namespace = $3 AND
	sensu_entity = $4 AND
	sensu_check = $5;
`

// Requires Postgres >= 9.4
const GetEventCountsByNamespaceQuery = `
SELECT
	sensu_namespace AS namespace,
	count(*) AS events_total,
	count(*) FILTER (WHERE status = 0) AS events_status_ok,
	count(*) FILTER (WHERE status = 1) AS events_status_warning,
	count(*) FILTER (WHERE status = 2) AS events_status_critical,
	count(*) FILTER (WHERE status > 2) AS events_status_other,
	count(*) FILTER (WHERE selectors ->> 'event.check.state' = 'passing') AS events_state_passing,
	count(*) FILTER (WHERE selectors ->> 'event.check.state' = 'flapping') AS events_state_flapping,
	count(*) FILTER (WHERE selectors ->> 'event.check.state' = 'failing') AS events_state_failing
FROM events
GROUP BY namespace;
`

// Requires Postgres >= 9.4
const GetKeepaliveCountsByNamespaceQuery = `
SELECT
	sensu_namespace AS namespace,
		count(*) FILTER (WHERE sensu_check = 'keepalive') AS total,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND status = 0) AS status_ok,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND status = 1) AS status_warning,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND status = 2) AS status_critical,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND status > 2) AS status_other,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'passing') AS state_passing,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'flapping') AS state_flapping,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'failing') AS state_failing
FROM events
GROUP BY namespace;
`

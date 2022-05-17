package postgres

import (
	_ "github.com/lib/pq"
)

// EventsDDL is the postgresql data definition language that defines the
// relational schema of events.
//
// Events are expected to be referred to by the triple of
// ( sensu_namespace, sensu_check, sensu_entity ), and not by their primary
// key. The primary key is intended only for future use by joins that will
// reference the events table.
const EventsDDL = `CREATE TABLE IF NOT EXISTS events (
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

const MigrateEventsID = `ALTER TABLE IF EXISTS events ALTER COLUMN id SET DATA TYPE bigint;
ALTER SEQUENCE IF EXISTS events_id_seq AS bigint;
`

const MigrateAddFlappingColumn = `ALTER TABLE IF EXISTS events ADD COLUMN history_flapping boolean[];
UPDATE events SET history_flapping = array_fill(FALSE, ARRAY[events.history_index]);
`

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

const MigrateAddTotalStateChangeFunc = `
CREATE OR REPLACE FUNCTION total_state_change(history_status INTEGER[])
RETURNS INTEGER AS $$
DECLARE
    state_changes  REAL;
    change_weight  REAL;
    array_len      INTEGER;
    prev_status    INTEGER;
    status         INTEGER;
BEGIN
    array_len := array_length(history_status, 1)::INTEGER;
    IF array_len IS NULL OR array_len < 2 THEN
        RETURN 0;
    END IF;
    state_changes := 0.0;
    change_weight := 0.8;
    prev_status := history_status[1];
    FOREACH status IN ARRAY history_status[2:array_len] LOOP
        IF status != prev_status THEN
            state_changes := state_changes + change_weight;
        END IF;
        change_weight := change_weight + 0.02;
        prev_status := status;
    END LOOP;
    RETURN state_changes / 20.0 * 100.0;
END; $$
LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION sort_history(a BOOLEAN[], idx INTEGER)
RETURNS BOOLEAN[] AS $$
DECLARE
    rhs BOOLEAN[];
    lhs BOOLEAN[];
BEGIN
	rhs := a[1:idx];
	lhs := a[idx + 1:array_length(a, 1)];
	IF lhs IS NULL THEN
	    return rhs;
	END IF;
	IF rhs IS NULL THEN
	    return lhs;
	END IF;
	return (lhs || rhs)::BOOLEAN[];
END; $$
LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION is_flapping(low_flap_threshold INTEGER, high_flap_threshold INTEGER, flapping BOOLEAN, history_status INTEGER[])
RETURNS BOOLEAN AS $$
DECLARE
    state_change INTEGER;
BEGIN
    IF low_flap_threshold = 0 THEN
        return FALSE;
    ELSIF high_flap_threshold = 0 THEN
        return FALSE;
    ELSIF flapping IS NULL THEN
        return FALSE;
    ELSIF flapping THEN
        state_change := total_state_change(history_status);
        return state_change > low_flap_threshold;
    ELSE
        state_change := total_state_change(history_status);
        return state_change >= high_flap_threshold;
    END IF;
END; $$
LANGUAGE plpgsql IMMUTABLE;
`

const DeleteEventQuery = `DELETE FROM events WHERE sensu_namespace = $1 AND sensu_entity = $2 AND sensu_check = $3;`

const GetAllEventsQuery = `SELECT
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
    OFFSET $2;`

const GetEventsQuery = `SELECT
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
    OFFSET $3;`

const GetEventsByEntityQuery = `SELECT
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
    OFFSET $4;`

const GetEventByEntityCheckQuery = `SELECT
        last_ok,
        occurrences,
        occurrences_wm,
        history_ts,
        history_status,
        history_index,
        serialized
    FROM events
    WHERE sensu_namespace = $1 AND sensu_entity = $2 AND sensu_check = $3;`

// UpdateEventQuery does an "upsert" - insert or update, depending if the
// event already exists. Requires Postgres >=9.5.
// $1 -> namespace
// $2 -> entity
// $3 -> check
// $4 -> status
// $5 -> last_ok
// $6 -> serialized
// $7 -> selectors
const UpdateEventQuery = `INSERT INTO events (
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
    RETURNING history_ts, history_status, history_index, last_ok, occurrences, occurrences_wm, previous_serialized;`

// UpdateEventOnlyQuery updates the serialized event.
const UpdateEventOnlyQuery = `UPDATE events SET (serialized, selectors) = ($1, $2) WHERE sensu_namespace = $3 AND sensu_entity = $4 AND sensu_check = $5;`

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

const MigrateAddSelectorColumn = `
ALTER TABLE IF EXISTS events ADD COLUMN selectors jsonb;

CREATE INDEX IF NOT EXISTS idxginevents ON events USING GIN (selectors jsonb_path_ops);
`

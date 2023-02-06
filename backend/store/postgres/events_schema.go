package postgres

// Requires Postgres >= 9.4
const GetEventCountsByNamespaceQuery = `
SELECT
	sensu_namespace AS namespace,
	count(*) AS events_total,
	count(*) FILTER (WHERE selectors ->> 'event.check.status' = 0) AS events_status_ok,
	count(*) FILTER (WHERE selectors ->> 'event.check.status' = 1) AS events_status_warning,
	count(*) FILTER (WHERE selectors ->> 'event.check.status' = 2) AS events_status_critical,
	count(*) FILTER (WHERE selectors ->> 'event.check.status' > 2) AS events_status_other,
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
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.status' = 0) AS status_ok,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.status' = 1) AS status_warning,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.status' = 2) AS status_critical,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.status' > 2) AS status_other,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'passing') AS state_passing,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'flapping') AS state_flapping,
		count(*) FILTER (WHERE sensu_check = 'keepalive' AND selectors ->> 'event.check.state' = 'failing') AS state_failing
FROM events
GROUP BY namespace;
`

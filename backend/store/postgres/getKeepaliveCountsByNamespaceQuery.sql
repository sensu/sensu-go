SELECT
	namespaces.name AS namespace,
		count(*) FILTER (WHERE events.check_name = 'keepalive') AS total,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND (events.selectors ->> 'event.check.status')::INTEGER = 0) AS status_ok,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND (events.selectors ->> 'event.check.status')::INTEGER = 1) AS status_warning,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND (events.selectors ->> 'event.check.status')::INTEGER = 2) AS status_critical,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND (events.selectors ->> 'event.check.status')::INTEGER > 2) AS status_other,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND events.selectors ->> 'event.check.state' = 'passing') AS state_passing,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND events.selectors ->> 'event.check.state' = 'flapping') AS state_flapping,
		count(*) FILTER (WHERE events.check_name = 'keepalive' AND events.selectors ->> 'event.check.state' = 'failing') AS state_failing
FROM events, namespaces
WHERE namespaces.id = events.namespace
GROUP BY namespaces.name;

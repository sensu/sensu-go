SELECT
	namespaces.name AS namespace,
	count(*) AS events_total,
	count(*) FILTER (WHERE (events.selectors ->> 'event.check.status')::INTEGER = 0) AS events_status_ok,
	count(*) FILTER (WHERE (events.selectors ->> 'event.check.status')::INTEGER = 1) AS events_status_warning,
	count(*) FILTER (WHERE (events.selectors ->> 'event.check.status')::INTEGER = 2) AS events_status_critical,
	count(*) FILTER (WHERE (events.selectors ->> 'event.check.status')::INTEGER > 2) AS events_status_other,
	count(*) FILTER (WHERE events.selectors ->> 'event.check.state' = 'passing') AS events_state_passing,
	count(*) FILTER (WHERE events.selectors ->> 'event.check.state' = 'flapping') AS events_state_flapping,
	count(*) FILTER (WHERE events.selectors ->> 'event.check.state' = 'failing') AS events_state_failing
FROM events, namespaces
WHERE namespaces.id = events.namespace
GROUP BY namespaces.name;

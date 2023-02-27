WITH ns AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
	LIMIT 1
)
DELETE FROM events
USING ns
WHERE namespace = ns.id
  AND entity_name = $2
  AND check_name = $3
RETURNING events.id;

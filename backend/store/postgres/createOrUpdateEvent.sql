WITH ns AS (
	SELECT id
	FROM namespaces
	WHERE name = $1
	LIMIT 1
)
INSERT INTO events ( namespace, entity_name, check_name, selectors, serialized )
SELECT ns.id, $2, $3, $4, $5 FROM ns
ON CONFLICT ( namespace, entity_name, check_name )
DO UPDATE SET (
	selectors,
	serialized
) = (
	$4, $5
)
RETURNING id;

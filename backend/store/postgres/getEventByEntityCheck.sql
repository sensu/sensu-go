WITH ns AS (
	SELECT id AS id 
	FROM namespaces
	WHERE name = $1
	LIMIT 1
)
SELECT events.serialized
FROM   events, ns
WHERE  events.namespace = ns.id AND
       events.entity_name = $2 AND
       events.check_name = $3

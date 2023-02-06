package postgres

const queueSchema = `
CREATE TABLE IF NOT EXISTS queue_items (
	id			bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	queue		text NOT NULL,
	updated_at	timestamptz NOT NULL DEFAULT NOW(),
	value		bytea NOT NULL
);
`

const queueEnqueue = `
INSERT INTO queue_items  (queue, value)
	VALUES ($1, $2);
`

const queueReserveItem = `
SELECT
	id, queue, value
 FROM queue_items
WHERE queue = $1
ORDER BY updated_at LIMIT 1
FOR UPDATE SKIP LOCKED;
`

const queueDeleteItem = `DELETE FROM queue_items WHERE id = $1;`

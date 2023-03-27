package postgres

const opcSchema = `
CREATE TABLE IF NOT EXISTS opc (
	id	             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	namespace        bigint REFERENCES namespaces (id) ON DELETE CASCADE,
	operator_type    integer NOT NULL,
	operator_name    text NOT NULL,
	controller       bigint REFERENCES opc (id) ON DELETE SET NULL,
	controller_type  integer DEFAULT 0,
	last_update      bigint NOT NULL,
	timeout_micro    bigint NOT NULL,
	present          boolean NOT NULL,
	metadata         jsonb NOT NULL DEFAULT '{}'::jsonb,
	UNIQUE (namespace, operator_type, operator_name)
);
`

const getOperatorID = `
-- $1 namespace (string)
-- $2 operator type (integer)
-- $3 operator name (string)
WITH ns AS (
	(SELECT id AS id
	FROM namespaces
	WHERE namespaces.name = $1)
	UNION (SELECT NULL AS id)
	ORDER BY id NULLS LAST
	LIMIT 1
)
SELECT opc.id FROM opc, ns
WHERE
	(($1 != '' AND opc.namespace = ns.id) OR (opc.namespace IS NULL AND $1 = ''))
	AND opc.operator_type = $2
	AND opc.operator_name = $3
LIMIT 1
;
`

const opcCheckInInsert = `
-- opcCheckInInsert inserts a new operator into the opc table, thereby checking
-- in the operator.
--
-- $1 namespace (string)
-- $2 operator type (integer)
-- $3 operator name (string)
-- $4 time out microseconds (integer)
-- $5 present (boolean)
-- $6 metadata (json)
-- $7 controller_namespace (string)
-- $8 controller_type (integer)
-- $9 controller_name (string)
WITH ns AS (
	(SELECT id AS id
	FROM namespaces
	WHERE namespaces.name = $1)
	UNION (SELECT null AS id)
	ORDER BY id NULLS LAST
	LIMIT 1
)
, ctlns AS (
	(SELECT id AS id
	FROM namespaces
	WHERE namespaces.name = $7)
	UNION (SELECT null AS id)
	ORDER BY id NULLS LAST
	LIMIT 1
)
, ctl AS (
	(SELECT opc.id AS id, opc.operator_type AS operator_type
	FROM opc, ctlns
	WHERE opc.operator_type = $8
	  AND opc.operator_name = $9
	  AND (opc.namespace IS NULL OR opc.namespace = ctlns.id))
	UNION (SELECT null AS id, NULL AS operator_type)
	ORDER BY id NULLS LAST
	LIMIT 1
)
INSERT INTO opc (
	namespace
  , operator_type
  , operator_name
  , controller
  , controller_type
  , last_update
  , timeout_micro
  , present
  , metadata
)
SELECT ns.id AS namespace
	 , $2 AS operator_type
	 , $3 AS operator_name
	 , ctl.id AS controller
	 , ctl.operator_type AS operator_type
	 , (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint AS last_update
	 , $4 AS timeout_micro
	 , $5 AS present
	 , $6 AS metadata
FROM ns, ctl
;
`

const opcCheckInUpdate = `
-- $1 operator id (bigint)
-- $2 check in timeout (bigint micros)
-- $3 present (boolean)
-- $4 metadata (jsonb)
-- $5 controller namespace (text)
-- $6 controller type (int)
-- $7 controller name (text)
WITH ctl AS (
	WITH ctl_namespace AS (
		SELECT namespaces.id AS nsid
		FROM namespaces
		WHERE namespaces.name = $5
		UNION
		SELECT NULL AS nsid
		ORDER BY nsid NULLS LAST
		LIMIT 1
	)
	SELECT id AS id, operator_type AS operator_type
	FROM opc, ctl_namespace
	WHERE (opc.namespace = ctl_namespace.nsid OR opc.namespace IS NULL)
	  AND opc.operator_type = $6
	  AND opc.operator_name = $7
	UNION
	SELECT NULL AS id, 0 AS operator_type
	ORDER BY id NULLS LAST
	LIMIT 1
)
UPDATE opc SET (
    last_update
  , timeout_micro
  , present
  , metadata
  , controller
  , controller_type
) = (
	SELECT (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint
		 , $2::bigint
		 , $3::boolean
         , COALESCE($4::jsonb, '{}')
         , ctl.id
         , ctl.operator_type
	FROM ctl
)
WHERE id = $1
;
`

const opcCheckOut = `
WITH ns AS (
	(SELECT id AS id
	FROM namespaces
	WHERE namespaces.name = $1)
	UNION (SELECT null AS id)
	ORDER BY id NULLS LAST
	LIMIT 1
)
DELETE FROM opc
USING ns
WHERE (namespace = ns.id OR namespace IS NULL)
  AND operator_type = $2
  AND operator_name = $3
;
`

const opcGetNotifications = `
-- opcGetNotifications gets all of the notifications that are ready for a given
-- type of operator. Operators have notifications ready when they have not
-- checked in for longer than their timeout_micro.
--
-- $1: operator type
-- $2: controller name
-- $3: controller type
-- $4: controller namespace
SELECT namespaces.name
	 , opc.operator_type
	 , opc.operator_name
	 , false -- present is always logically false here, and we use the same target params as the get queries
	 , to_timestamp(opc.last_update::double precision / 1000000)
	 , opc.timeout_micro * 1000 -- nanoseconds for Go time.Duration
	 , opc.metadata
	 , opc.controller
FROM opc
   LEFT OUTER JOIN namespaces ON opc.namespace = namespaces.id
   LEFT OUTER JOIN opc AS controller_opc ON opc.controller = controller_opc.id
   LEFT OUTER JOIN namespaces AS controller_namespaces ON controller_opc.namespace = namespaces.id
WHERE opc.operator_type = $1
  AND (controller_opc.operator_name = $2 OR (controller_opc.operator_name IS NULL AND $2 = ''))
  AND (controller_opc.operator_type = $3 OR (controller_opc.operator_type IS NULL AND $3 = 0))
  AND (controller_namespaces.name = $4 OR (controller_opc.namespace IS NULL AND $4 = ''))
  AND opc.timeout_micro < (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint - opc.last_update
;
`

const opcGetGrandchildNotifications = `
-- opcGetGrandchildNotifications gets all of the notifications that are ready for a given
-- type of operator's grandchildren. Operators have notifications ready when they have not
-- checked in for longer than their timeout_micro.
--
-- $1: operator type
-- $2: controller name
-- $3: controller type
-- $4: controller namespace
WITH children AS (
	SELECT opc.id AS id
	FROM opc
	   LEFT OUTER JOIN namespaces ON opc.namespace = namespaces.id
	   LEFT OUTER JOIN opc AS controller_opc ON opc.controller = controller_opc.id
	   LEFT OUTER JOIN namespaces AS controller_namespaces ON controller_opc.namespace = namespaces.id
	WHERE (controller_opc.operator_name = $2 OR (controller_opc.operator_name IS NULL AND $2 = ''))
      AND (controller_opc.operator_type = $3 OR (controller_opc.operator_type IS NULL AND $3 = 0))
      AND (controller_namespaces.name = $4 OR (controller_opc.namespace IS NULL AND $4 = ''))
)
SELECT namespaces.name
	 , opc.operator_type
	 , opc.operator_name
	 , false -- present is always logically false here, and we use the same target params as the get queries
	 , to_timestamp(opc.last_update::double precision / 1000000)
	 , opc.timeout_micro * 1000 -- nanoseconds for Go time.Duration
	 , opc.metadata
	 , opc.controller
FROM opc, namespaces, children
WHERE opc.operator_type = $1
  AND opc.timeout_micro < (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint - opc.last_update
  AND namespaces.id = opc.namespace
  AND opc.controller = children.id
`

const opcUpdateNotifications = `
-- opcUpdateNotifications resets the last_update column for all operators
-- of a given type that are controlled by one of the controllers in the
-- provided set. It also sets present false. opcUpdateNotifications only
-- applies to operators that have not checked in on time.
--
-- $1 array of controller IDs
-- $2 operator type (int)
UPDATE opc
SET last_update = (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint
  , present = false
WHERE opc.controller = ANY(($1::bigint[]))
  AND opc.operator_type = $2
  AND timeout_micro < (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint - opc.last_update
;
`

const opcReassignAbsentControllers = `
-- opcReassignAbsentControllers reassigns controllers to operators who have
-- had their controller checked out or fail to check in. Operators will receive
-- a new controller of the same type if one is available, randomly selected
-- from the available controllers that are present.
UPDATE opc
SET controller = (
	SELECT copc.id AS controller FROM opc AS copc
	WHERE copc.operator_type = opc.controller_type
	  AND copc.id != opc.id
	  AND copc.timeout_micro > (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint - copc.last_update
	ORDER BY random()
	LIMIT 1
)
FROM opc AS copc
WHERE opc.controller_type > 0
  AND (opc.controller IS NULL OR (opc.controller = copc.id AND copc.timeout_micro < (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint - copc.last_update))
;
`

const opcGetOperator = `
-- opcGetOperator gets an operator by namespace, type and name.
--
-- $1: namespace (text)
-- $2: operator type (int)
-- $3: operator name (text)
SELECT COALESCE(namespaces.name, '')
     , opc.operator_type
     , opc.operator_name
	 , present
	 , to_timestamp(opc.last_update::double precision / 1000000)
	 , opc.timeout_micro * 1000 -- nanoseconds for Go time.Duration
	 , opc.metadata
	 , opc.controller
FROM opc LEFT OUTER JOIN namespaces ON opc.namespace = namespaces.id
WHERE ($1 = '' OR namespaces.name = $1)
  AND opc.operator_type = $2
  AND opc.operator_name = $3
;
`

const opcListOperators = `
-- opcGetOperator gets an operator by namespace, type and name.
--
-- $1: namespace (text)
-- $2: operator type (int)
-- $3: operator name (text)
SELECT COALESCE(namespaces.name, '')
     , opc.operator_type
     , opc.operator_name
	 , present
	 , to_timestamp(opc.last_update::double precision / 1000000)
	 , opc.timeout_micro * 1000 -- nanoseconds for Go time.Duration
	 , opc.metadata
	 , opc.controller
FROM opc LEFT OUTER JOIN namespaces ON opc.namespace = namespaces.id
WHERE ($1 = '' OR namespaces.name = $1)
  AND ($2 = 0 OR opc.operator_type = $2)
  AND ($3 = '' OR opc.operator_name = $3)
;
`

const opcGetOperatorByID = `
-- opcGetOperatorByID is like opcGetOperator but works by operator ID
--
-- $1: operator ID (bigint)
SELECT COALESCE(namespaces.name, '')
     , opc.operator_type
     , opc.operator_name
	 , present
	 , to_timestamp(opc.last_update::double precision / 1000000)
	 , opc.timeout_micro * 1000 -- nanoseconds for Go time.Duration
	 , opc.metadata
	 , opc.controller
FROM opc LEFT OUTER JOIN namespaces ON opc.namespace = namespaces.id
WHERE opc.id = $1
`

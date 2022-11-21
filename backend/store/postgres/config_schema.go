package postgres

import (
	_ "github.com/lib/pq"
)

// ConfigurationDDL is the postgresql data definition language that defines the
// generic resource table schema.
const ConfigurationDDL = `CREATE TABLE IF NOT EXISTS configuration (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    api_version         text NOT NULL,
    api_type            text NOT NULL,
    namespace           text NOT NULL,
    name                text NOT NULL,
    labels              jsonb NOT NULL,
    annotations         bytea NOT NULL,
    resource            jsonb NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    deleted_at          timestamptz NOT NULL DEFAULT '-infinity',
    CONSTRAINT configuration_unique UNIQUE(api_version, api_type, namespace, name, deleted_at)
);

CREATE OR REPLACE FUNCTION refresh_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER configuration_updated_at BEFORE UPDATE
    ON configuration FOR EACH ROW EXECUTE PROCEDURE
    refresh_updated_at_column();
`

const CreateConfigIfNotExistsQuery = `
INSERT INTO configuration ( api_version
						  , api_type
						  , namespace
						  , name
						  , labels
						  , annotations
						  , resource
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING;`

const CreateOrUpdateConfigQuery = `INSERT INTO configuration (api_version, api_type, namespace, name, labels, annotations, resource)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (api_version, api_type, namespace, name, deleted_at) DO UPDATE
      SET labels=$5, annotations=$6, resource=$7;`

const DeleteConfigQuery = `UPDATE configuration SET deleted_at=NOW()
    WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const ExistsConfigQuery = `SELECT count(*) AS total FROM configuration
    WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const GetConfigQuery = `SELECT id, labels, annotations, resource, created_at, updated_at FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const CountConfigQueryTmpl = `
    SELECT count(*)
    FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND NOT isfinite(deleted_at)
        {{if ne .SelectorSQL ""}}AND {{.SelectorSQL}}{{end}};`

const ListConfigQueryTmpl = `
    SELECT id, labels, annotations, resource, created_at, updated_at
    FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND NOT isfinite(deleted_at)
        {{if ne .SelectorSQL ""}}AND {{.SelectorSQL}}{{end}}
    ORDER BY namespace, name ASC
    {{if (gt .Limit 0)}} LIMIT {{.Limit}} {{end}} OFFSET {{ .Offset }};`

const UpdateIfExistsConfigQuery = `UPDATE configuration SET labels=$1, annotations=$2, resource=$3
    WHERE api_version=$4 AND api_type=$5 AND namespace=$6 AND name=$7 AND NOT isfinite(deleted_at);`

const ConfigNotificationQuery = `
    SELECT id, api_version, api_type, namespace, name, resource, created_at, updated_at,
        CASE WHEN isfinite(deleted_at) THEN deleted_at ELSE NULL END AS deleted_at
    FROM Configuration
    WHERE api_version=$1 AND api_type=$2 AND (updated_at > $3 OR created_at > $3 OR (isfinite(deleted_at) AND deleted_at > $3));`

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
						  , fields
						  , resource
						  , etag
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, digest($8::jsonb::text, 'sha1'))
RETURNING etag;`

const CreateOrUpdateConfigQuery = `
INSERT INTO configuration (api_version, api_type, namespace, name, labels, annotations, fields, resource, etag)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, digest($8::jsonb::text, 'sha1'))
    ON CONFLICT (api_version, api_type, namespace, name, deleted_at) DO UPDATE
	  SET labels=$5, annotations=$6, fields=$7, resource=$8, etag=digest($8::jsonb::text, 'sha1')
	RETURNING xmax = 0, (SELECT cfg.etag FROM configuration cfg WHERE cfg.id = configuration.id), digest($8::jsonb::text, 'sha1')
;`

const DeleteConfigQuery = `UPDATE configuration SET deleted_at=NOW()
    WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const DeleteConfigIfMatchQuery = `UPDATE configuration SET deleted_at=NOW()
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at) AND etag = ANY(($5::bytea[]));`

const DeleteConfigIfNoneMatchQuery = `UPDATE configuration SET deleted_at=NOW()
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at) AND NOT etag = ANY(($5::bytea[]));`

const ExistsConfigQuery = `SELECT count(*) AS total FROM configuration
    WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const ExistsConfigIfMatchQuery = `SELECT count(*) AS total FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at) AND etag = ANY(($5::bytea[]));`

const ExistsConfigIfNoneMatchQuery = `SELECT count(*) AS total FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at) AND NOT etag = ANY(($5::bytea[]));`

const GetETagOnlyQuery = `SELECT etag FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const GetConfigQuery = `SELECT id, labels, annotations, resource, created_at, updated_at, etag FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND namespace=$3 AND name=$4 AND NOT isfinite(deleted_at);`

const CountConfigQueryTmpl = `
	{{if not .Namespaced}} WITH ignored as (SELECT $3::text) {{end}}
    SELECT count(*)
    FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND NOT isfinite(deleted_at)
        {{if .Namespaced}} AND namespace=$3 {{end}}
        {{if ne .SelectorSQL ""}}AND {{.SelectorSQL}}{{end}};`

const ListConfigQueryTmpl = `
	{{if not .Namespaced}} WITH ignored as (SELECT $3::text) {{end}}
    SELECT id, labels, annotations, resource, created_at, updated_at, NULLIF(deleted_at, '-infinity'), etag
    FROM configuration
	WHERE api_version=$1 AND api_type=$2 AND ($5 OR NOT isfinite(deleted_at))
		AND updated_at > $4
        {{if .Namespaced}} AND namespace=$3 {{end}}
        {{if ne .SelectorSQL ""}}AND {{.SelectorSQL}}{{end}}
    ORDER BY namespace, name ASC
    {{if (gt .Limit 0)}} LIMIT {{.Limit}} {{end}} OFFSET {{ .Offset }};`

const UpdateIfExistsConfigQuery = `
	UPDATE configuration AS old SET labels=$1, annotations=$2, fields=$3, resource=$4, etag=digest($4::jsonb::text, 'sha1')
    WHERE api_version=$5 AND api_type=$6 AND namespace=$7 AND name=$8 AND NOT isfinite(deleted_at)
    RETURNING etag;`

const ConfigNotificationQuery = `
    SELECT id, api_version, api_type, namespace, name, resource, created_at, updated_at,
        CASE WHEN isfinite(deleted_at) THEN deleted_at ELSE NULL END AS deleted_at
    FROM Configuration
    WHERE api_version=$1 AND api_type=$2 AND (updated_at > $3 OR created_at > $3 OR (isfinite(deleted_at) AND deleted_at > $3));`

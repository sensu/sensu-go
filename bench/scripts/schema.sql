DROP TABLE IF EXISTS counters;
DROP TABLE IF EXISTS configuration;
DROP TABLE IF EXISTS silenced;

DROP FUNCTION IF EXISTS trigger_set_updated_at();
CREATE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE counters (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    c                   bigint NOT NULL DEFAULT 0,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    deleted_at          timestamptz
);

DROP TRIGGER IF EXISTS set_updated_at ON counters;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON counters
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_updated_at();

CREATE TABLE configuration (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    api_version         text NOT NULL,
    type                text NOT NULL,
    namespace           text NOT NULL,
    name                text NOT NULL,
    labels              jsonb NOT NULL,
    annotations         bytea NOT NULL,
    resource            jsonb   NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT NOW(),
    updated_at       timestamptz NOT NULL DEFAULT NOW(),
    deleted_at        timestamptz,
    CONSTRAINT resource_unique UNIQUE(api_version, type, namespace, name)
);

DROP TRIGGER IF EXISTS set_updated_at ON configuration;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON configuration
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_updated_at();

CREATE TABLE silenced (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    begin_at        timestamptz,
    expire_at    timestamptz,
    expire_on_resolve   bool NOT NULL,
    namespace           text NOT NULL,
    name                text NOT NULL,
    creator             text,
    check_name          text,
    subscription        text,
    reason              text,
    CONSTRAINT sub_check_unique UNIQUE(namespace, check_name, subscription)
);


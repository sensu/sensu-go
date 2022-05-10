
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE IF NOT EXISTS counters (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    c                   bigint NOT NULL DEFAULT 0,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    deleted_at          timestamptz
);

CREATE OR REPLACE TRIGGER set_updated_at
BEFORE UPDATE ON counters
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_updated_at();

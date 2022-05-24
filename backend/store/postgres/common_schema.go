package postgres

const migrateRefreshUpdatedAtProcedure = `
-- This is a procedure that can be used to set the "updated_at" column to the
-- current time on a given table. It is intended to be used with update triggers
-- to ensure that the "updated_at" column represents the last time a row was
-- updated.
CREATE OR REPLACE FUNCTION refresh_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ language 'plpgsql';
`

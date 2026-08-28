-- Filename: 000005_create_updated_at_trigger.up.sql
BEGIN;

-- set_updated_at sets the updated_at column to the current timestamp.
CREATE
OR REPLACE FUNCTION set_updated_at () RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- consumers_updated_at applies the set_updated_at function to the consumers table.
CREATE TRIGGER consumers_updated_at BEFORE
UPDATE ON consumers FOR EACH ROW
EXECUTE FUNCTION set_updated_at ();

COMMIT;
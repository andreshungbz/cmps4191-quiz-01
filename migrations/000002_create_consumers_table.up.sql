-- Filename: 000002_create_consumers_table.up.sql
BEGIN;

CREATE TABLE IF NOT EXISTS
    consumers (
        id uuid PRIMARY KEY DEFAULT uuidv7 (),
        name text NOT NULL,
        email citext NOT NULL UNIQUE,
        -- A newly created consumer is active by default.
        status consumer_status NOT NULL DEFAULT 'active',
        version integer NOT NULL DEFAULT 1,
        -- Timestamp fields for keeping track of updates to the consumer record.
        created_at timestamptz NOT NULL DEFAULT now(),
        updated_at timestamptz NOT NULL DEFAULT now()
    );

COMMIT;
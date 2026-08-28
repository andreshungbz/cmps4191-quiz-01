-- Filename: 000003_create_api_keys_table.up.sql
BEGIN;

CREATE TABLE IF NOT EXISTS
    api_keys (
        id uuid PRIMARY KEY DEFAULT uuidv7 (),
        /*
        Every API key is mandatorily associated with a consumer.
        If a consumer is deleted, all of its API keys are also deleted.
         */
        consumer_id uuid NOT NULL REFERENCES consumers (id) ON DELETE CASCADE,
        key_hash text NOT NULL UNIQUE,
        key_prefix text NOT NULL,
        -- A newly created API key is active by default.
        status key_status NOT NULL DEFAULT 'active',
        -- Timestamp fields for keeping track of API key usage and expiration.
        last_used_at timestamptz,
        expires_at timestamptz,
        created_at timestamptz NOT NULL DEFAULT now()
    );

COMMIT;
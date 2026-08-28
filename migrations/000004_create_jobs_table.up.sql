-- Filename: 000004_create_jobs_table.up.sql
BEGIN;

CREATE TABLE IF NOT EXISTS
    jobs (
        id uuid PRIMARY KEY DEFAULT uuidv7 (),
        -- Q16: A foreign-key violation reveals that consumer associated with the job does not exist in the database,
        -- which is a data integrity issue. This is why the consumer_id column is a foreign key that references the consumers table.
        consumer_id uuid NOT NULL REFERENCES consumers (id),
        job_type text NOT NULL,
        status job_status NOT NULL DEFAULT 'queued',
        payload jsonb NOT NULL DEFAULT '{}',
        result jsonb,
        error_message text,
        started_at timestamptz,
        completed_at timestamptz,
        created_at timestamptz NOT NULL DEFAULT now()
    );

COMMIT;
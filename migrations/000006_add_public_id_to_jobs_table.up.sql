-- Filename: 000006_add_public_id_to_jobs_table.up.sql
BEGIN;

/*
A job public_id is what the client uses to poll for the status of a job. It is a UUIDv4, 
which doesn't have a timestamp component like UUIDv7. A job having both a UUIDv7 id and
a UUIDv4 public_id allows both the benefit of quick indexing and security.
 */
ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS public_id UUID NOT NULL DEFAULT uuidv4 ();

COMMIT;
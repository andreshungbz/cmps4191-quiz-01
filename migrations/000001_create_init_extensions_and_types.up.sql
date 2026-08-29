-- Filename: 000001_create_init_extensions_and_types.up.sql
BEGIN;

/*
citext stands for case-insensitive text, which is useful for email addresses and 
other identifiers that should not be case-sensitive. 

NOTE: The creation of the citext extension in the database was moved from the 
starter code documented imperative command to here for convenience.
 */
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE consumer_status AS ENUM('active', 'suspended', 'terminated');

CREATE TYPE key_status AS ENUM('active', 'rotating', 'revoked');

/*
Q10: The queued, processing, completed, and failed statuses communicate the different stages
of the job lifecycle by representing each distinct phase a job can be in. When initially created,
a job is in the queued state. When the background worker picks up the job, it transitions to
the processing state. From there, it can either transition to the completed state if the job
processing succeeds, or to the failed state if an error occurs during processing. Although not
shown in the implementation, a job can also be cancelled by the consumer.
 */
CREATE TYPE job_status AS ENUM(
    'queued',
    'processing',
    'completed',
    'failed',
    'cancelled'
);

COMMIT;
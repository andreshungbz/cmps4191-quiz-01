package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// ReportPayload represents the expected payload structure for a consumer activity report job.
type ReportPayload struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Job represents a record in the jobs table of the database.
//
// Q11: The fields of the struct mirrors the columns of the jobs table in the database.
//
// Q12: Some fields are pointers because they can be NULL in the database. A missing timestamp
// or error message would mean would mean those values are NULL in the database,
// and thus should be represented as nil in Go.
type Job struct {
	ID           string          `json:"-"`
	PublicID     string          `json:"id"`
	ConsumerID   string          `json:"consumer_id"`
	JobType      string          `json:"job_type"`
	Status       string          `json:"status"`
	Payload      ReportPayload   `json:"payload"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// JobModel wraps a sql.DB connection pool used to interact with the database.
type JobModel struct {
	DB *sql.DB
}

// Insert writes a new job record to the database.
func (m JobModel) Insert(job *Job) error {
	// Q13: The purpose of doing json.Marshal(job.Payload) before inserting the job into the database
	// is because the database expects the payload to be in JSON format, so we need to serialize the
	// Go struct into JSON before inserting it.
	payload, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	// Construct the query and context.
	//
	// Q15: RETURNING in the query populates the Go job struct immediately after the insert by
	// chaining the Scan method on the QueryRowContext method which returns the row.
	query := `INSERT INTO jobs (consumer_id, job_type, payload)
		VALUES ($1, $2, $3) RETURNING id, public_id, status, created_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query and scan the returned values into the passed job struct,
	// handling consumer ID foreign key violations (non-existent consumer) and
	// other errors as a catch-all.
	err = m.DB.QueryRowContext(ctx, query, job.ConsumerID, job.JobType, payload).Scan(
		&job.ID, &job.PublicID, &job.Status, &job.CreatedAt,
	)
	if err != nil {
		var pgErr *pq.Error
		// 23503 is the PostgreSQL error code for foreign_key_violation.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

// GetByPublicID reads a job record from the database based on the provided public ID.
func (m JobModel) GetByPublicID(publicID string) (*Job, error) {
	// Construct the query and context.
	query := `SELECT id, public_id, consumer_id, job_type, status, payload,
		COALESCE(result, 'null'::jsonb), error_message, started_at, completed_at, created_at
		FROM jobs WHERE public_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query and scan the returned values into a new job struct,
	// handling missing records and other errors as a catch-all. The payload field
	// needs to be unmarshaled from JSON since ReportPayload is a Go struct. On the other
	// hand, the result field is already a json.RawMessage type, so it can be scanned directly.
	var job Job
	var payload []byte
	err := m.DB.QueryRowContext(ctx, query, publicID).Scan(&job.ID, &job.PublicID,
		&job.ConsumerID, &job.JobType, &job.Status, &payload, &job.Result,
		&job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(payload, &job.Payload); err != nil {
		return nil, err
	}

	return &job, nil
}

// ClaimNext retrieves the next queued job of type "consumer_activity_report" from the database.
func (m JobModel) ClaimNext(ctx context.Context) (*Job, error) {
	// Q17: ClaimNext beigns a database transaction before selecting a job because it needs to
	// lock the selected job record to prevent other workers from claiming it at the same time.
	// The transaction ensures that the job is marked as "processing" only if it is successfully claimed,
	// and if any error occurs during the process, the transaction can be rolled back to maintain data integrity.
	//
	// Q21: Selecting a queued job and updating it to processing should occur in the same transaction
	// to ensure that the job is not claimed by another worker before it is marked as processing.
	// If that happens, it could lead to multiple workers processing the same job, which would be a
	// race condition and could cause data inconsistencies or duplicate work.
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Construct the query.
	//
	// Q18: The effect of the WHERE clause is to select only jobs that are in the "queued" status
	// and are of type "consumer_activity_report".
	//
	// Q19: The ORDER BY created_at clause ensures that the oldest queued job is the first row,
	// following a first-in-first-out (FIFO) approach. This is so that old jobs don't get stuck.
	// LIMIT 1 guarantees only one job is claimed at a time and that it is the oldest job.
	//
	// Q20: FOR UPDATE causes the selected row to be locked for the duration of the transaction,
	// preventing other workers or transactions from claiming it simultaneously. SKIP LOCKED matters
	// when multiple workers exist, as it allows them to skip over locked rows and claim the next
	// available job without waiting for the lock to be released.
	query := `SELECT id, public_id, consumer_id, job_type, payload FROM jobs
		WHERE status = 'queued' AND job_type = 'consumer_activity_report'
		ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`

	// Execute the query, scan the returned values into a new job struct, update job status to 'processing',
	// and handle other errors as a catch-all. The payload field needs to be unmarshaled
	// from JSON since ReportPayload is a Go struct.
	var job Job
	var payload []byte
	// Q22: If no queued jobs are available, the UPDATE query is not executed and the transaction
	// is rolled back. See worker.go for how the sql.ErrNoRows error is handled in the calling function.
	if err := tx.QueryRowContext(ctx, query).Scan(&job.ID, &job.PublicID,
		&job.ConsumerID, &job.JobType, &payload); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, &job.Payload); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE jobs SET status = 'processing', started_at = now() WHERE id = $1`, job.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	job.Status = "processing"

	return &job, nil
}

// MarkCompleted updates the status of a job to "completed" in the database and sets its result.
func (m JobModel) MarkCompleted(ctx context.Context, id string, result []byte) error {
	// Q41: MarkCompleted updates the status, result, and completed_at columns of a jobs table record.
	_, err := m.DB.ExecContext(ctx,
		`UPDATE jobs SET status = 'completed', result = $2, completed_at = now() WHERE id = $1`,
		id, result)
	return err
}

// MarkFailed updates the status of a job to "failed" in the database and sets its error message.
func (m JobModel) MarkFailed(ctx context.Context, id, message string) error {
	// Q41: MarkFailed updates the status, error_message, and completed_at columns of a jobs table record.
	_, err := m.DB.ExecContext(ctx,
		`UPDATE jobs SET status = 'failed', error_message = $2, completed_at = now() WHERE id = $1`,
		id, message)
	return err
}

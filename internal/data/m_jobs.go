package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

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

type JobModel struct {
	DB *sql.DB
}

func (m JobModel) Insert(job *Job) error {
	// Q13: The purpose of doing json.Marshal(job.Payload) before inserting the job into the database
	// is because the database expects the payload to be in JSON format, so we need to serialize the
	// Go struct into JSON before inserting it.
	payload, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	// Q15: RETURNING in the query populates the Go job struct immediatly after the insert by
	// immediately calling Scan on the returned row.
	query := `INSERT INTO jobs (consumer_id, job_type, payload)
		VALUES ($1, $2, $3) RETURNING id, public_id, status, created_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, job.ConsumerID, job.JobType, payload).Scan(
		&job.ID, &job.PublicID, &job.Status, &job.CreatedAt,
	)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m JobModel) GetByPublicID(publicID string) (*Job, error) {
	query := `SELECT id, public_id, consumer_id, job_type, status, payload,
		COALESCE(result, 'null'::jsonb), error_message, started_at, completed_at, created_at
		FROM jobs WHERE public_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
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

func (m JobModel) ClaimNext(ctx context.Context) (*Job, error) {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `SELECT id, public_id, consumer_id, job_type, payload FROM jobs
		WHERE status = 'queued' AND job_type = 'consumer_activity_report'
		ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`
	var job Job
	var payload []byte
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

func (m JobModel) MarkCompleted(ctx context.Context, id string, result []byte) error {
	_, err := m.DB.ExecContext(ctx,
		`UPDATE jobs SET status = 'completed', result = $2, completed_at = now() WHERE id = $1`,
		id, result)
	return err
}

func (m JobModel) MarkFailed(ctx context.Context, id, message string) error {
	_, err := m.DB.ExecContext(ctx,
		`UPDATE jobs SET status = 'failed', error_message = $2, completed_at = now() WHERE id = $1`,
		id, message)
	return err
}

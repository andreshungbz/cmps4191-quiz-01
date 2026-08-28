package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ConsumerActivityReport represents a report of a consumer's activity within a specified time range.
type ConsumerActivityReport struct {
	ConsumerID     string    `json:"consumer_id"`
	ConsumerName   string    `json:"consumer_name"`
	From           time.Time `json:"from"`
	To             time.Time `json:"to"`
	ActiveKeys     int       `json:"active_keys"`
	RevokedKeys    int       `json:"revoked_keys"`
	QueuedJobs     int       `json:"queued_jobs"`
	ProcessingJobs int       `json:"processing_jobs"`
	CompletedJobs  int       `json:"completed_jobs"`
	FailedJobs     int       `json:"failed_jobs"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// ReportModel wraps a sql.DB connection pool used to interact with the database.
type ReportModel struct {
	DB *sql.DB
}

// Generate creates a ConsumerActivityReport for the specified consumer ID and time range.
func (m ReportModel) Generate(consumerID string, from, to time.Time) (*ConsumerActivityReport, error) {
	// Construct the query and context.
	//
	// Q35: The report query starts from the consumers table and then uses LEFT (OUTER) JOINs
	// in order to include all consumers, even those without any associated API keys or jobs.
	// LEFT JOINs always include all records from the left table and only matching records from
	// the right tables (api_keys and jobs).
	//
	// Q36: Because LEFT JOINs for api_keys and jobs are being done independently on consumers,
	// every result of the rows of consumer x api_keys will form the Cartesian product with jobs,
	// which will result in double counting of jobs and API keys. The COUNT(DISTINCT ...)
	// deduplicates the keys and jobs to prevent inflated counts.
	//
	// Q37: FILTER (WHERE ...) is used to count only the rows that match the specified status
	// for each of the API keys and jobs, as the report requires counts of each status type. It
	// produces separate totals for each key or job status by filtering individually for each
	// COUNT(DISTINCT ...) aggregate function.
	//
	// Q38: The j.created_at >= $2 AND j.created_at < $3 condition only uses one inclusive bound
	// because having that inclusive bound can cause double counting of jobs that are created at
	// the exact same time as the end of the report's time range. For example, if a report is
	// generated with a created_at of exactly 2026-06-01 00:00:00, it would appear in both queries
	// of ranges 2026-05-01 00:00:00 to 2026-06-01 00:00:00 and 2026-06-01 00:00:00 to 2026-07-01 00:00:00.
	//
	// Q39: The GROUP BY c.id, c.name ensures that the query result is consolidated into a single row
	// for the specified consumer, as the report is only for one consumer.
	query := `
		SELECT c.id, c.name,
			COUNT(DISTINCT k.id) FILTER (WHERE k.status = 'active'),
			COUNT(DISTINCT k.id) FILTER (WHERE k.status = 'revoked'),
			COUNT(DISTINCT j.id) FILTER (WHERE j.status = 'queued'),
			COUNT(DISTINCT j.id) FILTER (WHERE j.status = 'processing'),
			COUNT(DISTINCT j.id) FILTER (WHERE j.status = 'completed'),
			COUNT(DISTINCT j.id) FILTER (WHERE j.status = 'failed')
		FROM consumers c
		LEFT JOIN api_keys k ON k.consumer_id = c.id
		LEFT JOIN jobs j ON j.consumer_id = c.id
			AND j.created_at >= $2 AND j.created_at < $3
		WHERE c.id = $1
		GROUP BY c.id, c.name`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the query and scan the returned values into a new ConsumerActivityReport struct,
	// handling missing records and other errors as a catch-all.
	//
	// Q39: QueryRowContext(...).Scan(...) executues the query given the context and argumants,
	// and then populates the resulting record into a destination which is then used elsewhere
	// in the Go code.
	report := &ConsumerActivityReport{From: from, To: to, GeneratedAt: time.Now()}
	err := m.DB.QueryRowContext(ctx, query, consumerID, from, to).Scan(
		&report.ConsumerID, &report.ConsumerName, &report.ActiveKeys, &report.RevokedKeys,
		&report.QueuedJobs, &report.ProcessingJobs, &report.CompletedJobs, &report.FailedJobs,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return report, nil
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// startReportWorker starts a background goroutine that polls the jobs table for new report jobs to process.
func (app *application) startReportWorker(ctx context.Context) {
	// Q25: The WaitGroup task counter is incremented before starting the goroutine to ensure that
	// the main goroutine waits for the report worker to finish before exiting. The worker calls
	// Done to decrement this counter when it exits.
	app.wg.Add(1)

	// Q26: go func() launches an anonymous goroutine that runs concurrently with the main goroutine.
	// It is the report worker that polls for the next job to process at the configured interval.
	// app.wg.Done() is deferred to ensure that the WaitGroup counter is decremented when once the
	// goroutine exits, allowing the main goroutine to proceed with shutdown.
	go func() {
		defer app.wg.Done()

		// Q27: The ticker is a Ticker struct that contains a channel which sends the current time
		// on the channel every configured interval duration, which is currently 250ms.
		//
		// Q30: defer ticker.Stop() is appropriate when the worker exits because it will prevent
		// additional ticks from being sent to the channel, which would otherwise cause a memory leak.
		ticker := time.NewTicker(app.config.workerPollInterval)
		defer ticker.Stop()

		// Until shutdown, process the next report job at the configured interval.
		//
		// Q28: Between ctx.Done() and ticker.C, the select statement chooses whichever channel is
		// ready first (i.e., channel close for ctx.Done() or received a value for ticker.C).
		// If both happen to be ready at the same time, one case is randomly chosen.
		for {
			select {
			// Q33: Cancellation (WithCancel arranges for Done to be closed) allows the worker to stop
			// before the timer naturally expires as the select statement will choose that ready case.
			// A closed channel is already considered ready for reading.
			case <-ctx.Done():
				app.logger.Info("report worker stopped")
				return
			case <-ticker.C:
				// Q22: If no queued jobs are available, the sql.ErrNoRows error ultimately received
				// here is not considered an actual worker failure, and therefore the log below does
				// not execute and the worker continues to poll for the next job at the configured interval.
				err := app.processNextReportJob(ctx)
				if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, context.Canceled) {
					app.logger.Error("report worker failed", "error", err)
				}
			}
		}
	}()
}

// processNextReportJob claims and processes the next report job from the database.
func (app *application) processNextReportJob(ctx context.Context) error {
	// Claim and log the next report job, handling errors.
	job, err := app.models.Jobs.ClaimNext(ctx)
	if err != nil {
		return err
	}
	app.logger.Info("report job started", "job_id", job.PublicID,
		"artificial_delay", app.config.reportDelay)

	// Apply the artificial report delay if configured to be greater than 0.
	//
	// Q31: Here is where reportDelay is being applied. It has to occur in this function,
	// which is called by the goroutine launched by startReportWorker rather than in the POST handler,
	// because otherwise the HTTP request would be synchronously blocked for the duration of the delay.
	if app.config.reportDelay > 0 {
		timer := time.NewTimer(app.config.reportDelay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	// Generate the report, marking and logging the job as completed, or marking it as
	// failed if an error occurs.
	//
	// Q40: The generated report is marshaled into JSON before MarkCompleted stores it because
	// the report is a Go struct and the database expects the JSON type. In particular,
	// the result field of a job record is of type JSONB.
	report, err := app.models.Reports.Generate(job.ConsumerID, job.Payload.From, job.Payload.To)
	if err != nil {
		return app.models.Jobs.MarkFailed(ctx, job.ID, err.Error())
	}
	result, err := json.Marshal(report)
	if err != nil {
		return app.models.Jobs.MarkFailed(ctx, job.ID, err.Error())
	}
	if err := app.models.Jobs.MarkCompleted(ctx, job.ID, result); err != nil {
		return err
	}
	app.logger.Info("report job completed", "job_id", job.PublicID)

	return nil
}

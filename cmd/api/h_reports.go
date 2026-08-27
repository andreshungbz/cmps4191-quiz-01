package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/andreshungbz/cmps4191-quiz-01/internal/data"
	"github.com/andreshungbz/cmps4191-quiz-01/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// createReportHandler creates a new consumer_activity_report job in the database.
func (app *application) createReportHandler(w http.ResponseWriter, r *http.Request) {
	// Attempt to read valid JSON input from the request and store it in an interim struct,
	// returning a 400 response if the JSON is invalid.
	var input struct {
		ConsumerID string    `json:"consumer_id"`
		From       time.Time `json:"from"`
		To         time.Time `json:"to"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the input fields for the job, returning a 422 response if any validation fails.
	// NOTE: These validation checks can be abstracted into a ValidateJob method in the
	// validator.Validator struct.
	v := validator.New()
	v.Check(input.ConsumerID != "", "consumer_id", "must be provided")
	v.Check(!input.From.IsZero(), "from", "must be provided")
	v.Check(!input.To.IsZero(), "to", "must be provided")
	v.Check(input.From.Before(input.To), "from", "must be earlier than to")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Copy input fields to a job struct and insert it into the database, handling errors.
	// A data.ErrRecordNotFound error can occur if the provided consumer_id does not exist in
	// the database, as it is a foreign key constraint on the jobs table.
	job := &data.Job{
		ConsumerID: input.ConsumerID,
		JobType:    "consumer_activity_report",
		Payload:    data.ReportPayload{From: input.From, To: input.To},
	}
	// NOTE: This error handling's form is different compared to that in createConsumerHandler,
	// the latter which uses a switch construct. However, they are both functionally the same.
	if err := app.models.Jobs.Insert(job); err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	// Add a Location header to the response with the new job's public ID (UUIDv4).
	statusURL := fmt.Sprintf("/v1/jobs/%s", job.PublicID)
	headers := make(http.Header)
	headers.Set("Location", statusURL)

	// Send a JSON response of the newly created job, handling any errors.
	response := envelope{"job_id": job.PublicID, "status": job.Status, "status_url": statusURL}
	if err := app.writeJSON(w, http.StatusAccepted, response, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getJobHandler(w http.ResponseWriter, r *http.Request) {
	job, err := app.models.Jobs.GetByPublicID((httprouter.ParamsFromContext(r.Context())).ByName("id")) // httprouter package version
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"job": job}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

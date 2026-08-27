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

func (app *application) createReportHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ConsumerID string    `json:"consumer_id"`
		From       time.Time `json:"from"`
		To         time.Time `json:"to"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.ConsumerID != "", "consumer_id", "must be provided")
	v.Check(!input.From.IsZero(), "from", "must be provided")
	v.Check(!input.To.IsZero(), "to", "must be provided")
	v.Check(input.From.Before(input.To), "from", "must be earlier than to")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	job := &data.Job{
		ConsumerID: input.ConsumerID,
		JobType:    "consumer_activity_report",
		Payload:    data.ReportPayload{From: input.From, To: input.To},
	}
	if err := app.models.Jobs.Insert(job); err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	statusURL := fmt.Sprintf("/v1/jobs/%s", job.PublicID)
	headers := make(http.Header)
	headers.Set("Location", statusURL)
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

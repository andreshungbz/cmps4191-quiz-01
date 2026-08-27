package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/andreshungbz/cmps4191-quiz-01/internal/data"
	"github.com/andreshungbz/cmps4191-quiz-01/internal/validator"
)

// createConsumerHandler creates a new consumer in the database.
func (app *application) createConsumerHandler(w http.ResponseWriter, r *http.Request) {
	// Attempt to read valid JSON input from the request and store it in an interim struct,
	// returning a 400 response if the JSON is invalid.
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy input fields to a consumer struct and validate the consumer,
	// returning a 422 response if the validation fails.
	consumer := &data.Consumer{
		Name:  input.Name,
		Email: input.Email,
	}
	v := validator.New()
	if data.ValidateConsumer(v, consumer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the consumer into the database, handling duplicate email errors
	// and other errors as a catch-all.
	err = app.models.Consumers.Insert(consumer)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Add a Location header to the response with the new consumer's ID (UUIDv7).
	// Returning a UUIDv7 is fine here because consumers are not created as frequently as
	// something such as jobs, so the security risk of exposing the timestamp component of
	// a UUIDv7 is not as significant. Endpoints for consumers also tend to not be public-facing.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/consumers/%s", consumer.ID))

	// Send a JSON response of the newly created consumer, handling any errors.
	err = app.writeJSON(w, http.StatusCreated, envelope{"consumer": consumer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

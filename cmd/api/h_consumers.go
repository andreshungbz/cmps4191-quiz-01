package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/andreshungbz/cmps4191-quiz-01/internal/data"
	"github.com/andreshungbz/cmps4191-quiz-01/internal/validator"
)

func (app *application) createConsumerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	consumer := &data.Consumer{
		Name:  input.Name,
		Email: input.Email,
	}

	v := validator.New()
	// validate the consumer struct and return a response if any checks fail
	if data.ValidateConsumer(v, consumer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the consumer into the database and handle any errors
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

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/consumers/%s", consumer.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"consumer": consumer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

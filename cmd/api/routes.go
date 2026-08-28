package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes returns the HTTP router configured with all handlers, route-specific middleware, and global middleware.
func (app *application) routes() http.Handler {
	// NOTE: The github.com/julienschmidt/httprouter package was used instead of mux := http.NewServeMux()
	// as a different implementation. However, the application architecture remains the same.
	router := httprouter.New()

	// BACKEND

	// Standard routes
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// DATABASE SCHEMA ROUTES

	// Consumer routes
	router.HandlerFunc(http.MethodPost, "/v1/consumers", app.createConsumerHandler)

	// Report routes
	router.HandlerFunc(http.MethodPost, "/v1/reports", app.createReportHandler)

	// Job routes
	// NOTE: Path slug differs from the starter code "/v1/jobs/{id}" because of the
	// github.com/julienschmidt/httprouter package being used.
	//
	// Q09: This GET endpoint differs from the POST endpoint in that it retrieves a job
	// from the database by its public ID, instead of creating a new job.
	router.HandlerFunc(http.MethodGet, "/v1/jobs/:id", app.getJobHandler)

	// GLOBAL MIDDLEWARE

	return app.requestLogger(router)
}

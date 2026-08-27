package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes returns the HTTP router configured with all handlers, route-specific middleware,
// and global middleware.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// BACKEND

	// Defined handlers for 404 and 205 status code
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Healthcheck route
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// DATABASE SCHEMA ROUTES

	// consumer routes
	router.HandlerFunc(http.MethodPost, "/v1/consumers", app.createConsumerHandler)

	// report routes
	router.HandlerFunc(http.MethodPost, "/v1/reports", app.createReportHandler)

	// job routes
	router.HandlerFunc(http.MethodGet, "/v1/jobs/:id", app.getJobHandler)

	// GLOBAL MIDDLEWARE

	return app.requestLogger(
		router,
	)
}

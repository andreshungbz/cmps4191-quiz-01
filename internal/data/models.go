package data

import (
	"database/sql"
	"errors"
)

// Errors
var ErrDuplicateEmail = errors.New("duplicate email address")
var ErrRecordNotFound = errors.New("no record found")
var ErrEditConflict = errors.New("edit conflict")

// Models is a wrapper struct that holds references to the different model types.
type Models struct {
	Consumers ConsumerModel
	Reports   ReportModel
	Jobs      JobModel
}

// NewModels initializes the Models struct with the provided database connection.
func NewModels(db *sql.DB) Models {
	return Models{
		Consumers: ConsumerModel{DB: db},
		Reports:   ReportModel{DB: db},
		Jobs:      JobModel{DB: db},
	}
}

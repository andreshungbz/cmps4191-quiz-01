package data

import (
	"database/sql"
)

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

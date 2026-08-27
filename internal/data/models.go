package data

import (
	"database/sql"
)

type Models struct {
	Consumers ConsumerModel
	Reports   ReportModel
	Jobs      JobModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Consumers: ConsumerModel{DB: db},
		Reports:   ReportModel{DB: db},
		Jobs:      JobModel{DB: db},
	}
}

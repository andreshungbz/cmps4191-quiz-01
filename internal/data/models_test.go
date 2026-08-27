package data

import (
	"database/sql"
	"testing"
)

// TestNewModelsWiresDB verifies each model receives the provided DB handle.
func TestNewModelsWiresDB(t *testing.T) {
	db := &sql.DB{}
	models := NewModels(db)

	if models.Consumers.DB != db {
		t.Fatal("Consumer model DB was not wired correctly")
	}

	if models.Jobs.DB != db {
		t.Fatal("Job model DB was not wired correctly")
	}

	if models.Reports.DB != db {
		t.Fatal("ConsumerActivityReport model DB was not wired correctly")
	}
}

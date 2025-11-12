package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func SkipIfNotIntegration(t *testing.T) {
	if value, _ := os.LookupEnv("INTEGRATION"); value != "true" {
		t.Skip("skipping integration test: INTEGRATION environment variable not set to 'true'")
	}
}

func OpenDB() (*sql.DB, error) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		"user",
		"pass",
		"localhost:5432",
		"queue",
	)

	return sql.Open("pgx", dbURL)
}

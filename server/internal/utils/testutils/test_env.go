package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func SkipIfNotInTestEnv(t *testing.T) {
	t.Helper()
	if value, _ := os.LookupEnv("TEST_ENV"); value != "true" {
		t.Skip("skipping test: not in test environment")
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

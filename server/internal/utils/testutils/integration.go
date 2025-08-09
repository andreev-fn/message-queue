package testutils

import (
	"database/sql"
	"fmt"
	"os"
)

func ShouldRunIntegrationTests() bool {
	if value, _ := os.LookupEnv("INTEGRATION"); value == "true" {
		return true
	}
	return false
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

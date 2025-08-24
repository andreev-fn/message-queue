package config

import (
	"errors"
	"os"
	"strconv"

	"server/internal/appbuilder"
)

const prefix = "APP_"

func Parse() (*appbuilder.Config, error) {
	dbHost := "127.0.0.1:5432"
	if val, ok := os.LookupEnv(prefix + "DB_HOST"); ok {
		dbHost = val
	}

	dbUser := "user"
	if val, ok := os.LookupEnv(prefix + "DB_USER"); ok {
		dbUser = val
	}

	dbPassword := "pass"
	if val, ok := os.LookupEnv(prefix + "DB_PASS"); ok {
		dbPassword = val
	}

	dbName := "queue"
	if val, ok := os.LookupEnv(prefix + "DB_NAME"); ok {
		dbName = val
	}

	maxBatchSize := 100
	if val, ok := os.LookupEnv(prefix + "BATCH_SIZE_MAX"); ok {
		parsedVal, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		if parsedVal < 1 {
			return nil, errors.New("batch size must be a positive number")
		}
		maxBatchSize = parsedVal
	}

	return &appbuilder.Config{
		DatabaseHost:     dbHost,
		DatabaseUser:     dbUser,
		DatabasePassword: dbPassword,
		DatabaseName:     dbName,
		MaxBatchSize:     maxBatchSize,
	}, nil
}

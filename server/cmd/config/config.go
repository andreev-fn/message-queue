package config

import (
	"os"
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

	return &appbuilder.Config{
		DatabaseHost:     dbHost,
		DatabaseUser:     dbUser,
		DatabasePassword: dbPassword,
		DatabaseName:     dbName,
	}, nil
}

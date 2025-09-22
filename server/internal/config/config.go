package config

import (
	"errors"
	"fmt"

	"server/internal/domain"
	"server/internal/utils/opt"
)

type DBType string

const (
	DBTypePostgres DBType = "postgres"
)

type Config struct {
	databaseType   DBType
	postgresConfig opt.Val[*PostgresConfig]
	batchSizeLimit int
	queues         map[string]*domain.QueueConfig
}

func NewConfig(
	pgConfig opt.Val[*PostgresConfig],
	batchSizeLimit int,
	queues map[string]*domain.QueueConfig,
) (*Config, error) {
	if !pgConfig.IsSet() {
		return nil, fmt.Errorf("postgres config required")
	}

	if batchSizeLimit <= 0 {
		return nil, errors.New("batch size limit must be greater than zero")
	}

	if len(queues) == 0 {
		return nil, errors.New("at least one queue must be defined")
	}

	return &Config{
		databaseType:   DBTypePostgres,
		postgresConfig: pgConfig,
		batchSizeLimit: batchSizeLimit,
		queues:         queues,
	}, nil
}

func (c *Config) DatabaseType() DBType                     { return c.databaseType }
func (c *Config) PostgresConfig() opt.Val[*PostgresConfig] { return c.postgresConfig }
func (c *Config) BatchSizeLimit() int                      { return c.batchSizeLimit }

func (c *Config) GetQueueConfig(name string) (*domain.QueueConfig, bool) {
	conf, ok := c.queues[name]
	return conf, ok
}

type PostgresConfig struct {
	host     string
	dbName   string
	username string
	password string
}

func NewPostgresConfig(
	host string,
	dbName string,
	username string,
	password string,
) (*PostgresConfig, error) {
	if host == "" {
		return nil, errors.New("host must not be empty")
	}
	if dbName == "" {
		return nil, errors.New("dbName must not be empty")
	}
	if username == "" {
		return nil, errors.New("username must not be empty")
	}

	return &PostgresConfig{
		host:     host,
		dbName:   dbName,
		username: username,
		password: password,
	}, nil
}

func (c *PostgresConfig) Host() string     { return c.host }
func (c *PostgresConfig) DBName() string   { return c.dbName }
func (c *PostgresConfig) Username() string { return c.username }
func (c *PostgresConfig) Password() string { return c.password }

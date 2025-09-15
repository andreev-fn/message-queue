package config

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"server/internal/domain"
	"server/internal/utils/opt"
)

const defaultBatchSizeLimit = 100

type DBType string

const (
	DBTypePostgres DBType = "postgres"
)

type Config struct {
	databaseType   DBType
	postgresConfig opt.Val[*PostgresConfig]
	batchSizeLimit int
	queues         map[string]*QueueConfig
}

func NewConfig(
	pgConfig opt.Val[*PostgresConfig],
	batchSizeLimit opt.Val[int],
	queues map[string]*QueueConfig,
) (*Config, error) {
	if !pgConfig.IsSet() {
		return nil, fmt.Errorf("postgres config required")
	}

	dbType := DBTypePostgres

	if !batchSizeLimit.IsSet() {
		batchSizeLimit = opt.Some(defaultBatchSizeLimit)
	}
	if batchSizeLimit.MustValue() <= 0 {
		return nil, errors.New("batch size limit must be greater than zero")
	}

	if len(queues) == 0 {
		return nil, errors.New("at least one queue must be defined")
	}

	return &Config{
		databaseType:   dbType,
		postgresConfig: pgConfig,
		batchSizeLimit: batchSizeLimit.MustValue(),
		queues:         queues,
	}, nil
}

func (c *Config) DatabaseType() DBType                     { return c.databaseType }
func (c *Config) PostgresConfig() opt.Val[*PostgresConfig] { return c.postgresConfig }
func (c *Config) BatchSizeLimit() int                      { return c.batchSizeLimit }
func (c *Config) QueueConfig(name string) *QueueConfig     { return c.queues[name] }

func (c *Config) IsQueueDefined(name string) bool {
	_, ok := c.queues[name]
	return ok
}

type QueueConfig struct {
	backoff           opt.Val[*BackoffConfig]
	processingTimeout time.Duration
	retentionPeriod   opt.Val[time.Duration]
}

func NewQueueConfig(
	backoff opt.Val[*BackoffConfig],
	processingTimeout time.Duration,
	retentionPeriod opt.Val[time.Duration],
) (*QueueConfig, error) {
	if processingTimeout < time.Second {
		return nil, errors.New("processing timeout must be at least 1 second")
	}
	if value, isSet := retentionPeriod.Value(); isSet && value <= 0 {
		return nil, errors.New("retention period must be greater than zero")
	}

	return &QueueConfig{
		backoff:           backoff,
		processingTimeout: processingTimeout,
		retentionPeriod:   retentionPeriod,
	}, nil
}

func (c *QueueConfig) Backoff() opt.Val[*BackoffConfig]        { return c.backoff }
func (c *QueueConfig) ProcessingTimeout() time.Duration        { return c.processingTimeout }
func (c *QueueConfig) RetentionPeriod() opt.Val[time.Duration] { return c.retentionPeriod }

type BackoffConfig struct {
	shape       []time.Duration
	maxAttempts opt.Val[int]
}

var _ domain.BackoffConfig = (*BackoffConfig)(nil)

func NewBackoffConfig(
	shape []time.Duration,
	maxAttempts opt.Val[int],
) (*BackoffConfig, error) {
	if len(shape) < 1 {
		return nil, errors.New("shape must have at least one element")
	}

	if value, isSet := maxAttempts.Value(); isSet && value <= 0 {
		return nil, errors.New("max attempts must be greater than zero if provided")
	}

	return &BackoffConfig{
		shape:       shape,
		maxAttempts: maxAttempts,
	}, nil
}

func (c *BackoffConfig) Shape() []time.Duration    { return slices.Clone(c.shape) }
func (c *BackoffConfig) MaxAttempts() opt.Val[int] { return c.maxAttempts }

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

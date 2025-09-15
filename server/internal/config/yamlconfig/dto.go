package yamlconfig

import (
	"fmt"
	"io"
	"time"

	"go.yaml.in/yaml/v4"
)

type ConfigDTO struct {
	DB struct {
		PostgresConfig *PostgresConfig `yaml:"postgres"`
	} `yaml:"db"`
	App *struct {
		BatchSizeLimit *int `yaml:"batch_size_limit"`
	} `yaml:"app"`
	Queues map[string]QueueConfig `yaml:"queues"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	DBName   string `yaml:"db_name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type QueueConfig struct {
	Backoff           *BackoffConfig `yaml:"backoff"`
	ProcessingTimeout time.Duration  `yaml:"processing_timeout"`
	RetentionPeriod   *time.Duration `yaml:"retention_period"`
}

type BackoffConfig struct {
	Shape       []time.Duration `yaml:"shape"`
	MaxAttempts *int            `yaml:"max_attempts"` // nil => no maximum
}

func ReadConfigDTO(r io.Reader) (*ConfigDTO, error) {
	var root yaml.Node
	if err := yaml.NewDecoder(r).Decode(&root); err != nil {
		return nil, fmt.Errorf("dec.Decode: %w", err)
	}

	if err := walkNodeWithHil(&root, newHilEvalConfig()); err != nil {
		return nil, err
	}

	var dto ConfigDTO
	if err := root.Decode(&dto); err != nil {
		return nil, fmt.Errorf("root.Decode: %w", err)
	}

	return &dto, nil
}

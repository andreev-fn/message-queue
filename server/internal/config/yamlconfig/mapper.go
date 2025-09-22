package yamlconfig

import (
	"fmt"

	"server/internal/config"
	"server/internal/domain"
	"server/internal/utils/opt"
)

// MapToAppConfig converts a raw configuration DTO into the immutable application
// configuration object. It applies default values for omitted fields and delegates
// validation and invariant checks to the app/domain constructors.
func MapToAppConfig(dto *ConfigDTO) (*config.Config, error) {
	var postgresConfig opt.Val[*config.PostgresConfig]
	if dto.DB.PostgresConfig != nil {
		tmp, err := config.NewPostgresConfig(
			dto.DB.PostgresConfig.Host,
			dto.DB.PostgresConfig.DBName,
			dto.DB.PostgresConfig.Username,
			dto.DB.PostgresConfig.Password,
		)
		if err != nil {
			return nil, fmt.Errorf("config.NewPostgresConfig: %w", err)
		}

		postgresConfig = opt.Some(tmp)
	}

	queues := make(map[string]*domain.QueueConfig, len(dto.Queues))
	for qName, qConf := range dto.Queues {
		backoffConfig, err := mapBackoffConfig(qConf.Backoff)
		if err != nil {
			return nil, fmt.Errorf("queue %s: %w", qName, err)
		}

		queues[qName], err = domain.NewQueueConfig(
			backoffConfig,
			qConf.ProcessingTimeout,
		)
		if err != nil {
			return nil, fmt.Errorf("queue %s: domain.NewQueueConfig: %w", qName, err)
		}
	}

	batchSizeLimit := config.DefaultBatchSizeLimit
	if dto.App != nil {
		if dto.App.BatchSizeLimit != nil {
			batchSizeLimit = *dto.App.BatchSizeLimit
		}
	}

	return config.NewConfig(
		postgresConfig,
		batchSizeLimit,
		queues,
	)
}

func mapBackoffConfig(dto *BackoffConfig) (opt.Val[*domain.BackoffConfig], error) {
	none := opt.None[*domain.BackoffConfig]()

	if dto == nil {
		return opt.Some(config.DefaultBackoffConfig()), nil
	}

	if derefOrDefault(dto.Enabled, config.DefaultBackoffEnabled) == false {
		return none, nil
	}

	shape := config.DefaultBackoffShape()
	if dto.Shape != nil {
		shape = dto.Shape
	}

	conf, err := domain.NewBackoffConfig(
		shape,
		mapMaxAttempts(dto.MaxAttempts),
	)
	if err != nil {
		return none, fmt.Errorf("domain.NewBackoffConfig: %w", err)
	}

	return opt.Some(conf), nil
}

func mapMaxAttempts(value *OptionalLimit) opt.Val[int] {
	if value == nil {
		return opt.Some(config.DefaultBackoffMaxAttempts)
	}

	maxAttempts, isSet := value.Value()
	if !isSet {
		return opt.None[int]()
	}
	return opt.Some(maxAttempts)
}

func derefOrDefault[T any](ref *T, def T) T {
	if ref != nil {
		return *ref
	}
	return def
}

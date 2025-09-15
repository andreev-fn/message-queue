package yamlconfig

import (
	"fmt"
	"os"

	"server/internal/config"
	"server/internal/utils/opt"
)

func Load(configPath string) (*config.Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	defer f.Close()

	dto, err := ReadConfigDTO(f)
	if err != nil {
		return nil, fmt.Errorf("ReadConfigDTO: %w", err)
	}

	var postgresConfig opt.Val[*config.PostgresConfig]
	if dto.DB.PostgresConfig != nil {
		tmp, err := config.NewPostgresConfig(
			dto.DB.PostgresConfig.Host,
			dto.DB.PostgresConfig.DBName,
			dto.DB.PostgresConfig.Username,
			dto.DB.PostgresConfig.Password,
		)
		if err != nil {
			return nil, fmt.Errorf("NewPostgresConfig: %w", err)
		}

		postgresConfig = opt.Some(tmp)
	}

	queues := make(map[string]*config.QueueConfig, len(dto.Queues))
	for qName, qConf := range dto.Queues {
		backoffConfig := opt.None[*config.BackoffConfig]()
		if qConf.Backoff != nil {
			tmp, err := config.NewBackoffConfig(
				qConf.Backoff.Shape,
				opt.FromRef(qConf.Backoff.MaxAttempts),
			)
			if err != nil {
				return nil, fmt.Errorf("queue %s: NewBackoffConfig: %w", qName, err)
			}

			backoffConfig = opt.Some(tmp)
		}

		queues[qName], err = config.NewQueueConfig(
			backoffConfig,
			qConf.ProcessingTimeout,
			opt.FromRef(qConf.RetentionPeriod),
		)
		if err != nil {
			return nil, fmt.Errorf("queue %s: NewQueueConfig: %w", qName, err)
		}
	}

	var batchSizeLimit opt.Val[int]
	if dto.App != nil {
		batchSizeLimit = opt.FromRef(dto.App.BatchSizeLimit)
	}

	return config.NewConfig(
		postgresConfig,
		batchSizeLimit,
		queues,
	)
}

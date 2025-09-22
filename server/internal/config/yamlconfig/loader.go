package yamlconfig

import (
	"fmt"
	"os"

	"server/internal/config"
)

func LoadFromFile(configPath string) (*config.Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	defer f.Close()

	dto, err := NewFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("NewFromReader: %w", err)
	}

	return MapToAppConfig(dto)
}

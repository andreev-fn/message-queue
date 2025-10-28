package openapi

import (
	"bytes"
	_ "embed"
	"encoding/json"

	"go.yaml.in/yaml/v4"
)

//go:embed schema.yaml
var specYAML []byte

func GetSpecYAML() []byte {
	return bytes.Clone(specYAML)
}

func GetSpecJSON() ([]byte, error) {
	var data any
	if err := yaml.Unmarshal(specYAML, &data); err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}

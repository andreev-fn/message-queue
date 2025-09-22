package yamlconfig

import (
	"errors"
	"strconv"

	"go.yaml.in/yaml/v4"
)

type OptionalLimit struct {
	isSet bool
	value int
}

func (l *OptionalLimit) Value() (int, bool) {
	return l.value, l.isSet
}

func (l *OptionalLimit) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return errors.New("expected scalar node")
	}

	if node.Value == "unlimited" {
		l.isSet = false
		return nil
	}

	if parsedValue, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
		l.isSet = true
		l.value = int(parsedValue)
		return nil
	}

	return errors.New("invalid value, expected int or 'unlimited'")
}

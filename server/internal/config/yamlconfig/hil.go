package yamlconfig

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"go.yaml.in/yaml/v4"
)

func newHilEvalConfig() *hil.EvalConfig {
	envFunc := ast.Function{
		Variadic:     true,
		ArgTypes:     []ast.Type{ast.TypeString},
		VariadicType: ast.TypeString,
		ReturnType:   ast.TypeString,
		Callback: func(inputs []any) (any, error) {
			if len(inputs) > 2 {
				return "", errors.New("too many arguments")
			}
			envName := inputs[0].(string)
			if val, ok := os.LookupEnv(envName); ok {
				return val, nil
			}
			if len(inputs) == 2 {
				return inputs[1].(string), nil
			}
			return "", fmt.Errorf("variable not set: %s", envName)
		},
	}

	return &hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			FuncMap: map[string]ast.Function{
				"env": envFunc,
			},
		},
	}
}

func walkNodeWithHil(node *yaml.Node, evalConf *hil.EvalConfig) error {
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		// Root or array -> walk children
		for _, c := range node.Content {
			if err := walkNodeWithHil(c, evalConf); err != nil {
				return err
			}
		}

	case yaml.MappingNode:
		// key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			if err := walkNodeWithHil(keyNode, evalConf); err != nil {
				return err
			}
			if err := walkNodeWithHil(valNode, evalConf); err != nil {
				return err
			}
		}

	case yaml.ScalarNode:
		if node.Tag == "!!str" {
			tree, err := hil.Parse(node.Value)
			if err != nil {
				return fmt.Errorf("HIL error at line %d, col %d: %w", node.Line, node.Column, err)
			}

			result, err := hil.Eval(tree, evalConf)
			if err != nil {
				return fmt.Errorf("HIL error at line %d, col %d: %w", node.Line, node.Column, err)
			}

			node.Value = fmt.Sprint(result.Value)
			if _, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
				node.Tag = "!!int"
			}
			if _, err := strconv.ParseFloat(node.Value, 64); err == nil {
				node.Tag = "!!float"
			}
		}
	case yaml.AliasNode:
	}
	return nil
}
